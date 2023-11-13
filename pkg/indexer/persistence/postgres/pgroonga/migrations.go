package pgroonga

import migrate "github.com/rubenv/sql-migrate"

const (
	createExtensionUp = `
create extension if not exists pgroonga;
`
	createSegmentIndexUp = `
create index if not exists "idx_index_record_segment_groonga" on "index_record" using pgroonga ("segment") with (tokenizer='TokenNgram("unify_alphabet", false, "unify_symbol", false, "unify_digit", false)');
`
	createSegmentIndexDown = `
drop index if exists "idx_index_record_segment_groonga";
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
// the "shared" migrations for the "pgroonga" search module to work,
// the "pgroonga" migrations IDs range is [1xxxxx ... 19xxxx]
func Migrations() []*migrate.Migration {
	return []*migrate.Migration{
		createExtension("100001"),
		createSegmentIndex("100002"),
	}
}
