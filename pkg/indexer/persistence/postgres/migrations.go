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
	"database/sql"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/fts"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/groonga"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/trigram"
)

const (
	initSchemaUp = `
create table if not exists "format"
(
    "id"          varchar(255)             not null,
    "basis"       jsonb                    not null default '{}'::jsonb,
    "created_at"  timestamp with time zone not null default (now() at time zone 'utc'),
    "updated_at"  timestamp with time zone not null default (now() at time zone 'utc'),
    primary key ("id")
);

create index if not exists "idx_format_basis" on "format" using gin ("basis");

create table if not exists "index"
(
    "id"         varchar(255)             not null,
    "format"  	 varchar(255)             not null references "format" ("id") on delete restrict,
    "tags"       jsonb                    not null default '{}'::jsonb,
    "created_at" timestamp with time zone not null default (now() at time zone 'utc'),
    "updated_at" timestamp with time zone not null default (now() at time zone 'utc'),
    primary key ("id")
);

create index if not exists "idx_index_tags" on "index" using gin ("tags");
create index if not exists "idx_index_created_at" on "index" ("created_at");

create table if not exists "index_record"
(
    "id"          varchar(64)   not null,
    "index_id"    varchar(255)  not null references "index" ("id") on delete cascade,
    "segment"     text         	not null,
    "vector"   	  jsonb        	not null default '{}'::jsonb,
    "created_at"  timestamp with time zone not null default (now() at time zone 'utc'),
    "updated_at"  timestamp with time zone not null default (now() at time zone 'utc'),
    primary key ("index_id", "id")
);

create index if not exists "idx_index_record_vector" on "index_record" using gin ("vector");
create index if not exists "idx_index_record_created_at" on "index_record" ("created_at");
`
	initSchemaDown = `
drop table if exists "index_record";
drop table if exists "index";
drop table if exists "format";
`

	addTxtFormatUp = `
insert into format (id) values('txt') on conflict do nothing;
`
	addTxtFormatDown = `
delete from format where id='txt';
`
)

func initSchema(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:   id,
		Up:   []string{initSchemaUp},
		Down: []string{initSchemaDown},
	}
}

func addTxtFormat(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:   id,
		Up:   []string{addTxtFormatUp},
		Down: []string{addTxtFormatDown},
	}
}

// migrations returns migrations to be reused for
// all the specific search implementations, the range of
// "common" migrations IDs [0-999]
func migrations() []*migrate.Migration {
	return []*migrate.Migration{
		initSchema("0"),
		addTxtFormat("1"),
	}
}

func migrateCommonUp(ctx context.Context, db *sql.DB) error {
	mms := migrate.MemoryMigrationSource{Migrations: migrations()}
	if _, err := migrate.ExecContext(ctx, db, "postgres", mms, migrate.Up); err != nil {
		return err
	}
	return nil
}

func migrateCommonDown(ctx context.Context, db *sql.DB) error {
	mms := migrate.MemoryMigrationSource{Migrations: migrations()}
	if _, err := migrate.ExecContext(ctx, db, "postgres", mms, migrate.Down); err != nil {
		return err
	}
	return nil
}

// groonga

func migrateGroongaUp(ctx context.Context, db *sql.DB) error {
	var migrs []*migrate.Migration
	migrs = append(migrs, migrations()...)
	migrs = append(migrs, groonga.Migrations()...)
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	if _, err := migrate.ExecContext(ctx, db, "postgres", mms, migrate.Up); err != nil {
		return err
	}
	return nil
}

func migrateGroongaDown(ctx context.Context, db *sql.DB) error {
	var migrs []*migrate.Migration
	migrs = append(migrs, migrations()...)
	migrs = append(migrs, groonga.Migrations()...)
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	if _, err := migrate.ExecContext(ctx, db, "postgres", mms, migrate.Down); err != nil {
		return err
	}
	return nil
}

// trigram

func migrateTrigramUp(ctx context.Context, db *sql.DB) error {
	var migrs []*migrate.Migration
	migrs = append(migrs, migrations()...)
	migrs = append(migrs, trigram.Migrations()...)
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	if _, err := migrate.ExecContext(ctx, db, "postgres", mms, migrate.Up); err != nil {
		return err
	}
	return nil

}

func migrateTrigramDown(ctx context.Context, db *sql.DB) error {
	var migrs []*migrate.Migration
	migrs = append(migrs, migrations()...)
	migrs = append(migrs, trigram.Migrations()...)
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	if _, err := migrate.ExecContext(ctx, db, "postgres", mms, migrate.Down); err != nil {
		return err
	}
	return nil
}

// fts

func migrateFtsUp(ctx context.Context, db *sql.DB) error {
	var migrs []*migrate.Migration
	migrs = append(migrs, migrations()...)
	migrs = append(migrs, fts.Migrations()...)
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	if _, err := migrate.ExecContext(ctx, db, "postgres", mms, migrate.Up); err != nil {
		return err
	}
	return nil

}

func migrateFtsDown(ctx context.Context, db *sql.DB) error {
	var migrs []*migrate.Migration
	migrs = append(migrs, migrations()...)
	migrs = append(migrs, fts.Migrations()...)
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	if _, err := migrate.ExecContext(ctx, db, "postgres", mms, migrate.Down); err != nil {
		return err
	}
	return nil
}

// rollback NOT currSearch module migrations,
// this will leave only "common" migrations and
// will allow switching between search modules,
// if we set migrate.SetIgnoreUnknown(true)

func rollbackOthers(ctx context.Context, db *sql.DB, currSearch SearchModuleName) error {
	migrate.SetIgnoreUnknown(true)
	var migrs []*migrate.Migration
	if currSearch != SearchModuleFts {
		migrs = append(migrs, fts.Migrations()...)
	}
	if currSearch != SearchModuleGroonga {
		migrs = append(migrs, groonga.Migrations()...)
	}
	if currSearch != SearchModuleTrigram {
		migrs = append(migrs, trigram.Migrations()...)
	}
	mms := migrate.MemoryMigrationSource{
		Migrations: migrs,
	}
	if _, err := migrate.ExecContext(ctx, db, "postgres", mms, migrate.Down); err != nil {
		return err
	}
	migrate.SetIgnoreUnknown(false)
	return nil
}
