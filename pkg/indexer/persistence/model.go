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
	"encoding/base64"
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

	Index struct {
		ID        string    `db:"id"`
		Format    string    `db:"format"`
		Tags      Tags      `db:"tags"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	IndexQuery struct {
		Format        string
		Tags          Tags
		CreatedAfter  time.Time
		CreatedBefore time.Time
		FromID        string
		Limit         int
	}

	IndexRecord struct {
		ID        string    `db:"id"`
		IndexID   string    `db:"index_id"`
		Segment   string    `db:"segment"`
		Vector    Vector    `db:"vector"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	IndexRecordQuery struct {
		IndexIDs      []string
		CreatedAfter  time.Time
		CreatedBefore time.Time
		FromID        string
		Limit         int
	}

	SearchQuery struct {
		IndexIDs []string
		Query    string // underlying search engine query
		Tags     Tags   // index tags
		Distinct bool   // if true, returns at most 1 result per index
		FromID   string
		Limit    int
	}

	SearchQueryResultItem struct {
		IndexRecord
		Score int
	}

	QueryResult[T any, N any] struct {
		Items  []T
		NextID N
		Total  int64
	}

	IndexRecordID struct {
		IndexID  string `json:"index_id"`
		RecordID string `json:"record_id"`
	}
)

func (id IndexRecordID) Encode() string {
	if len(id.RecordID) == 0 && len(id.IndexID) == 0 {
		return ""
	}
	bb, err := json.Marshal(id)
	if err != nil {
		panic(fmt.Sprintf("failed to json marshal IndexRecordID: %v", err))
	}
	return base64.StdEncoding.EncodeToString(bb)
}

func (id *IndexRecordID) Decode(s string) error {
	if len(s) == 0 {
		return nil
	}
	bb, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return fmt.Errorf("failed to base64 decode IndexRecordID: %v", err)
	}
	if err = json.Unmarshal(bb, id); err != nil {
		return fmt.Errorf("failed to json unmashal IndexRecordID: %v", err)
	}
	return nil
}

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
