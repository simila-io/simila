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
	mtx := s.db.NewModelTx(context.Background())

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(s.T(), err)

	frmt, err := mtx.CreateFormat(Format{ID: "pdf", Basis: bas})
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", frmt.ID)

	_, err = mtx.CreateFormat(Format{ID: "pdf", Basis: bas})
	assert.ErrorIs(s.T(), err, errors.ErrExist)

	frmt, err = mtx.GetFormat("pdf")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "pdf", frmt.ID)

	_, err = mtx.GetFormat("notFound")
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)

	fmts, err := mtx.ListFormats()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(fmts))

	err = mtx.DeleteFormat(frmt.ID)
	assert.Nil(s.T(), err)
	err = mtx.DeleteFormat(frmt.ID)
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)
}

func (s *pureSqlTestSuite) TestIndex() {
	mtx := s.db.NewModelTx(context.Background())

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(s.T(), err)
	_, err = mtx.CreateFormat(Format{ID: "pdf", Basis: bas})
	assert.Nil(s.T(), err)

	idx, err := mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "abc.txt", idx.ID)

	idx, err = mtx.GetIndex(idx.ID)
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
	mtx := s.db.NewModelTx(context.Background())

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(s.T(), err)
	vec, err := json.Marshal([]any{7, "word"})
	assert.Nil(s.T(), err)

	_, err = mtx.CreateFormat(Format{ID: "pdf", Basis: bas})
	assert.Nil(s.T(), err)
	idx, err := mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.Nil(s.T(), err)
	_, err = mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.ErrorIs(s.T(), err, errors.ErrExist)

	err = mtx.UpsertIndexRecords(IndexRecord{ID: "123", IndexID: idx.ID, Segment: "haha", Vector: vec},
		IndexRecord{ID: "456", IndexID: idx.ID, Segment: "hello world", Vector: vec},
		IndexRecord{ID: "789", IndexID: idx.ID, Segment: "no no я француз", Vector: vec})
	assert.Nil(s.T(), err)
	err = mtx.UpsertIndexRecords(IndexRecord{ID: "456", IndexID: idx.ID, Segment: "hello world", Vector: vec})
	assert.Nil(s.T(), err)

	rec, err := mtx.GetIndexRecord("789", idx.ID)
	assert.Nil(s.T(), err)
	_, err = mtx.GetIndexRecord("notFound", idx.ID)
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)

	rec.Segment = "no no я француз haha too"
	err = mtx.UpdateIndexRecord(rec)
	assert.Nil(s.T(), err)

	res, err := mtx.QueryIndexRecords(IndexRecordQuery{IndexIDs: []string{idx.ID}, CreatedBefore: time.Now(), Limit: 2})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(3), res.Total)
	assert.Equal(s.T(), 2, len(res.Items))
	assert.Equal(s.T(), IndexRecordID{IndexID: idx.ID, RecordID: "789"}.Encode(), res.NextID)

	n, err := mtx.DeleteIndexRecords(rec)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, n)
	n, err = mtx.DeleteIndexRecords(rec)
	assert.ErrorIs(s.T(), err, errors.ErrNotExist)
	assert.Equal(s.T(), 0, n)
}

