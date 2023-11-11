package postgres

import (
	"fmt"
	"github.com/logrange/linker"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/pgroonga"
)

const (
	GroongaExt = "pgroonga"
	TrigramExt = "pgtrgm"
)

type (
	LifecycleDb interface {
		persistence.Db
		linker.Initializer
		linker.Shutdowner
	}
)

func GetDb(dsName, searchExt string) LifecycleDb {
	switch ext := searchExt; ext {
	case GroongaExt:
		return pgroonga.NewDb(dsName)
	default:
		panic(fmt.Sprintf("unsupported postgres extension: %s", ext))
	}
}
