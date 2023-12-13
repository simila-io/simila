package fts

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"strings"
)

const (
	createTsConfigUp = `
create text search configuration public.simila (copy = pg_catalog.english);

alter text search configuration simila
	alter mapping for asciiword, asciihword, hword_asciipart, word, hword, hword_part
	with spanish_stem, english_stem;
`
	createTsConfigDown = `
do $$
begin
    if exists(
		select "cfgname" from "pg_ts_config" where "cfgname" = 'simila'
	)
    then
		drop text search configuration public.simila cascade;
    end if;
end
$$;
`
	createSegmentTsVectorUp = ` 
alter table "index_record"
    add column if not exists "segment_tsvector" tsvector generated always as (to_tsvector('public.simila', "segment")) stored;

create index if not exists "idx_index_record_segment_tsvector" on "index_record" using gin ("segment_tsvector");
`
	createSegmentTsVectorDown = ` 
drop index if exists "idx_index_record_segment_tsvector";

alter table "index_record" drop column if exists "segment_tsvector";
`
)

// TODO: we can extend public.simila TS configuration with other dictionaries,
// this will produce more lexems and help with search, but it will increase the
// size of the "segment_tsvector", probably this should be made configurable on
// customer basis.

func createTsConfig(id string, rollback bool) *migrate.Migration {
	m := &migrate.Migration{
		Id:                     id,
		Down:                   []string{createTsConfigDown},
		DisableTransactionDown: true,
	}
	if !rollback {
		m.Up = []string{createTsConfigUp}
		m.DisableTransactionUp = true
	}
	return m
}

func createSegmentTsVector(id string, rollback bool) *migrate.Migration {
	m := &migrate.Migration{
		Id:   id,
		Down: []string{createSegmentTsVectorDown},
	}
	if !rollback {
		m.Up = []string{createSegmentTsVectorUp}
	}
	return m
}

// Migrations returns migrations to be applied on top of
// the "common" migrations for the postgres built-in  full-text search
// module to work, the module migration IDs range is [3000-3999]
func Migrations(rollback bool) []*migrate.Migration {
	return []*migrate.Migration{
		createTsConfig("3000", rollback),
		createSegmentTsVector("3001", rollback),
	}
}

// Search is an implementation of the postgres.SearchFn
// function based on the postgres built-in full-text search.
// Queries must be formed in accordance with the `websearch_to_tsquery()` query syntax,
// see https://www.postgresql.org/docs/current/textsearch-controls.html#TEXTSEARCH-PARSING-QUERIES.
func Search(ctx context.Context, qx sqlx.QueryerContext, r persistence.Node, q persistence.SearchQuery) (persistence.SearchQueryResult, error) {
	var params []any
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(" segment_tsvector @@ websearch_to_tsquery('simila', $%d) ", len(params)+1))
	params = append(params, q.Query)

	if len(q.Format) > 0 {
		sb.WriteString(fmt.Sprintf(" and format = $%d ", len(params)+1))
		params = append(params, q.Format)
	}

	var count string
	var query string

	qrPrm := len(params)
	kwFmt := "MaxFragments=10, MaxWords=7, MinWords=1, StartSel=<<, StopSel=>>"

	if q.Strict {
		sb.WriteString(fmt.Sprintf(" and node_id = $%d and n.tags @> $%d ", len(params)+1, len(params)+2))
		params = append(params, r.ID, q.Tags.JSON())

		where := sb.String()
		count = fmt.Sprintf("select count(*) from index_record "+
			"inner join node as n on n.id = node_id "+
			"where %s", where)

		query = fmt.Sprintf("select index_record.*, "+
			"concat(n.path, n.name) as path, "+
			"(ts_rank_cd(segment_tsvector, websearch_to_tsquery('simila', $%d))*rank_multiplier) as score, "+
			"ts_headline(segment, websearch_to_tsquery('simila', $%d), '%s') as matched_keywords "+
			"from index_record "+
			"inner join node as n on n.id = node_id "+
			"where %s "+
			"order by score desc, id "+
			"offset $%d limit $%d", qrPrm, qrPrm, kwFmt, where, len(params)+1, len(params)+2)

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
		count = fmt.Sprintf("select count(distinct node_id) from index_record "+
			"inner join node as n on n.id = node_id "+
			"where %s", where)

		query = fmt.Sprintf("select distinct on(score, path) index_record.*, "+
			"concat(n.path, n.name) as path, "+
			"r.score as score, "+
			"ts_headline(segment, websearch_to_tsquery('simila', $%d), '%s') as matched_keywords "+
			"from ("+
			//
			"select node_id, "+
			"max(ts_rank_cd(segment_tsvector, websearch_to_tsquery('simila', $%d))*rank_multiplier) as score "+
			"from index_record "+
			"where %s "+
			"group by node_id"+
			//
			") as r "+
			"inner join index_record on index_record.node_id = r.node_id and "+
			"(ts_rank_cd(segment_tsvector, websearch_to_tsquery('simila', $%d))*rank_multiplier) = r.score "+
			"inner join node as n on n.id = r.node_id "+
			"order by score desc, path, id "+
			"offset $%d limit $%d", qrPrm, kwFmt, qrPrm, where, qrPrm, len(params)+1, len(params)+2)
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
		persistence.MapKeywordsToListFn("<<", ">>"))
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}
	return persistence.SearchQueryResult{Items: res, Total: total}, nil
}
