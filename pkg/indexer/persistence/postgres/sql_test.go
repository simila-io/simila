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

package postgres

import (
	"context"
	"encoding/json"
	"github.com/acquirecloud/golibs/errors"
	_ "github.com/lib/pq"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type (
	pgCommonTestSuite struct {
		pgTestSuite
	}

	pgGroongaTestSuite struct {
		pgTestSuite
	}

	pgTrigramTestSuite struct {
		pgTestSuite
	}

	pgFtsTestSuite struct {
		pgTestSuite
	}
)

func TestRunCommonTestSuite(t *testing.T) {
	suite.Run(t, &pgCommonTestSuite{newPqTestSuite(SearchModuleNone)})
}

func TestRunGroongaTestSuite(t *testing.T) {
	suite.Run(t, &pgGroongaTestSuite{newPqTestSuite(SearchModuleGroonga)})
}

func TestRunTrgmTestSuite(t *testing.T) {
	suite.Run(t, &pgTrigramTestSuite{newPqTestSuite(SearchModuleTrigram)})
}

func TestRunFtsTestSuite(t *testing.T) {
	suite.Run(t, &pgFtsTestSuite{newPqTestSuite(SearchModuleFts)})
}

// common

func (ts *pgCommonTestSuite) TestFormat() {
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
