package groonga

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/simila-io/simila/pkg/ql"
	"strings"
)

const (
	createExtensionUp = `
create extension if not exists pgroonga;
`
	createSegmentIndexUp = `
create index if not exists "idx_index_record_segment_groonga" on "index_record" using pgroonga ("segment") with (tokenizer='TokenNgram("unify_alphabet", false, "unify_symbol", false, "unify_digit", false)');
`
	createSegmentIndexDown = `
drop index if exists "idx_index_record_segment_groonga";
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
// the "common" migrations for the "groonga" search module to work,
// the "groonga" module migration IDs range is [1000-1999].
func Migrations(rollback bool) []*migrate.Migration {
	return []*migrate.Migration{
		createExtension("1000", rollback),
		createSegmentIndex("1001", rollback),
	}
}

// Search is an implementation of the postgres.SearchFn
// function based on the "pgroonga" postgres extension.
// SearchQuery.TextQuery must be formed in accordance with the "pgroonga" query syntax,
// see https://pgroonga.github.io/reference/operators/query-v2.html.
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
	sb.WriteString(fmt.Sprintf(" segment &@~ $%d ", len(params)+1))
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
			(pgroonga_score(ir.tableoid, ir.ctid)*ir.rank_multiplier) as score,
			pgroonga_highlight_html(ir.segment, pgroonga_query_extract_keywords($%d)) as matched_keywords
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
			r.score as score,
			pgroonga_highlight_html(segment, pgroonga_query_extract_keywords($%d)) as matched_keywords
			from (
				select ir.node_id,
				n.name as fullpath,
				max(pgroonga_score(ir.tableoid, ir.ctid)*ir.rank_multiplier) as score
				from index_record as ir
				inner join node as n on n.id = ir.node_id
				where %s
				group by ir.node_id, n.name
			) as r
			inner join index_record on index_record.node_id = r.node_id and
			(pgroonga_score(index_record.tableoid, index_record.ctid)*rank_multiplier) = r.score
			order by score desc, path, id
			offset $%d limit $%d`, qrPrm, where, len(params)+1, len(params)+2)
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
	res, err := persistence.ScanRowsQueryResultAndMap(rows,
		persistence.MapKeywordsToListFn("<span class=\"keyword\">", "</span>"))
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}
	return persistence.SearchQueryResult{Items: res, Total: total}, nil
}
