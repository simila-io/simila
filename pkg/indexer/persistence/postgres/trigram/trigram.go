package trigram

import (
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/strutil"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/simila-io/simila/pkg/ql"
	"strings"
)

const (
	createExtensionUp = `
create extension if not exists pg_trgm;
`
	createSegmentIndexUp = `
create index if not exists "idx_index_record_segment_trgm" on "index_record" using gin ("segment" gin_trgm_ops);
`
	createSegmentIndexDown = `
drop index if exists "idx_index_record_segment_trgm";
`
)

func createExtension(id string, rollback bool) *migrate.Migration {
	m := &migrate.Migration{
		Id: id,
	}
	if !rollback {
		m.Up = []string{createExtensionUp}
		m.DisableTransactionUp = true
	}
	return m
}

func createSegmentIndex(id string, rollback bool) *migrate.Migration {
	m := &migrate.Migration{
		Id:   id,
		Down: []string{createSegmentIndexDown},
	}
	if !rollback {
		m.Up = []string{createSegmentIndexUp}
	}
	return m
}

// FcTranslator is the filter conditions translator from simila QL to the Postgres dialect
var FcTranslator = ql.NewTranslator(ql.PqFilterConditionsDialect)

// Migrations returns migrations to be applied on top of
// the "common" migrations for the "trigram" search module to work,
// the "trigram" module migration IDs range is [2000-2999]
func Migrations(rollback bool) []*migrate.Migration {
	return []*migrate.Migration{
		createExtension("2000", rollback),
		createSegmentIndex("2001", rollback),
	}
}

// SessionParams returns a map of k:v pairs, which represent DB settings
// to be applied as soon as the DB session is started. This may be needed
// to tweak parameters of the DB extension, since some controlled DB envs
// (e.g. RDS) do not always allow to modify system or DB settings on a
// permanent basis (i.e. ALTER SYSTEM/DATABASE SET..).
func SessionParams() map[string]any {
	return map[string]any{"pg_trgm.word_similarity_threshold": 0.3}
}

// Search is an implementation of the postgres.SearchFn
// function based on the "pg_trgm" postgres extension.
// SearchQuery.TextQuery is text, no expressions are supported for now, the whole
// segment of text is matched against the whole query text using `trigram word similarity`,
// see https://www.postgresql.org/docs/current/pgtrgm.html.
func Search(ctx context.Context, qx sqlx.QueryerContext, q persistence.SearchQuery) (persistence.SearchQueryResult, error) {
	var sb strings.Builder
	sb.Grow(2 * len(q.FilterConditions))
	if err := FcTranslator.Translate(&sb, q.FilterConditions); err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}
	if sb.Len() > 0 {
		sb.WriteString(" and ")
	}

	var params []any
	sb.WriteString(fmt.Sprintf(" segment %%> $%d ", len(params)+1))
	params = append(params, q.TextQuery)

	qrPrm := 1
	where := sb.String()

	var count string
	var query string

	if q.GroupByPathOff {
		count = fmt.Sprintf(`select count(*)
			from (
				select ir.id from index_record as ir
				inner join node as n on n.id = ir.node_id
				where %s
			) as r`, where)

		query = fmt.Sprintf(`select ir.*,
			n.name as path,
			((1 - (ir.segment <->> $%d))*ir.rank_multiplier) as score
			from index_record as ir
			inner join node as n on n.id = ir.node_id
			where %s
			order by score desc, ir.id
			offset $%d limit $%d`, qrPrm, where, len(params)+1, len(params)+2)

	} else {
		count = fmt.Sprintf(`select count(*)
			from (
				select ir.node_id from index_record as ir
				inner join node as n on n.id = ir.node_id
				where %s 
				group by ir.node_id
			) as r`, where)

		query = fmt.Sprintf(`select distinct on(score, path) index_record.*,
			r.fullpath as path,
			r.score as score
			from (
				select ir.node_id,
				n.name as fullpath,
				max((1 - (ir.segment <->> $%d))*ir.rank_multiplier) as score
				from index_record as ir
				inner join node as n on n.id = ir.node_id
				where %s
				group by ir.node_id, n.name
			) as r
			inner join index_record on index_record.node_id = r.node_id and
			((1 - (segment <->> $%d))*rank_multiplier) = r.score
			order by score desc, path, id
			offset $%d limit $%d`, qrPrm, where, qrPrm, len(params)+1, len(params)+2)
	}

	// count
	total, err := persistence.Count(ctx, qx, count, params...)
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}

	// query
	if q.Limit <= 0 {
		return persistence.SearchQueryResult{Total: total}, nil
	}
	params = append(params, q.Offset, q.Limit)
	rows, err := qx.QueryxContext(ctx, query, params...)
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}
	defer func() {
		_ = rows.Close()
	}()
	// results
	res, err := persistence.ScanRowsQueryResultAndMap(rows, mapKeywordsToListFn(q.TextQuery))
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}
	return persistence.SearchQueryResult{Items: res, Total: total}, nil
}

func mapKeywordsToListFn(query string) func(item persistence.SearchQueryResultItem) persistence.SearchQueryResultItem {
	trimSet := "!@#$%^&*(){}[]|;:\"'`<>.,?"
	wordMap := make(map[string]struct{})
	for _, w := range strings.Fields(query) {
		wordMap[strings.ToLower(strings.Trim(w, trimSet))] = struct{}{}
	}
	return func(item persistence.SearchQueryResultItem) persistence.SearchQueryResultItem {
		item.MatchedKeywordsList = make([]string, 0)
		for _, w := range strings.Fields(item.Segment) {
			w = strings.Trim(w, trimSet)
			if _, ok := wordMap[strings.ToLower(w)]; ok {
				item.MatchedKeywordsList = append(item.MatchedKeywordsList, w)
			}
		}
		if len(item.MatchedKeywordsList) > 0 {
			item.MatchedKeywordsList = strutil.RemoveDups(item.MatchedKeywordsList)
		}
		return item
	}
}
