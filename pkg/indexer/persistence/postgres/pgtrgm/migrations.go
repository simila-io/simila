package pgtrgm

import migrate "github.com/rubenv/sql-migrate"

const (
	createExtensionUp = `
create extension if not exists pg_trgm;
`
	createSegmentIndexUp = `
create index if not exists "idx_index_record_segment_trgm" on "index_record" using gin ("segment" gin_trgm_ops);
`
	createSegmentIndexDown = `
drop index if exists "idx_index_record_segment_trgm";
`
)

func createExtension(id string) *migrate.Migration {
	return &migrate.Migration{
		Id: id,
		Up: []string{createExtensionUp},
	}
}

func createSegmentIndex(id string) *migrate.Migration {
	return &migrate.Migration{
		Id:   id,
		Up:   []string{createSegmentIndexUp},
		Down: []string{createSegmentIndexDown},
	}
}

// Migrations returns migrations to be applied on top of
// the "shared" migrations for the "pgtrgm" search module to work,
// the "pgtrgm" migrations IDs range is [2xxxxx ... 29xxxx]
func Migrations() []*migrate.Migration {
	return []*migrate.Migration{
		createExtension("200001"),
		createSegmentIndex("200002"),
	}
}
