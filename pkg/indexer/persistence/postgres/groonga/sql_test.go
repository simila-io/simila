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

package groonga

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type (
	pgTestSuite struct {
		suite.Suite
		cDb persistence.ContainerDb
		db  *Db
	}
)

func TestRunSuite(t *testing.T) {
	c, err := persistence.NewPgContainerDb("groonga/pgroonga:latest-debian-16", persistence.WithDbName("simila_test"))
	assert.Nil(t, err)
	//c, err := persistence.NewNilContainerDb(persistence.WithDbName("simila_test"))
	//assert.Nil(t, err)
	suite.Run(t, newPqTestSuite(c))
}

func newPqTestSuite(cDb persistence.ContainerDb) *pgTestSuite {
	return &pgTestSuite{cDb: cDb}
}

func (ts *pgTestSuite) SetupSuite() {
}

func (ts *pgTestSuite) TearDownSuite() {
	if ts.db != nil {
		ts.db.Shutdown()
	}
	if ts.cDb != nil {
		_ = ts.cDb.Close()
	}
}

func (ts *pgTestSuite) BeforeTest(suiteName, testName string) {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	dbCfg := ts.cDb.DbConfig()
	assert.Nil(ts.T(), ts.dropCreatePgDb(ctx))

	ts.db = NewDb(dbCfg.DataSourceFull())
	assert.Nil(ts.T(), ts.db.Init(ctx))
}

func (ts *pgTestSuite) AfterTest(suiteName, testName string) {
	if ts.db != nil {
		ts.db.Shutdown()
	}
}

func (ts *pgTestSuite) dropCreatePgDb(ctx context.Context) error {
	dbCfg := ts.cDb.DbConfig()
	dbConn, err := sqlx.ConnectContext(ctx, "postgres", dbCfg.DataSourceNoDb())
	if err != nil {
		return err
	}
	defer func() {
		_ = dbConn.Close()
	}()
	if _, err = dbConn.DB.Exec(fmt.Sprintf("drop database if exists %s with (force)", dbCfg.DbName)); err != nil {
		return err
	}
	if _, err = dbConn.DB.Exec(fmt.Sprintf("create database %s", dbCfg.DbName)); err != nil {
		return err
	}
	return nil
}

func (ts *pgTestSuite) TestFormat() {
	mtx := ts.db.NewModelTx(context.Background())

	fmts, err := mtx.ListFormats()
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), 1, len(fmts)) // 1 = txt (created by default)

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(ts.T(), err)

	frmt, err := mtx.CreateFormat(persistence.Format{ID: "pdf", Basis: bas})
	assert.Nil(ts.T(), err)
	assert.NotEqual(ts.T(), "", frmt.ID)

	_, err = mtx.CreateFormat(persistence.Format{ID: "pdf", Basis: bas})
	assert.ErrorIs(ts.T(), err, errors.ErrExist)

	frmt, err = mtx.GetFormat("pdf")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), "pdf", frmt.ID)

	_, err = mtx.GetFormat("notFound")
	assert.ErrorIs(ts.T(), err, errors.ErrNotExist)

	fmts, err = mtx.ListFormats()
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), 2, len(fmts)) // 2 = pdf + txt (created by default)

	err = mtx.DeleteFormat(frmt.ID)
	assert.Nil(ts.T(), err)
	err = mtx.DeleteFormat(frmt.ID)
	assert.ErrorIs(ts.T(), err, errors.ErrNotExist)
}

func (ts *pgTestSuite) TestIndex() {
	mtx := ts.db.NewModelTx(context.Background())

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(ts.T(), err)
	_, err = mtx.CreateFormat(persistence.Format{ID: "pdf", Basis: bas})
	assert.Nil(ts.T(), err)

	idx, err := mtx.CreateIndex(persistence.Index{ID: "abc.txt", Format: "pdf", Tags: persistence.Tags{"key": "val"}})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), "abc.txt", idx.ID)

	idx, err = mtx.GetIndex(idx.ID)
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), "abc.txt", idx.ID)

	_, err = mtx.GetIndex("notFound")
	assert.ErrorIs(ts.T(), err, errors.ErrNotExist)

	idx.Tags["key1"] = "val1"
	err = mtx.UpdateIndex(idx)
	assert.Nil(ts.T(), err)

	res, err := mtx.QueryIndexes(persistence.IndexQuery{Format: "pdf", Tags: idx.Tags, Limit: 10})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(1), res.Total)
	assert.Equal(ts.T(), 1, len(res.Items))
	assert.Equal(ts.T(), "", res.NextID)

	err = mtx.DeleteIndex(idx.ID)
	assert.Nil(ts.T(), err)
	err = mtx.DeleteIndex(idx.ID)
	assert.ErrorIs(ts.T(), err, errors.ErrNotExist)
}

