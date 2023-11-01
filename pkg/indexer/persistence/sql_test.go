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
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type pureSqlTestSuite struct {
	SqlTestSuite
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(pureSqlTestSuite))
}

func (s *pureSqlTestSuite) TestCreateIndexRecord() {
	// format
	mtx := s.db.NewModelTx()
	bas, err := NewBasis(Dimension{Name: "page", Type: DTypeNumber, Min: 0, Max: 10}, Dimension{Name: "mark", Type: DTypeString, Min: 3, Max: 20})
	assert.Nil(s.T(), err)
	frmt := Format{Name: "pdf", Basis: bas}
	frmtID, err := mtx.CreateFormat(frmt)
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", frmtID)
	fmts, err := mtx.ListFormats()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(fmts))
	idx := Index{ID: "abc.txt", Format: frmt.Name, Tags: Tags{"key": "val"}}

	// index
	idxID, err := mtx.CreateIndex(idx)
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", idxID)
	idx, err = mtx.GetIndex(idxID)
	assert.Nil(s.T(), err)
	idx.Tags["key1"] = "val1"
	err = mtx.UpdateIndex(idx)
	assert.Nil(s.T(), err)
	res1, err := mtx.QueryIndexes(IndexQuery{Format: "pdf", Tags: idx.Tags, Limit: 10})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(1), res1.Total)
	assert.Equal(s.T(), 1, len(res1.Items))
	assert.Equal(s.T(), "", res1.NextID)

	// index record
	vec, err := NewVector(frmt.Basis, FromNumber(7), FromString("word"))
	assert.Nil(s.T(), err)
	rec := IndexRecord{ID: "123", IndexID: idxID, Segment: "haha", Vector: vec}
	recID, err := mtx.CreateIndexRecord(rec)
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", recID)
	rec = IndexRecord{ID: "456", IndexID: idxID, Segment: "hello world", Vector: vec}
	recID, err = mtx.CreateIndexRecord(rec)
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", recID)
	rec = IndexRecord{ID: "789", IndexID: idxID, Segment: "no no я француз", Vector: vec}
	recID, err = mtx.CreateIndexRecord(rec)
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", recID)
	rec, err = mtx.GetIndexRecord(recID)
	assert.Nil(s.T(), err)
	rec.Segment = "no no я француз haha too"
	err = mtx.UpdateIndexRecord(rec)
	assert.Nil(s.T(), err)
	res2, err := mtx.QueryIndexRecords(IndexRecordQuery{IndexIDs: []string{idx.ID}, CreatedBefore: time.Now(), Limit: 2})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(3), res2.Total)
	assert.Equal(s.T(), 2, len(res2.Items))
	assert.Equal(s.T(), "789", res2.NextID)

	// search
	res3, err := mtx.Search(SearchQuery{IndexIDs: []string{idx.ID}, Query: "hello OR француз", Limit: 1})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(2), res3.Total)
	assert.Equal(s.T(), 1, len(res3.Items))
	assert.Equal(s.T(), "789", res3.NextID)

	// delete
	err = mtx.DeleteIndexRecord(rec.ID)
	assert.Nil(s.T(), err)
	err = mtx.DeleteIndex(idx.ID)
	assert.Nil(s.T(), err)
	err = mtx.DeleteFormat(frmt.Name)
	assert.Nil(s.T(), err)
}
