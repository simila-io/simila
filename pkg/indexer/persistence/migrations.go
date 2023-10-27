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
	"database/sql"
	_ "database/sql"
	"github.com/acquirecloud/golibs/logging"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/simila-io/simila/pkg/indexer/persistence/migrations"
)

type migration struct {
	driverName     string
	dataSourceName string
	logger         logging.Logger
}

func NewMigration(driverName, dataSourceName string) *migration {
	return &migration{
		driverName:     driverName,
		dataSourceName: dataSourceName,
		logger:         logging.NewLogger("migration"),
	}
}

func (m *migration) Init(ctx context.Context) error {
	m.logger.Infof("starting...")
	m.logger.Debugf("connection to %s %s", m.driverName, m.dataSourceName)
	db, err := sqlx.Connect(m.driverName, m.dataSourceName)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = migrateUp(db.DB); err != nil {
		m.logger.Errorf("migration failed: %s", err.Error())
	}
	return err
}

func migrateUp(db *sql.DB) error {
	migrs := make([]*migrate.Migration, 0)
	migrs = append(migrs, migrations.InitTable("1"))
	migrs = append(migrs, dummy("2"))
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	if _, err := migrate.Exec(db, "postgres", mms, migrate.Up); err != nil {
		return err
	}
	return nil
}

func migrateDown(db *sql.DB) error {
	migrs := make([]*migrate.Migration, 0)
	migrs = append(migrs, migrations.InitTable("1"))
	migrs = append(migrs, dummy("2"))
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	if _, err := migrate.Exec(db, "postgres", mms, migrate.Down); err != nil {
		return err
	}
	return nil
}

func dummy(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:   id,
		Up:   []string{},
		Down: []string{},
	}
}
