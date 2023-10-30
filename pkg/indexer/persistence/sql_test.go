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
)

type pureSqlTestSuite struct {
	SqlTestSuite
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(pureSqlTestSuite))
}

func (s *pureSqlTestSuite) TestCreateIndexRecord() {
	mtx := s.db.NewModelTx()
	frmt := Format{Name: "pdf", Basis: StrStrMap{"d1": "1", "d2": "2"}}
	frmtID, err := mtx.CreateFormat(frmt)
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", frmtID)

	idx := Index{ID: "abc.txt", Format: frmt.Name, Tags: StrStrMap{"k": "v"}}
	idxID, err := mtx.CreateIndex(idx)
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", idxID)

	rec := IndexRecord{IndexID: idxID, Segment: "haha", Vector: StrStrMap{"x1": "1", "x2": "2"}}
	recID, err := mtx.CreateIndexRecord(rec)
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), "", recID)
}