func (ts *pgTestSuite) TestIndexRecord() {
	mtx := ts.db.NewModelTx(context.Background())

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(ts.T(), err)
	vec, err := json.Marshal([]any{7, "word"})
	assert.Nil(ts.T(), err)

	_, err = mtx.CreateFormat(persistence.Format{ID: "pdf", Basis: bas})
	assert.Nil(ts.T(), err)
	idx, err := mtx.CreateIndex(persistence.Index{ID: "abc.txt", Format: "pdf", Tags: persistence.Tags{"key": "val"}})
	assert.Nil(ts.T(), err)
	_, err = mtx.CreateIndex(persistence.Index{ID: "abc.txt", Format: "pdf", Tags: persistence.Tags{"key": "val"}})
	assert.ErrorIs(ts.T(), err, errors.ErrExist)

	err = mtx.UpsertIndexRecords(persistence.IndexRecord{ID: "123", IndexID: idx.ID, Segment: "haha", Vector: vec},
		persistence.IndexRecord{ID: "456", IndexID: idx.ID, Segment: "hello world", Vector: vec},
		persistence.IndexRecord{ID: "789", IndexID: idx.ID, Segment: "no no я француз", Vector: vec})
	assert.Nil(ts.T(), err)
	err = mtx.UpsertIndexRecords(persistence.IndexRecord{ID: "456", IndexID: idx.ID, Segment: "hello world", Vector: vec})
	assert.Nil(ts.T(), err)

	rec, err := mtx.GetIndexRecord("789", idx.ID)
	assert.Nil(ts.T(), err)
	_, err = mtx.GetIndexRecord("notFound", idx.ID)
	assert.ErrorIs(ts.T(), err, errors.ErrNotExist)

	rec.Segment = "no no я француз haha too"
	err = mtx.UpdateIndexRecord(rec)
	assert.Nil(ts.T(), err)

	res, err := mtx.QueryIndexRecords(persistence.IndexRecordQuery{IndexIDs: []string{idx.ID}, CreatedBefore: time.Now(), Limit: 2})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(3), res.Total)
	assert.Equal(ts.T(), 2, len(res.Items))
	assert.Equal(ts.T(), persistence.IndexRecordID{IndexID: idx.ID, RecordID: "789"}.Encode(), res.NextID)

	n, err := mtx.DeleteIndexRecords(rec)
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), 1, n)
	n, err = mtx.DeleteIndexRecords(rec)
	assert.ErrorIs(ts.T(), err, errors.ErrNotExist)
	assert.Equal(ts.T(), 0, n)
}

