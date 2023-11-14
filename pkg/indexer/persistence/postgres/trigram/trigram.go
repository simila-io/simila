package trigram

import (
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
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

func createExtension(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:                   id,
		Up:                   []string{createExtensionUp},
		DisableTransactionUp: true,
	}
}

func createSegmentIndex(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:   id,
		Up:   []string{createSegmentIndexUp},
		Down: []string{createSegmentIndexDown},
	}
}

// Migrations returns migrations to be applied on top of
// the "common" migrations for the "trigram" search module to work,
// the "trigram" module migration IDs range is [2000-2999]
func Migrations() []*migrate.Migration {
	return []*migrate.Migration{
		createExtension("2000"),
		createSegmentIndex("2001"),
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
// Queries are just text, no conditional expressions are supported for now,
// the whole segment of text is matched against the whole query text using `word similarity`,
// see https://www.postgresql.org/docs/current/pgtrgm.html.
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
		sb.WriteString(" index_record.segment %> ? ")
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
		"from (select %s index_record.*, 1.0 - (index_record.segment <->> $%d) as score "+
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
		"1 - (index_record.segment <->> $%d) as score "+
		"from index_record "+
		"inner join index on index.id = index_record.index_id %s %s offset $%d limit $%d",
		distinct, len(args)-2, where, orderBy, len(args)-1, len(args)), args...)
	if err != nil {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, persistence.MapError(err)
	}

	// results
	res, err := persistence.ScanRowsQueryResultAndMap(rows, mapKeywordsToListFn(query.Query))
	if err != nil {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, persistence.MapError(err)
	}
	var nextID persistence.IndexRecordID
	if len(res) > query.Limit {
		nextID = persistence.IndexRecordID{IndexID: res[len(res)-1].IndexID, RecordID: res[len(res)-1].ID}
		res = res[:query.Limit]
	}
	return persistence.QueryResult[persistence.SearchQueryResultItem, string]{Items: res, NextID: nextID.Encode(), Total: total}, nil
}

func mapKeywordsToListFn(query string) func(item persistence.SearchQueryResultItem) persistence.SearchQueryResultItem {
	trimSet := "!@#$%^&*(){}[]|:\".,?"
	wordMap := make(map[string]struct{})
	for _, w := range strings.Fields(query) {
		wordMap[strings.Trim(strings.ToLower(w), trimSet)] = struct{}{}
	}
	return func(item persistence.SearchQueryResultItem) persistence.SearchQueryResultItem {
		for _, w := range strings.Fields(item.Segment) {
			if _, ok := wordMap[strings.Trim(strings.ToLower(w), trimSet)]; ok {
				item.MatchedKeywordsList = append(item.MatchedKeywordsList, w)
			}
		}
		item.MatchedKeywordsList = strutil.RemoveDups(item.MatchedKeywordsList)
		return item
	}
}
