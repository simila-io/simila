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
	"encoding/json"
	"github.com/acquirecloud/golibs/errors"
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

func (s *pureSqlTestSuite) TestFormat() {
	mtx := s.db.NewModelTx()

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(s.T(), err)

	frmtID, err := mtx.CreateFormat(Format{Name: "pdf", Basis: bas})
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", frmtID)

	_, err = mtx.CreateFormat(Format{Name: "pdf", Basis: bas})
	assert.ErrorIs(s.T(), err, errors.ErrExist)

	frmt, err := mtx.GetFormat("pdf")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "pdf", frmt.Name)

	_, err = mtx.GetFormat("notFound")
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)

	fmts, err := mtx.ListFormats()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(fmts))

	err = mtx.DeleteFormat(frmt.Name)
	assert.Nil(s.T(), err)
	err = mtx.DeleteFormat(frmt.Name)
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)
}

func (s *pureSqlTestSuite) TestIndex() {
	mtx := s.db.NewModelTx()

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(s.T(), err)
	_, err = mtx.CreateFormat(Format{Name: "pdf", Basis: bas})
	assert.Nil(s.T(), err)

	idxID, err := mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "abc.txt", idxID)

	idx, err := mtx.GetIndex(idxID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "abc.txt", idx.ID)

	_, err = mtx.GetIndex("notFound")
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)

	idx.Tags["key1"] = "val1"
	err = mtx.UpdateIndex(idx)
	assert.Nil(s.T(), err)

	res, err := mtx.QueryIndexes(IndexQuery{Format: "pdf", Tags: idx.Tags, Limit: 10})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(1), res.Total)
	assert.Equal(s.T(), 1, len(res.Items))
	assert.Equal(s.T(), "", res.NextID)

	err = mtx.DeleteIndex(idx.ID)
	assert.Nil(s.T(), err)
	err = mtx.DeleteIndex(idx.ID)
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)
}

func (s *pureSqlTestSuite) TestIndexRecord() {
	mtx := s.db.NewModelTx()

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(s.T(), err)
	vec, err := json.Marshal([]any{7, "word"})
	assert.Nil(s.T(), err)

	_, err = mtx.CreateFormat(Format{Name: "pdf", Basis: bas})
	assert.Nil(s.T(), err)
	idxID, err := mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.Nil(s.T(), err)
	_, err = mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.ErrorIs(s.T(), err, errors.ErrExist)

	err = mtx.CreateIndexRecords([]IndexRecord{{ID: "123", IndexID: idxID, Segment: "haha", Vector: vec}})
	assert.Nil(s.T(), err)
	err = mtx.CreateIndexRecords([]IndexRecord{{ID: "456", IndexID: idxID, Segment: "hello world", Vector: vec}})
	assert.Nil(s.T(), err)
	err = mtx.CreateIndexRecords([]IndexRecord{{ID: "789", IndexID: idxID, Segment: "no no я француз", Vector: vec}})
	assert.Nil(s.T(), err)

	rec, err := mtx.GetIndexRecord("789")
	assert.Nil(s.T(), err)
	_, err = mtx.GetIndexRecord("notFound")
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)

	rec.Segment = "no no я француз haha too"
	err = mtx.UpdateIndexRecord(rec)
	assert.Nil(s.T(), err)

	res, err := mtx.QueryIndexRecords(IndexRecordQuery{IndexIDs: []string{idxID}, CreatedBefore: time.Now(), Limit: 2})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(3), res.Total)
	assert.Equal(s.T(), 2, len(res.Items))
	assert.Equal(s.T(), IndexRecordID{IndexID: idxID, RecordID: "789"}.Encode(), res.NextID)

	err = mtx.DeleteIndexRecord(rec.ID)
	assert.Nil(s.T(), err)
	err = mtx.DeleteIndexRecord(rec.ID)
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)
}

func (s *pureSqlTestSuite) TestSearch() {
	mtx := s.db.NewModelTx()

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(s.T(), err)
	vec, err := json.Marshal([]any{7, "word"})
	assert.Nil(s.T(), err)

	_, err = mtx.CreateFormat(Format{Name: "pdf", Basis: bas})
	assert.Nil(s.T(), err)
	_, err = mtx.CreateFormat(Format{Name: "doc", Basis: bas})
	assert.Nil(s.T(), err)
	idx1ID, err := mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.Nil(s.T(), err)
	idx2ID, err := mtx.CreateIndex(Index{ID: "def.txt", Format: "doc", Tags: Tags{"org": "123"}})
	assert.Nil(s.T(), err)

	err = mtx.CreateIndexRecords([]IndexRecord{{ID: "123", IndexID: idx1ID, Segment: "ha haha", Vector: vec}})
	assert.Nil(s.T(), err)
	err = mtx.CreateIndexRecords([]IndexRecord{{ID: "123", IndexID: idx1ID, Segment: "ha haha", Vector: vec}})
	assert.ErrorIs(s.T(), err, errors.ErrExist)
	err = mtx.CreateIndexRecords([]IndexRecord{{ID: "456", IndexID: idx1ID, Segment: "hello world", Vector: vec}})
	assert.Nil(s.T(), err)
	err = mtx.CreateIndexRecords([]IndexRecord{{ID: "789", IndexID: idx1ID, Segment: "ha no no я Français", Vector: vec}})
	assert.Nil(s.T(), err)
	err = mtx.CreateIndexRecords([]IndexRecord{{ID: "101", IndexID: idx2ID, Segment: "万事如意 ha", Vector: vec}})
	assert.Nil(s.T(), err)

	res1, err := mtx.Search(SearchQuery{IndexIDs: []string{idx1ID, idx2ID}, Query: "(HELLO OR Français OR 如意) (-haha)", Limit: 2})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(3), res1.Total)
	assert.Equal(s.T(), 2, len(res1.Items))
	assert.Equal(s.T(), IndexRecordID{IndexID: idx2ID, RecordID: "101"}.Encode(), res1.NextID)

	res2, err := mtx.Search(SearchQuery{IndexIDs: []string{idx1ID, idx2ID}, Query: "ha", Distinct: true, Limit: 2})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(2), res2.Total)
	assert.Equal(s.T(), 2, len(res2.Items))
	assert.Equal(s.T(), "", res2.NextID)
}

func (s *pureSqlTestSuite) TestConstraints() {
	mtx := s.db.NewModelTx()

	bas, err := json.Marshal([]map[string]any{})
	assert.Nil(s.T(), err)

	_, err = mtx.CreateFormat(Format{Name: "pdf", Basis: bas})
	assert.Nil(s.T(), err)
	_, err = mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.Nil(s.T(), err)

	err = mtx.DeleteFormat("pdf")
	assert.ErrorIs(s.T(), err, errors.ErrConflict)
}
