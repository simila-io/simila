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

		// CreateNodes allows to create one or several new Nodes. If a Node with the path already exists,
		// the function will return ErrExists.
		CreateNodes(nodes ...Node) ([]Node, error)
		// ListNodes returns all nodes for the path. For example for the path="/a/b/doc.txt"
		// the result nodes will be {<"/", "a">, {<"/a/", "b">, <"/a/b/", "doc.txt">}
		ListNodes(path string) ([]Node, error)
		// ListChildren returns the children for the path: all nodes of the 1st level with the path prefix in Path
		//
		// Example: for the nodes={<"/", "a">, <"/a/", "c">, <"/a/", "b">, <"/a/b/", "cc">}
		// the result for the "/a/" will be {<"/a/", "c">, <"/a/", "b">}
		ListChildren(path string) ([]Node, error)

		// GetNode returns the node by its fqnp
		GetNode(fqnp string) (Node, error)
		// UpdateNode updates node data
		UpdateNode(node Node) error

		// DeleteNodes deletes the Nodes that matches to the DeleteNodesQuery and all the records associated with the nodes.
		// force allows to delete folder nodes with children. If the node is a folder, and there are children,
		// but the force flag is false, the function will return ErrConflict error
		//
		// NOTE: The operation is not atomic until an external transaction is not started,
		// the caller MUST start the transaction before using this method.
		DeleteNodes(DeleteNodesQuery) error

		// UpsertIndexRecords creates or updates index record entries. It returns the new records created
		UpsertIndexRecords(records ...IndexRecord) (int64, error)
		// DeleteIndexRecords deletes index record entries
		DeleteIndexRecords(records ...IndexRecord) (int64, error)
		// QueryIndexRecords lists query matching index record entries
		QueryIndexRecords(query IndexRecordQuery) (QueryResult[IndexRecord, string], error)

		// Search performs search across existing index records
		// the query string should be formed in accordance with the query
		// language of the underlying search engine
		//
		// NOTE: The operation is not atomic until an external transaction is not started,
		// the caller MUST start the transaction before using this method.
		Search(query SearchQuery) (SearchQueryResult, error)
	}

	// Db interface exposes
	Db interface {
		// NewModelTx creates new ModelTx object
		NewModelTx(ctx context.Context) ModelTx
		// NewTx creates Tx object
		NewTx(ctx context.Context) Tx
	}
)