func (ts *pgTestSuite) TestSearch() {
	mtx := ts.db.NewModelTx(context.Background())

	bas, err := json.Marshal([]map[string]any{{"Name": "page", "Type": "number"}, {"Name": "mark", "Type": "string"}})
	assert.Nil(ts.T(), err)
	vec, err := json.Marshal([]any{7, "word"})
	assert.Nil(ts.T(), err)

	_, err = mtx.CreateFormat(persistence.Format{ID: "pdf", Basis: bas})
	assert.Nil(ts.T(), err)
	_, err = mtx.CreateFormat(persistence.Format{ID: "doc", Basis: bas})
	assert.Nil(ts.T(), err)
	idx1, err := mtx.CreateIndex(persistence.Index{ID: "abc.txt", Format: "pdf", Tags: persistence.Tags{"key": "val"}})
	assert.Nil(ts.T(), err)
	idx2, err := mtx.CreateIndex(persistence.Index{ID: "def.txt", Format: "doc", Tags: persistence.Tags{"org": "123"}})
	assert.Nil(ts.T(), err)

	err = mtx.UpsertIndexRecords(
		persistence.IndexRecord{ID: "123", IndexID: idx1.ID, Segment: "ha haha", Vector: vec},
		persistence.IndexRecord{ID: "456", IndexID: idx1.ID, Segment: "hello world", Vector: vec},
		persistence.IndexRecord{ID: "789", IndexID: idx1.ID, Segment: "ha no no я Français", Vector: vec},
		persistence.IndexRecord{ID: "101", IndexID: idx2.ID, Segment: "万事如意 ha", Vector: vec},
		persistence.IndexRecord{ID: "111", IndexID: idx2.ID, Segment: "ping pong pung", Vector: vec},
		persistence.IndexRecord{ID: "121", IndexID: idx2.ID, Segment: "pong pung", Vector: vec},
		persistence.IndexRecord{ID: "131", IndexID: idx2.ID, Segment: "pung", Vector: vec})
	assert.Nil(ts.T(), err)

	res1, err := mtx.Search(persistence.SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID},
		Query: "(HELLO OR Français OR 如意) (-haha)", Limit: 2})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(3), res1.Total)
	assert.Equal(ts.T(), 2, len(res1.Items))
	assert.Equal(ts.T(), persistence.IndexRecordID{IndexID: idx2.ID, RecordID: "101"}.Encode(), res1.NextID)

	res2, err := mtx.Search(persistence.SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID},
		Query: "ha", Distinct: true, Limit: 2})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(2), res2.Total)
	assert.Equal(ts.T(), 2, len(res2.Items))
	assert.Equal(ts.T(), "", res2.NextID)

	res3, err := mtx.Search(persistence.SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID},
		Query: "ping OR pong OR pung", OrderByScore: true, Limit: 10})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(3), res3.Total)
	assert.Equal(ts.T(), 3, len(res3.Items))
	assert.Equal(ts.T(), "", res3.NextID)

	assert.Greater(ts.T(), res3.Items[0].Score, float32(2.9))
	assert.Equal(ts.T(), "ping pong pung", res3.Items[0].Segment)

	assert.Greater(ts.T(), res3.Items[1].Score, float32(1.9))
	assert.Equal(ts.T(), "pong pung", res3.Items[1].Segment)

	assert.Greater(ts.T(), res3.Items[2].Score, float32(0.9))
	assert.Equal(ts.T(), "pung", res3.Items[2].Segment)

	res4, err := mtx.Search(persistence.SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID},
		Query: "ping OR pong OR pung", OrderByScore: true, Offset: 1, Limit: 2})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(3), res4.Total)
	assert.Equal(ts.T(), 2, len(res4.Items))
	assert.Equal(ts.T(), "", res4.NextID)

	assert.Greater(ts.T(), res4.Items[0].Score, float32(1.9))
	assert.Equal(ts.T(), "pong pung", res4.Items[0].Segment)

	assert.Greater(ts.T(), res4.Items[1].Score, float32(0.9))
	assert.Equal(ts.T(), "pung", res4.Items[1].Segment)

	res5, err := mtx.Search(persistence.SearchQuery{IndexIDs: []string{idx1.ID, idx2.ID},
		Query: "ping OR pong OR pung OR ha", OrderByScore: true, Distinct: true, Limit: 10})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(5), res5.Total)
	assert.Equal(ts.T(), 5, len(res5.Items))
	assert.Equal(ts.T(), "", res5.NextID)

	assert.Greater(ts.T(), res5.Items[0].Score, float32(2.9))
	assert.Equal(ts.T(), "ha haha", res5.Items[0].Segment)
	assert.Equal(ts.T(), 1, len(res5.Items[0].MatchedKeywordsList))
	assert.Equal(ts.T(), "ha", res5.Items[0].MatchedKeywordsList[0])

	assert.Greater(ts.T(), res5.Items[1].Score, float32(2.9))
	assert.Equal(ts.T(), "ping pong pung", res5.Items[1].Segment)
	assert.Equal(ts.T(), 3, len(res5.Items[1].MatchedKeywordsList))
	assert.Equal(ts.T(), "ping", res5.Items[1].MatchedKeywordsList[0])
	assert.Equal(ts.T(), "pong", res5.Items[1].MatchedKeywordsList[1])
	assert.Equal(ts.T(), "pung", res5.Items[1].MatchedKeywordsList[2])

	assert.Greater(ts.T(), res5.Items[2].Score, float32(1.9))
	assert.Equal(ts.T(), "pong pung", res5.Items[2].Segment)
	assert.Equal(ts.T(), 2, len(res5.Items[2].MatchedKeywordsList))
	assert.Equal(ts.T(), "pong", res5.Items[2].MatchedKeywordsList[0])
	assert.Equal(ts.T(), "pung", res5.Items[2].MatchedKeywordsList[1])

	assert.Greater(ts.T(), res5.Items[3].Score, float32(0.9))
	assert.Equal(ts.T(), "ha no no я Français", res5.Items[3].Segment)
	assert.Equal(ts.T(), 1, len(res5.Items[3].MatchedKeywordsList))
	assert.Equal(ts.T(), "ha", res5.Items[3].MatchedKeywordsList[0])

	assert.Greater(ts.T(), res5.Items[4].Score, float32(0.9))
	assert.Equal(ts.T(), "万事如意 ha", res5.Items[4].Segment)
	assert.Equal(ts.T(), 1, len(res5.Items[4].MatchedKeywordsList))
	assert.Equal(ts.T(), "ha", res5.Items[4].MatchedKeywordsList[0])
}

func (ts *pgTestSuite) TestConstraints() {
	mtx := ts.db.NewModelTx(context.Background())

	bas, err := json.Marshal([]map[string]any{})
	assert.Nil(ts.T(), err)

	_, err = mtx.CreateFormat(persistence.Format{ID: "pdf", Basis: bas})
	assert.Nil(ts.T(), err)
	_, err = mtx.CreateIndex(persistence.Index{ID: "abc.txt", Format: "pdf", Tags: persistence.Tags{"key": "val"}})
	assert.Nil(ts.T(), err)

	err = mtx.DeleteFormat("pdf")
	assert.ErrorIs(ts.T(), err, errors.ErrConflict)
}
