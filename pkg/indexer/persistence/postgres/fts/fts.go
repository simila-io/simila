package fts

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

// FcTranslator is the filter conditions translator from simila QL to the Postgres dialect
var FcTranslator = ql.NewTranslator(ql.PqFilterConditionsDialect)

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
// SearchQuery.TextQuery must be formed in accordance with the `websearch_to_tsquery()` query syntax,
// see https://www.postgresql.org/docs/current/textsearch-controls.html#TEXTSEARCH-PARSING-QUERIES.
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
	sb.WriteString(fmt.Sprintf(" segment_tsvector @@ websearch_to_tsquery('simila', $%d) ", len(params)+1))
	params = append(params, q.TextQuery)

	qrPrm := 1
	kwFmt := "MaxFragments=10, MaxWords=7, MinWords=1, StartSel=<<, StopSel=>>"
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
			(ts_rank_cd(ir.segment_tsvector, websearch_to_tsquery('simila', $%d))*ir.rank_multiplier) as score,
			ts_headline('simila', ir.segment, websearch_to_tsquery('simila', $%d), '%s') as matched_keywords
			from index_record as ir
			inner join node as n on n.id = ir.node_id
			where %s
			order by score desc, ir.id
			offset $%d limit $%d`, qrPrm, qrPrm, kwFmt, where, len(params)+1, len(params)+2)

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
			ts_headline('simila', segment, websearch_to_tsquery('simila', $%d), '%s') as matched_keywords
			from (
				select ir.node_id,
				n.name as fullpath,
				max(ts_rank_cd(ir.segment_tsvector, websearch_to_tsquery('simila', $%d))*ir.rank_multiplier) as score
				from index_record as ir
				inner join node as n on n.id = ir.node_id
				where %s
				group by ir.node_id, n.name
			) as r
			inner join index_record on index_record.node_id = r.node_id and
			(ts_rank_cd(segment_tsvector, websearch_to_tsquery('simila', $%d))*rank_multiplier) = r.score
			order by score desc, path, id
			offset $%d limit $%d`, qrPrm, kwFmt, qrPrm, where, qrPrm, len(params)+1, len(params)+2)
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
		persistence.MapKeywordsToListFn("<<", ">>"))
	if err != nil {
		return persistence.SearchQueryResult{}, persistence.MapError(err)
	}
	return persistence.SearchQueryResult{Items: res, Total: total}, nil
}
