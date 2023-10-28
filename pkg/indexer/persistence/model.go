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
	AnyMap map[string]any

	Index struct {
		ID        string    `db:"id"`
		Format    string    `db:"format"`
		Tags      AnyMap    `db:"tags"`
		CreatedAt time.Time `db:"created_at"`
	}

	IndexRecord struct {
		ID      string `db:"id"`
		IndexID string `db:"index_id"`
		Segment string `db:"segment"`
		Vector  AnyMap `db:"vector"`
	}

	QueryResult[T any] struct {
		Items []T
		Total int64
	}
)

func (m AnyMap) Value() (value driver.Value, err error) {
	return json.Marshal(m)
}

func (m *AnyMap) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unexpected value in scan")
	}
	return json.Unmarshal(b, &m)
}
