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
	"bytes"
	"crypto/sha1"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"time"
)

type (
	StrStrMap map[string]string

	Format struct {
		ID        string    `db:"id"`
		Name      string    `db:"name"`
		Basis     StrStrMap `db:"basis"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	Index struct {
		ID        string    `db:"id"`
		Format    string    `db:"format"`
		Tags      StrStrMap `db:"tags"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	IndexQuery struct {
		Format        string
		Tags          StrStrMap
		CreatedAfter  time.Time
		CreatedBefore time.Time
		FromID        string
		Limit         int
	}

	IndexRecord struct {
		ID        string    `db:"id"`
		IndexID   string    `db:"index_id"`
		Segment   string    `db:"segment"`
		Vector    StrStrMap `db:"vector"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	IndexRecordQuery struct {
		IndexIDs      []string
		Tags          StrStrMap // index tags
		Query         string    // underlying search engine query
		Distinct      bool      // if true, returns at most 1 result per index
		CreatedAfter  time.Time
		CreatedBefore time.Time
		FromID        string
		Limit         int
	}

	QueryResult[T any, N any] struct {
		Items  []T
		NextID N
		Total  int64
	}
)

// RecordID converts a tuple of (index ID, record vector) to index record ID
func RecordID(indexID string, vector StrStrMap) (string, error) {
	if len(indexID) == 0 || len(vector) == 0 {
		return "", fmt.Errorf("indexID and record vector must be specified: %w", errors.ErrInvalid)
	}
	var bb bytes.Buffer
	bb.WriteString(indexID)
	bb.Write(mustEncode(vector))
	hSum := sha1.Sum(bb.Bytes())
	return fmt.Sprintf("%x", hSum), nil
}

func (m StrStrMap) Value() (value driver.Value, err error) {
	return json.Marshal(m)
}

func (m *StrStrMap) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unexpected value in scan")
	}
	return json.Unmarshal(b, &m)
}
