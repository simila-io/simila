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

package migrations

import migrate "github.com/rubenv/sql-migrate"

const (
	initUp = `
create extension pgroonga;

create table if not exists "format"
(
    "id"          varchar(64)              not null,
	"name"     	  varchar(255)             not null,
    "basis"       jsonb                    not null default '{}'::jsonb,
    "created_at"  timestamp with time zone not null default (now() at time zone 'utc'),
    "updated_at"  timestamp with time zone not null default (now() at time zone 'utc'),
    primary key ("id")
);

create unique index if not exists "idx_format_name" ON "format" ("name");
create index if not exists "idx_format_basis" on "format" using gin ("basis");

create table if not exists "index"
(
    "id"         varchar(255)             not null,
    "format"  	 varchar(255)             not null references "format" ("name") on delete restrict,
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

create index if not exists "idx_index_record_segment" on "index_record" using pgroonga ("segment");
create index if not exists "idx_index_record_vector" on "index_record" using gin ("vector");
create index if not exists "idx_index_record_created_at" on "index_record" ("created_at");
`

	initDown = `
drop table if exists "index_record";
drop table if exists "index";
drop table if exists "format";

drop extension pgroonga;
`
)

func InitTable(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:   id,
		Up:   []string{initUp},
		Down: []string{initDown},
	}
}
