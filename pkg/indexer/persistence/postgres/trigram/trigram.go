package trigram

import (
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/strutil"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/simila-io/simila/pkg/indexer/persistence"
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
// Queries are just text, no expressions are supported for now, the whole
// segment of text is matched against the whole query text using `trigram word similarity`,
// see https://www.postgresql.org/docs/current/pgtrgm.html.
func Search(ctx context.Context, qx sqlx.QueryerContext, n persistence.Node, q persistence.SearchQuery) (persistence.SearchQueryResult, error) {
	sb := strings.Builder{}
	args := make([]any, 0)

	sb.WriteString(" segment %> ? ")
	args = append(args, q.Query)

	if len(q.Tags) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		var tb strings.Builder
		tb.WriteString(" {")
		oldLen := tb.Len()
		for k, v := range q.Tags {
			if tb.Len() > oldLen {
				tb.WriteByte(',')
			}
			tb.WriteString(fmt.Sprintf("%q:%q", k, v))
		}
		tb.WriteString("}")
		sb.WriteString(" n.tags @> ?")
		args = append(args, tb.String())
	}
	if q.Strict || n.Flags == persistence.NodeFlagDocument {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" node_id = ? ")
		args = append(args, n.ID)
	} else {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" n.path like concat(?::text, '%%') ")
		args = append(args, fmt.Sprintf("%s%s/", n.Path, n.Name))
	}

	var where string
	if sb.Len() > 0 {
		where = " where " + sqlx.Rebind(sqlx.DOLLAR, sb.String())
	}

	var count string
	var query string

	qryArg, offArg, limArg := 1, len(args)+1, len(args)+2

	if q.Strict {
		count = fmt.Sprintf("select count(*) from index_record "+
			"inner join node as n on n.id = node_id %s", where)

		query = fmt.Sprintf("select index_record.*, "+
			"concat(n.path, n.name) as path, "+
			"((1 - (segment <->> $%d))*rank_multiplier) as score "+
			"from index_record "+
			"inner join node as n on n.id = node_id "+
			"%s "+ // where
			"order by score desc, id "+
			"offset $%d limit $%d", qryArg, where, offArg, limArg)

	} else {
		count = fmt.Sprintf("select count(*) from index_record "+
			"inner join node as n on n.id = node_id %s group by node_id", where)

		query = fmt.Sprintf("select distinct on(score, path) index_record.*, r.fullpath as path, "+
			"r.score as score "+
			"from ("+
			//
			"select node_id, "+
			"concat(n.path, n.name) as fullpath, "+
			"max((1 - (segment <->> $%d))*rank_multiplier) as score "+
			"from index_record "+
			"inner join node as n on n.id = node_id "+
			"%s "+ // where
			"group by node_id, fullpath"+
			//
			") as r "+
			"inner join index_record on index_record.node_id = r.node_id and "+
			"((1 - (segment <->> $%d))*rank_multiplier) = r.score "+
			"order by score desc, path, id "+
			"offset $%d limit $%d", qryArg, where, qryArg, offArg, limArg)
	}

	// count
	total, err := persistence.Count(ctx, qx, count, args...)
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}

	// query
	if q.Limit <= 0 {
		return persistence.SearchQueryResult{Total: total}, nil
	}
	args = append(args, q.Offset, q.Limit)
	rows, err := qx.QueryxContext(ctx, query, args...)
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}

	// results
	res, err := persistence.ScanRowsQueryResultAndMap(rows, mapKeywordsToListFn(q.Query))
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
