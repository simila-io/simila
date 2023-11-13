package postgres

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"time"
)

type (
	pgTestSuite struct {
		suite.Suite
		sModule SearchModuleName
		dbCont  persistence.DbContainer
		db      *Db
	}
)

func newPqTestSuite(pgExt SearchModuleName) pgTestSuite {
	return pgTestSuite{sModule: pgExt}
}

func (ts *pgTestSuite) SetupSuite() {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	// For localhost DB use NewNilDbContainer:
	// 	ts.dbCont, err = persistence.NewNilDbContainer(persistence.WithPort("5432"), persistence.WithDbName("simila_test"))
	// 	assert.Nil(ts.T(), err)

	var err error
	switch ts.sModule {
	case SearchModuleGroonga:
		ts.dbCont, err = persistence.NewPgDbContainer(ctx,
			"groonga/pgroonga:latest-debian-16", persistence.WithDbName("simila_test"))
		assert.Nil(ts.T(), err)
	case SearchModuleNone, SearchModuleTrigram:
		ts.dbCont, err = persistence.NewPgDbContainer(ctx,
			"postgres:16-alpine", persistence.WithDbName("simila_test"))
		assert.Nil(ts.T(), err)
	}
}

func (ts *pgTestSuite) TearDownSuite() {
	if ts.db != nil {
		ts.db.Shutdown()
	}
	if ts.dbCont != nil {
		_ = ts.dbCont.Close()
	}
}

func (ts *pgTestSuite) BeforeTest(suiteName, testName string) {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	dbCfg := ts.dbCont.DbConfig()
	assert.Nil(ts.T(), ts.dropCreatePgDb(ctx))

	var err error
	ts.db, err = GetDb(ctx, dbCfg.DataSourceFull(), ts.sModule)
	assert.Nil(ts.T(), err)
	assert.Nil(ts.T(), ts.db.Init(ctx))
}

func (ts *pgTestSuite) AfterTest(suiteName, testName string) {
	if ts.db != nil {
		ts.db.Shutdown()
		ts.db = nil
	}
}

func (ts *pgTestSuite) dropCreatePgDb(ctx context.Context) error {
	dbCfg := ts.dbCont.DbConfig()
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
