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
		MustBeginSerializable(ctx context.Context)
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
		CreateFormat(format Format) (string, error)
		// GetFormat retrieves format entry by name
		GetFormat(name string) (Format, error)
		// DeleteFormat deletes format entry by name (only if not referenced)
		DeleteFormat(name string) error
		// ListFormats lists all the existing format entries
		ListFormats() ([]Format, error)

		// CreateIndex creates index entry based on source ID
		CreateIndex(index Index) (string, error)
		// GetIndex retrieves index info by ID
		GetIndex(ID string) (Index, error)
		// UpdateIndex updates index info
		UpdateIndex(index Index) error
		// DeleteIndex deletes index entry and all the related records
		DeleteIndex(ID string) error
		// QueryIndexes lists query matching index entries
		QueryIndexes(query IndexQuery) (QueryResult[Index, string], error)

		// CreateIndexRecords creates index records entry
		CreateIndexRecords(records []IndexRecord) error
		// GetIndexRecord retrieves index record entry by ID
		GetIndexRecord(ID string) (IndexRecord, error)
		// UpdateIndexRecord updates index record
		UpdateIndexRecord(record IndexRecord) error
		// DeleteIndexRecord deletes index record by ID
		DeleteIndexRecord(ID string) error
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
		NewModelTx() ModelTx
		// NewTx creates Tx object
		NewTx() Tx
	}
)
