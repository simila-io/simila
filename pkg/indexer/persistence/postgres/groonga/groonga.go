package groonga

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/simila-io/simila/pkg/indexer/persistence"
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
// Queries must be formed in accordance with the "pgroonga" query syntax,
// see https://pgroonga.github.io/reference/operators/query-v2.html.
func Search(ctx context.Context, qx sqlx.QueryerContext, r persistence.Node, q persistence.SearchQuery) (persistence.SearchQueryResult, error) {
	var params []any
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(" segment &@~ $%d ", len(params)+1))
	params = append(params, q.Query)

	if len(q.Format) > 0 {
		sb.WriteString(fmt.Sprintf(" and format = $%d ", len(params)+1))
		params = append(params, q.Format)
	}

	var count string
	var query string

	qrPrm := len(params)
	if q.Strict {
		sb.WriteString(fmt.Sprintf(" and node_id = $%d and n.tags @> $%d ", len(params)+1, len(params)+2))
		params = append(params, r.ID, q.Tags.JSON())

		where := sb.String()
		count = fmt.Sprintf("select count(*) from index_record "+
			"inner join node as n on n.id = node_id "+
			"where %s", where)

		query = fmt.Sprintf("select index_record.*, "+
			"concat(n.path, n.name) as path, "+
			"(pgroonga_score(index_record.tableoid, index_record.ctid)*rank_multiplier) as score, "+
			"pgroonga_highlight_html(segment, pgroonga_query_extract_keywords($%d)) as matched_keywords "+
			"from index_record "+
			"inner join node as n on n.id = node_id "+
			"where %s "+
			"order by score desc, id "+
			"offset $%d limit $%d", qrPrm, where, len(params)+1, len(params)+2)

	} else {
		if r.Flags == persistence.NodeFlagDocument {
			sb.WriteString(fmt.Sprintf(" and node_id = $%d and n.tags @> $%d ", len(params)+1, len(params)+2))
			params = append(params, r.ID, q.Tags.JSON())
		} else {
			sb.WriteString(fmt.Sprintf(" and node_id in (select id from node "+
				"where path like concat($%d::text, '%%') and tags @> $%d) ", len(params)+1, len(params)+2))
			params = append(params, persistence.ToNodePath(q.Path), q.Tags.JSON())
		}

		where := sb.String()
		count = fmt.Sprintf("select count(distinct node_id) from index_record where %s", where)

		query = fmt.Sprintf("select distinct on(score, path) index_record.*, "+
			"concat(n.path, n.name) as path, "+
			"r.score as score, "+
			"pgroonga_highlight_html(segment, pgroonga_query_extract_keywords($%d)) as matched_keywords "+
			"from ("+
			//
			"select node_id, "+
			"max(pgroonga_score(index_record.tableoid, index_record.ctid)*rank_multiplier) as score "+
			"from index_record "+
			"where %s "+
			"group by node_id"+
			//
			") as r "+
			"inner join index_record on index_record.node_id = r.node_id and "+
			"(pgroonga_score(index_record.tableoid, index_record.ctid)*rank_multiplier) = r.score "+
			"inner join node as n on n.id = r.node_id "+
			"order by score desc, path, id "+
			"offset $%d limit $%d", qrPrm, where, len(params)+1, len(params)+2)
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

	// results
	res, err := persistence.ScanRowsQueryResultAndMap(rows,
		persistence.MapKeywordsToListFn("<span class=\"keyword\">", "</span>"))
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}
	return persistence.SearchQueryResult{Items: res, Total: total}, nil
}
