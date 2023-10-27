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
CREATE TABLE IF NOT EXISTS "testtable"(
    "id" varchar(32) NOT NULL,
    PRIMARY KEY("id")
);
`

	initDown = `
DROP TABLE IF EXISTS "testtable";
`
)

func InitTable(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:   id,
		Up:   []string{initUp},
		Down: []string{initDown},
	}
}
