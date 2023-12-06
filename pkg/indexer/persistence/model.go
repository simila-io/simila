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
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type (
	Basis json.RawMessage

	Vector json.RawMessage

	Format struct {
		ID        string    `db:"id"`
		Basis     Basis     `db:"basis"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	Tags map[string]string

	// Node describes an object. The node has Path and Name, and the pair <Path, Name> must
	// be unique within the tree. Name cannot be empty.
	//
	// Root path is "/". The 3rd level path, for example, could be "/aaa/bbb/ccc/". Path is always ends by "/"
	Node struct {
		ID   int64  `db:"id"`
		Path string `db:"path"`
		// Name is either name of the folder or a document
		Name      string    `db:"name"`
		Tags      Tags      `db:"tags"`
		Flags     int32     `db:"flags"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	IndexRecord struct {
		ID        string    `db:"id"`
		NodeID    int64     `db:"node_id"`
		Segment   string    `db:"segment"`
		Vector    Vector    `db:"vector"`
		Format    string    `db:"format"`
		RankMult  float64   `db:"rank_multiplier"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	IndexRecordQuery struct {
		// Format is filter records by the format. If empty no format filter
		Format string
		// NodeID sets the ID of the node
		NodeID        int64
		CreatedAfter  time.Time
		CreatedBefore time.Time
		FromID        string
		Limit         int
	}

	SearchQuery struct {
		Path  string
		Query string // underlying search engine query
		Tags  Tags   // index tags
		// Strict defines the search records behavior:
		// - If true, the search will select between all records associated with the Path node only (Path == "PathXName" for the Node with <PathX, Name> pair)
		// - if false, the search will select between records in the Path subtree for all records there (every Node which's path has Path prefix),
		// but only one most relevant record per one node (one record per <PathX, Name> pair)
		Strict bool
		Offset int
		Limit  int
	}

	SearchQueryResultItem struct {
		IndexRecord
		Path                string
		MatchedKeywordsList []string // mapped manually after filling the MatchedKeywords
		MatchedKeywords     string   `db:"matched_keywords"`
		Score               float32  `db:"score"`
	}

	QueryResult[T any, N any] struct {
		Items  []T
		NextID N
		Total  int64
	}
)

const (
	NODE_FLAG_DOCUMENT = 1 // if it is set, it is a document. If not set, it is a folder
)

func (t Tags) Value() (value driver.Value, err error) {
	return json.Marshal(t)
}

func (t *Tags) Scan(value any) error {
	buf, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("not a []byte value in scan")
	}
	return json.Unmarshal(buf, &t)
}
