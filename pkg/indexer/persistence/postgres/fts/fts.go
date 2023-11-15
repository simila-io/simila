package fts

import (
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
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

func createTsConfig(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:                     id,
		Up:                     []string{createTsConfigUp},
		Down:                   []string{createTsConfigDown},
		DisableTransactionUp:   true,
		DisableTransactionDown: true,
	}
}

func createSegmentTsVector(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:   id,
		Up:   []string{createSegmentTsVectorUp},
		Down: []string{createSegmentTsVectorDown},
	}
}

// Migrations returns migrations to be applied on top of
// the "common" migrations for the postgres built-in  full-text search
// module to work, the module migration IDs range is [3000-3999]
func Migrations() []*migrate.Migration {
	return []*migrate.Migration{
		createTsConfig("3000"),
		createSegmentTsVector("3001"),
	}
}

// Search is an implementation of the postgres.SearchFn
// function based on the postgres built-in full-text search.
// Queries must be formed in accordance with the `websearch_to_tsquery()` query syntax,
// see https://www.postgresql.org/docs/current/textsearch-controls.html#TEXTSEARCH-PARSING-QUERIES.
func Search(ctx context.Context, q sqlx.QueryerContext, query persistence.SearchQuery) (persistence.QueryResult[persistence.SearchQueryResultItem, string], error) {
	if len(query.Query) == 0 {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, fmt.Errorf("search query must be non-empty: %w", errors.ErrInvalid)
	}
	sb := strings.Builder{}
	args := make([]any, 0)

	if len(query.FromID) > 0 {
		var fromID persistence.IndexRecordID
		if err := fromID.Decode(query.FromID); err != nil {
			return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, fmt.Errorf("invalid FromID: %w", errors.ErrInvalid)
		}
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" index_record.index_id >= ? and index_record.id >= ? ")
		args = append(args, fromID.IndexID, fromID.RecordID)
	}
	if len(query.IndexIDs) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		oldLen := len(args)
		sb.WriteString(" index_record.index_id in ( ")
		for _, id := range query.IndexIDs {
			if len(args) > oldLen {
				sb.WriteString(", ")
			}
			sb.WriteString("?")
			args = append(args, id)
		}
		sb.WriteString(")")
	}
	if len(query.Tags) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		var tb strings.Builder
		tb.WriteString(" {")
		oldLen := tb.Len()
		for k, v := range query.Tags {
			if tb.Len() > oldLen {
				tb.WriteByte(',')
			}
			tb.WriteString(fmt.Sprintf("%q:%q", k, v))
		}
		tb.WriteString("}")
		sb.WriteString(" index.tags @> ?")
		args = append(args, tb.String())
	}
	if len(query.Query) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" index_record.segment_tsvector @@ websearch_to_tsquery('simila', ?) ")
		args = append(args, query.Query)
	}

	where := sqlx.Rebind(sqlx.DOLLAR, sb.String())
	if len(where) > 0 {
		where = " where " + where
	}

	distinct := ""
	if query.Distinct {
		if query.OrderByScore {
			distinct = "distinct on(score, index_record.index_id)"
		} else {
			distinct = "distinct on(index_record.index_id)"
		}
	}

	orderBy, limit := "", 0
	if query.OrderByScore {
		orderBy = "order by score desc, index_record.index_id asc, index_record.id asc"
		limit = query.Limit // no +1, since no pagination
	} else {
		orderBy = "order by index_record.index_id asc, index_record.id asc"
		limit = query.Limit + 1
	}

	// count
	args = append(args, query.Query)
	total, err := persistence.Count(ctx, q, fmt.Sprintf("select count(*) "+
		"from (select %s index_record.*, ts_rank_cd(index_record.segment_tsvector, websearch_to_tsquery('simila', $%d)) as score "+
		"from index_record "+
		"inner join index on index.id = index_record.index_id %s %s) as r", distinct, len(args), where, orderBy), args...)
	if err != nil {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, persistence.MapError(err)
	}

	// query
	if query.Limit <= 0 {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{Total: total}, nil
	}
	args = append(args, query.Offset, limit)
	rows, err := q.QueryxContext(ctx, fmt.Sprintf("select %s index_record.*, "+
		"ts_rank_cd(index_record.segment_tsvector, websearch_to_tsquery('simila', $%d)) as score, "+
		"ts_headline(index_record.segment, websearch_to_tsquery('simila', $%d), 'MaxFragments=10, MaxWords=7, MinWords=1, StartSel=<<, StopSel=>>') as matched_keywords "+
		"from index_record "+
		"inner join index on index.id = index_record.index_id %s %s offset $%d limit $%d",
		distinct, len(args)-2, len(args)-2, where, orderBy, len(args)-1, len(args)), args...)
	if err != nil {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, persistence.MapError(err)
	}

	// results
	fRes, err := persistence.ScanRowsQueryResult[ftsSearchQueryResultItem](rows)
	if err != nil {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, persistence.MapError(err)
	}
	res := toSearchQueryResultItem(fRes)
	var nextID persistence.IndexRecordID
	if len(res) > query.Limit {
		nextID = persistence.IndexRecordID{IndexID: res[len(res)-1].IndexID, RecordID: res[len(res)-1].ID}
		res = res[:query.Limit]
	}
	return persistence.QueryResult[persistence.SearchQueryResultItem, string]{Items: res, NextID: nextID.Encode(), Total: total}, nil
}

// includes SegmentTsVector that is specific
// to FTS search only and is not needed by other modules
type ftsSearchQueryResultItem struct {
	persistence.SearchQueryResultItem
	SegmentTsVector string `db:"segment_tsvector"`
}

func toSearchQueryResultItem(fRes []ftsSearchQueryResultItem) []persistence.SearchQueryResultItem {
	res := make([]persistence.SearchQueryResultItem, len(fRes))
	mapFn := persistence.MapKeywordsToListFn("<<", ">>")
	for i, fr := range fRes {
		res[i] = mapFn(fr.SearchQueryResultItem)
	}
	return res
}
