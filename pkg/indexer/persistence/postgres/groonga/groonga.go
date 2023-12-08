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
func Search(ctx context.Context, qx sqlx.QueryerContext, n persistence.Node, q persistence.SearchQuery) (persistence.SearchQueryResult, error) {
	sb := strings.Builder{}
	args := make([]any, 0)

	sb.WriteString(" segment &@~ ? ")
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
			"(pgroonga_score(index_record.tableoid, index_record.ctid)*rank_multiplier) as score, "+
			"pgroonga_highlight_html(segment, pgroonga_query_extract_keywords($%d)) as matched_keywords "+
			"from index_record "+
			"inner join node as n on n.id = node_id "+
			"%s "+ // where
			"order by score desc, id "+
			"offset $%d limit $%d", qryArg, where, offArg, limArg)

	} else {
		count = fmt.Sprintf("select count(*) from index_record "+
			"inner join node as n on n.id = node_id %s group by node_id", where)

		query = fmt.Sprintf("select distinct on(score, path) index_record.*, r.fullpath as path, "+
			"r.score as score, "+
			"pgroonga_highlight_html(segment, pgroonga_query_extract_keywords($%d)) as matched_keywords "+
			"from ("+
			//
			"select node_id, "+
			"concat(n.path, n.name) as fullpath, "+
			"max(pgroonga_score(index_record.tableoid, index_record.ctid)*rank_multiplier) as score "+
			"from index_record "+
			"inner join node as n on n.id = node_id "+
			"%s "+ // where
			"group by node_id, fullpath"+
			//
			") as r "+
			"inner join index_record on index_record.node_id = r.node_id and "+
			"(pgroonga_score(index_record.tableoid, index_record.ctid)*rank_multiplier) = r.score "+
			"order by score desc, path, id "+
			"offset $%d limit $%d", qryArg, where, offArg, limArg)
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
	res, err := persistence.ScanRowsQueryResultAndMap(rows,
		persistence.MapKeywordsToListFn("<span class=\"keyword\">", "</span>"))
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}
	return persistence.SearchQueryResult{Items: res, Total: total}, nil
}
