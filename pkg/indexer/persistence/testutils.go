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
	"fmt"
	"github.com/docker/docker/api/types/container"
	"io"
	"time"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/simila-io/simila/pkg/indexer/persistence/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type (
	// SqlTestSuite struct used to wrap all related database connection stuff into one suite
	SqlTestSuite struct {
		Dir string
		suite.Suite
		dsp dsProvider
		db  *db
	}

	dsProvider interface {
		io.Closer
		datasource() string
	}

	containerDatabase struct {
		container testcontainers.Container
		url       string
	}

	constDatabase struct {
		url string
	}
)

var _ dsProvider = (*containerDatabase)(nil)
var _ dsProvider = (*constDatabase)(nil)

func newConstDatabase(ds string) (dsProvider, error) {
	return &constDatabase{url: ds}, nil
}

// newContainerDatabase creates a new test database via docker container with PostgreSQL
// and returns a TestDatabase struct.
func newContainerDatabase() (dsProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	req := testcontainers.ContainerRequest{
		Image:        "groonga/pgroonga:latest-debian-16",
		ExposedPorts: []string{"5432/tcp"},
		HostConfigModifier: func(config *container.HostConfig) {
			config.AutoRemove = true
		},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "simila_test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	return &containerDatabase{
		container: container,
		url:       fmt.Sprintf("host=127.0.0.1 port=%d user=postgres password=postgres sslmode=disable", port.Int()),
	}, nil
}

func (testDB *containerDatabase) datasource() string {
	return testDB.url
}

// Close closes the container database.
func (testDB *containerDatabase) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return testDB.container.Terminate(ctx)
}

// Close implements io.Closer
func (c constDatabase) Close() error {
	return nil
}

func (c *constDatabase) datasource() string {
	return c.url
}

// getTestDSProvider returns dtatasource provider
func getTestDSProvider() (dsProvider, error) {
	// returns container based postgres
	return newContainerDatabase()
	// returns postgres running locally
	return newConstDatabase("host=0.0.0.0 port=5432 user=postgres password=postgres sslmode=disable")
}

func (s *SqlTestSuite) SetupSuite() {
	var err error
	s.dsp, err = getTestDSProvider()
	assert.Nil(s.T(), err)
	s.T().Log("Test database setup successfully")
}

func (s *SqlTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Shutdown()
	}
	if s.dsp != nil {
		assert.Nil(s.T(), s.dsp.Close())
	}
}

func (s *SqlTestSuite) BeforeTest(suiteName, testName string) {
	assert.Nil(s.T(), dropCreatePgDb(context.Background(), s.dsp.datasource(), "simila_test"))
	s.db = NewDb("postgres", s.dsp.datasource()+" dbname=simila_test").(*db)
	assert.Nil(s.T(), s.db.Init(context.Background()))

	mtx := s.db.NewTx()
	mtx.MustBegin()
	defer mtx.Rollback()
	s.setupDb()
	mtx.Commit()
}

func (s *SqlTestSuite) AfterTest(suiteName, testName string) {
	s.cleanupDb()
	if s.db != nil {
		s.db.Shutdown()
	}
}

func (s *SqlTestSuite) setupDb() {
	migrs := make([]*migrate.Migration, 0)
	migrs = append(migrs, migrations.InitTable("1"))
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	_, err := migrate.Exec(s.db.db.DB, "postgres", mms, migrate.Up)
	assert.Nil(s.T(), err)
}

func (s *SqlTestSuite) cleanupDb() {
	migrs := make([]*migrate.Migration, 0)
	migrs = append(migrs, migrations.InitTable("1"))
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	_, err := migrate.Exec(s.db.db.DB, "postgres", mms, migrate.Down)
	assert.Nil(s.T(), err)
}

func (s *SqlTestSuite) GetDb() Db {
	return s.db
}

func dropCreatePgDb(ctx context.Context, ds string, dbName string) error {
	db, err := sqlx.ConnectContext(ctx, "postgres", ds)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err = db.DB.Exec(fmt.Sprintf("drop database if exists %s with (force)", dbName)); err != nil {
		return err
	}
	if _, err = db.DB.Exec(fmt.Sprintf("create database %s", dbName)); err != nil {
		return err
	}
	return nil
}