func (s *pureSqlTestSuite) TestSearch() {
	mtx := s.db.NewModelTx(context.Background())

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(s.T(), err)
	vec, err := json.Marshal([]any{7, "word"})
	assert.Nil(s.T(), err)

	_, err = mtx.CreateFormat(Format{ID: "pdf", Basis: bas})
	assert.Nil(s.T(), err)
	_, err = mtx.CreateFormat(Format{ID: "doc", Basis: bas})
	assert.Nil(s.T(), err)
	idx1, err := mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.Nil(s.T(), err)
	idx2, err := mtx.CreateIndex(Index{ID: "def.txt", Format: "doc", Tags: Tags{"org": "123"}})
	assert.Nil(s.T(), err)

	err = mtx.UpsertIndexRecords(
		IndexRecord{ID: "123", IndexID: idx1.ID, Segment: "ha haha", Vector: vec},
		IndexRecord{ID: "456", IndexID: idx1.ID, Segment: "hello world", Vector: vec},
		IndexRecord{ID: "789", IndexID: idx1.ID, Segment: "ha no no я Français", Vector: vec},
		IndexRecord{ID: "101", IndexID: idx2.ID, Segment: "万事如意 ha", Vector: vec},
		IndexRecord{ID: "111", IndexID: idx2.ID, Segment: "ping pong pung", Vector: vec},
		IndexRecord{ID: "121", IndexID: idx2.ID, Segment: "pong pung", Vector: vec},
		IndexRecord{ID: "131", IndexID: idx2.ID, Segment: "pung", Vector: vec})
	assert.Nil(s.T(), err)

	res1, err := mtx.Search(SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID}, Query: "(HELLO OR Français OR 如意) (-haha)", Limit: 2})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(3), res1.Total)
	assert.Equal(s.T(), 2, len(res1.Items))
	assert.Equal(s.T(), IndexRecordID{IndexID: idx2.ID, RecordID: "101"}.Encode(), res1.NextID)

	res2, err := mtx.Search(SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID}, Query: "ha", Distinct: true, Limit: 2})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(2), res2.Total)
	assert.Equal(s.T(), 2, len(res2.Items))
	assert.Equal(s.T(), "", res2.NextID)

	res3, err := mtx.Search(SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID}, Query: "ping OR pong OR pung", OrderByScore: true, Limit: 10})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(3), res3.Total)
	assert.Equal(s.T(), 3, len(res3.Items))
	assert.Equal(s.T(), "", res3.NextID)
	assert.Equal(s.T(), 3, res3.Items[0].Score)
	assert.Equal(s.T(), "ping pong pung", res3.Items[0].Segment)
	assert.Equal(s.T(), 2, res3.Items[1].Score)
	assert.Equal(s.T(), "pong pung", res3.Items[1].Segment)
	assert.Equal(s.T(), 1, res3.Items[2].Score)
	assert.Equal(s.T(), "pung", res3.Items[2].Segment)

	res4, err := mtx.Search(SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID}, Query: "ping OR pong OR pung", OrderByScore: true, Offset: 1, Limit: 2})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(3), res4.Total)
	assert.Equal(s.T(), 2, len(res4.Items))
	assert.Equal(s.T(), "", res4.NextID)
	assert.Equal(s.T(), 2, res4.Items[0].Score)
	assert.Equal(s.T(), "pong pung", res4.Items[0].Segment)
	assert.Equal(s.T(), 1, res4.Items[1].Score)
	assert.Equal(s.T(), "pung", res4.Items[1].Segment)

	res5, err := mtx.Search(SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID}, Query: "ping OR pong OR pung OR ha", OrderByScore: true, Distinct: true, Limit: 10})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(4), res5.Total)
	assert.Equal(s.T(), 4, len(res5.Items))
	assert.Equal(s.T(), "", res5.NextID)
	assert.Equal(s.T(), 3, res5.Items[0].Score)
	assert.Equal(s.T(), "ping pong pung", res5.Items[0].Segment)
	assert.Equal(s.T(), 2, res5.Items[1].Score)
	assert.Equal(s.T(), "pong pung", res5.Items[1].Segment)
	assert.Equal(s.T(), 1, res5.Items[2].Score)
	assert.Equal(s.T(), "ha haha", res5.Items[2].Segment)
	assert.Equal(s.T(), 1, res5.Items[3].Score)
	assert.Equal(s.T(), "万事如意 ha", res5.Items[3].Segment)
}

func (s *pureSqlTestSuite) TestConstraints() {
	mtx := s.db.NewModelTx(context.Background())

	bas, err := json.Marshal([]map[string]any{})
	assert.Nil(s.T(), err)

	_, err = mtx.CreateFormat(Format{ID: "pdf", Basis: bas})
	assert.Nil(s.T(), err)
	_, err = mtx.CreateIndex(Index{ID: "abc.txt", Format: "pdf", Tags: Tags{"key": "val"}})
	assert.Nil(s.T(), err)

	err = mtx.DeleteFormat("pdf")
	assert.ErrorIs(s.T(), err, errors.ErrConflict)
}
