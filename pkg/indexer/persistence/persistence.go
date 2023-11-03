// Copyright 2023 The Simila Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package persistence

import (
	"context"
)

type (
	// Tx interface describes an abstract DB transaction.
	Tx interface {
		// MustBegin starts the transaction
		MustBegin()
		// MustBeginSerializable starts new transaction with Serializable isolation level
		MustBeginSerializable()
		// Commit commits the changes made within the transaction
		Commit() error
		// Rollback rolls the transaction back
		Rollback() error
		// ExecScript allows to execute the sql statements from the file provided
		ExecScript(sqlScript string) error
	}

	// ModelTx provides a transaction with some methods for accessing to the Model objects
	ModelTx interface {
		Tx

		// CreateFormat creates format entry
		CreateFormat(format Format) (Format, error)
		// GetFormat retrieves format entry by ID
		GetFormat(ID string) (Format, error)
		// DeleteFormat deletes format entry by ID (only if not referenced)
		DeleteFormat(ID string) error
		// ListFormats lists all the existing format entries
		ListFormats() ([]Format, error)

		// CreateIndex creates index entry based on source ID
		CreateIndex(index Index) (Index, error)
		// GetIndex retrieves index info by ID
		GetIndex(ID string) (Index, error)
		// UpdateIndex updates index info
		UpdateIndex(index Index) error
		// DeleteIndex deletes index entry and all the related records
		DeleteIndex(ID string) error
		// QueryIndexes lists query matching index entries
		QueryIndexes(query IndexQuery) (QueryResult[Index, string], error)

		// UpsertIndexRecords creates or updates index record entries
		UpsertIndexRecords(records ...IndexRecord) error
		// GetIndexRecord retrieves index record entry
		GetIndexRecord(ID, indexID string) (IndexRecord, error)
		// UpdateIndexRecord updates index record
		UpdateIndexRecord(record IndexRecord) error
		// DeleteIndexRecords deletes index record entries
		DeleteIndexRecords(records ...IndexRecord) (int, error)
		// QueryIndexRecords lists query matching index record entries
		QueryIndexRecords(query IndexRecordQuery) (QueryResult[IndexRecord, string], error)

		// Search performs full text search across existing index records
		// the query string should be formed in accordance with the groonga manual
		// for the `&@~` operator, useful links:
		// - https://pgroonga.github.io/reference/operators/query-v2.html
		// - https://groonga.org/docs/reference/grn_expr/query_syntax.html
		Search(query SearchQuery) (QueryResult[SearchQueryResultItem, string], error)
	}

	// Db interface exposes
	Db interface {
		// NewModelTx creates new ModelTx object
		NewModelTx(ctx context.Context) ModelTx
		// NewTx creates Tx object
		NewTx(ctx context.Context) Tx
	}
)
