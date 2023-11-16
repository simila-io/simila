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
	var err error
	var dbCont persistence.DbContainer

	switch ts.sModule {
	case SearchModuleNone, SearchModuleGroonga, SearchModuleTrigram, SearchModuleFts:
		ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
		defer cancelFn()
		dbCont, err = persistence.NewPgDbContainer(ctx, "simila/similadb:latest", persistence.WithDbName("simila_test"))
		assert.Nil(ts.T(), err)
	default:
		err = fmt.Errorf("unsupported postgres search module: %s", ts.sModule)
	}

	// For non-container DB use NewNilDbContainer:
	//
	//dbCont, err = persistence.NewNilDbContainer(
	//	persistence.WithHost("127.0.0.1"),
	//	persistence.WithUser("postgres"),
	//	persistence.WithPassword("postgres"),
	//	persistence.WithPort("5432"),
	//	persistence.WithDbName("simila_test"),
	//	persistence.WithSslMode("disable"))

	assert.Nil(ts.T(), err)
	ts.dbCont = dbCont
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
