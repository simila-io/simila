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
	"strings"
	"time"
)

type (
	Format struct {
		ID        string    `db:"id"`
		Basis     []byte    `db:"basis"`
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
		Vector    []byte    `db:"vector"`
		Format    string    `db:"format"`
		RankMult  float64   `db:"rank_multiplier"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`

		// search module specific fields
		SegmentTsVector string `db:"segment_tsvector"`
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
		TextQuery        string
		FilterConditions string
		GroupByPathOff   bool // GroupByPathOff turns off results grouping by path.
		Offset           int
		Limit            int
	}

	SearchQueryResultItem struct {
		IndexRecord
		Path                string   `db:"path"`
		MatchedKeywordsList []string // mapped manually after filling the MatchedKeywords
		MatchedKeywords     string   `db:"matched_keywords"`
		Score               float32  `db:"score"`
	}

	SearchQueryResult struct {
		Items []SearchQueryResultItem
		Total int64
	}

	// DeleteNodesQuery provides parameters for deleting multiple nodes
	DeleteNodesQuery struct {
		// FilterConditions contains the node selection filter
		FilterConditions string
		// Force flag allows to delete children for the selected folders. For example, if
		// the selected node is a folder, and it has children, that don't match the filter criteria,
		// they also will be deleted if the force is true.
		Force bool
	}

	// ListNodesQuery allows to select nodes by the condition provided
	ListNodesQuery struct {
		// FilterConditions contains the node selection filter
		FilterConditions string
		Offset           int64
		Limit            int64
	}

	QueryResult[T any, N any] struct {
		Items  []T
		NextID N
		Total  int64
	}
)

const (
	NodeFlagFolder   = 0
	NodeFlagDocument = 1 // if it is set, it is a document. If not set, it is a folder
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

func (t Tags) JSON() string {
	var sb strings.Builder
	sb.WriteString("{")
	if len(t) > 0 {
		for k, v := range t {
			if sb.Len() > 1 {
				sb.WriteByte(',')
			}
			sb.WriteString(fmt.Sprintf("%q:%q", k, v))
		}
	}
	sb.WriteString("}")
	return sb.String()
}
