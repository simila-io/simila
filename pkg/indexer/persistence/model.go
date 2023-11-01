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
	"github.com/acquirecloud/golibs/errors"
	"time"
)

const (
	DTypeNumber Type = "number"
	DTypeString Type = "string"
)

type (
	Type string

	Dimension struct {
		Name string `json:"name"`
		Type Type   `json:"type"`
		Min  int64  `json:"min"`
		Max  int64  `json:"max"`
	}

	Basis []Dimension

	Component any

	Vector []Component

	Format struct {
		ID        string    `db:"id"`
		Name      string    `db:"name"`
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
		Tags          Tags   // index tags
		Query         string // underlying search engine query
		Distinct      bool   // if true, returns at most 1 result per index
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

func FromNumber(f float64) Component {
	return f
}

func FromString(s string) Component {
	return s
}

func DType(c Component) Type {
	switch c.(type) {
	case string:
		return DTypeString
	case float64:
		return DTypeNumber
	}
	return ""
}

func NewBasis(dims ...Dimension) (Basis, error) {
	if len(dims) == 0 {
		return Basis{}, fmt.Errorf("basis must have non-zero number of dimentions: %w", errors.ErrInvalid)
	}
	for i := 0; i < len(dims); i++ {
		if len(dims[i].Name) == 0 {
			return Basis{}, fmt.Errorf("dimention=%v name must be non-empty: %w", dims[i], errors.ErrInvalid)
		}
		if dims[i].Min == dims[i].Max {
			return Basis{}, fmt.Errorf("dimention=%v min max values must be different: %w", dims[i], errors.ErrInvalid)
		}
		if dims[i].Type != DTypeString && dims[i].Type != DTypeNumber {
			return Basis{}, fmt.Errorf("dimention=%v type must be either %q or %q: %w",
				dims[i], DTypeString, DTypeNumber, errors.ErrInvalid)
		}
	}
	return dims, nil
}

func NewVector(basis Basis, comps ...Component) (Vector, error) {
	if len(comps) != len(basis) {
		return Vector{}, fmt.Errorf("# of vector components=%d must match # of basis dimensions=%d: %w",
			len(comps), len(basis), errors.ErrInvalid)
	}
	for i := 0; i < len(comps); i++ {
		if basis[i].Type != DType(comps[i]) {
			return Vector{}, fmt.Errorf("component=%q type does not match the basis dimension=%v type: %w",
				comps[i], basis[i], errors.ErrInvalid)
		}
		switch DType(comps[i]) {
		case DTypeString:
			v := comps[i].(string)
			if int64(len(v)) < basis[i].Min || int64(len(v)) > basis[i].Max {
				return Vector{}, fmt.Errorf("component=%q value does not meet the basis dimension=%v constraints: %w",
					comps[i], basis[i], errors.ErrInvalid)
			}
		case DTypeNumber:
			v := int64(comps[i].(float64))
			if v < basis[i].Min || v > basis[i].Max {
				return Vector{}, fmt.Errorf("component=%q value does not meet the basis dimension=%v constraints: %w",
					comps[i], basis[i], errors.ErrInvalid)
			}
		default:
			return Vector{}, fmt.Errorf("unknown component=%q type=%s: %w",
				comps[i], DType(comps[i]), errors.ErrInvalid)
		}
	}
	return comps, nil
}

func (v Vector) Value() (value driver.Value, err error) {
	return json.Marshal(v)
}

func (v *Vector) Scan(value any) error {
	buf, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("not []byte value in scan")
	}
	return json.Unmarshal(buf, v)
}

func (b Basis) Value() (value driver.Value, err error) {
	return json.Marshal(b)
}

func (b *Basis) Scan(value any) error {
	buf, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("not []byte value in scan")
	}
	return json.Unmarshal(buf, b)
}

func (t Tags) Value() (value driver.Value, err error) {
	return json.Marshal(t)
}

func (t *Tags) Scan(value any) error {
	buf, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("not []byte value in scan")
	}
	return json.Unmarshal(buf, &t)
}
