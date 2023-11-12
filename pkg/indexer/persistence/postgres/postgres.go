package postgres

import (
	"fmt"
	"github.com/logrange/linker"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/groonga"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/trigram"
)

const (
	DefaultMode = "pg_default"
	GroongaMode = "pg_groonga"
	TrigramMode = "pg_trigram"
)

type (
	SearchMode string

	LifecycleDb interface {
		persistence.Db
		linker.Initializer
		linker.Shutdowner
	}
)

func GetDb(dsName string, searchMode SearchMode) LifecycleDb {
	switch ext := searchMode; ext {
	case "", DefaultMode:
		fallthrough // TODO: implement Postgres built-in full-text search
	case GroongaMode:
		return groonga.NewDb(dsName)
	case TrigramMode:
		return trigram.NewDb(dsName)
	default:
		panic(fmt.Sprintf("unsupported postgres search mode: %s", ext))
	}
}
