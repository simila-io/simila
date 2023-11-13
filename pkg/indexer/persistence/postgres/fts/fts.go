package fts

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/simila-io/simila/pkg/indexer/persistence"
)

// Search is an implementation of the postgres.SearchFn
// function based on the postgres built-in full-text search.
func Search(ctx context.Context, q sqlx.QueryerContext, query persistence.SearchQuery) (persistence.QueryResult[persistence.SearchQueryResultItem, string], error) {
	panic("TODO: implement me!")
}
