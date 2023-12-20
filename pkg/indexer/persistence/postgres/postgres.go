package postgres

import (
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"github.com/jmoiron/sqlx"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/fts"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/groonga"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/trigram"
)

const (
	SearchModuleNone    = ""
	SearchModuleGroonga = "pgroonga"
	SearchModuleTrigram = "pgtrigram"
	SearchModuleFts     = "pgfts"
)

type SearchModuleName string

// MustGetDb does the same as GetDb but panics in case of an error
func MustGetDb(ctx context.Context, dsName string, search SearchModuleName) *Db {
	db, err := GetDb(ctx, dsName, search)
	if err != nil {
		panic(err)
	}
	return db
}

// GetDb returns the Db object built for the given configuration
func GetDb(ctx context.Context, dsName string, search SearchModuleName) (*Db, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", dsName)
	if err != nil {
		return nil, fmt.Errorf("could not connect to the database: %w", err)
	}
	if err = rollbackOthers(ctx, db.DB, search); err != nil {
		return nil, fmt.Errorf("rollback failed: %w", err)
	}
	switch search {
	case SearchModuleNone:
		return getDefaultDb(ctx, db)
	case SearchModuleGroonga:
		return getGroongaDb(ctx, db)
	case SearchModuleTrigram:
		return getTrigramDb(ctx, db)
	case SearchModuleFts:
		return getFtsDb(ctx, db)
	}
	return nil, fmt.Errorf("unsupported postgres search module=%s: %w", search, errors.ErrInvalid)
}

func getDefaultDb(ctx context.Context, db *sqlx.DB) (*Db, error) {
	if err := migrateCommonUp(ctx, db.DB); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return newDb(db, dbExt{}), nil
}

func getGroongaDb(ctx context.Context, db *sqlx.DB) (*Db, error) {
	if err := migrateGroongaUp(ctx, db.DB); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return newDb(db, dbExt{tr: groonga.FcTranslator, searchFn: groonga.Search}), nil
}

func getTrigramDb(ctx context.Context, db *sqlx.DB) (*Db, error) {
	if err := migrateTrigramUp(ctx, db.DB); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	if err := setSessionParams(ctx, db, trigram.SessionParams()); err != nil {
		return nil, fmt.Errorf("session params set failed: %w", err)
	}
	return newDb(db, dbExt{tr: trigram.FcTranslator, searchFn: trigram.Search}), nil
}

func getFtsDb(ctx context.Context, db *sqlx.DB) (*Db, error) {
	if err := migrateFtsUp(ctx, db.DB); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return newDb(db, dbExt{tr: fts.FcTranslator, searchFn: fts.Search}), nil
}

func setSessionParams(ctx context.Context, db *sqlx.DB, sessParams map[string]any) error {
	for k, v := range sessParams {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("set %s = %v ", k, v)); err != nil {
			return err
		}
	}
	return nil
}
