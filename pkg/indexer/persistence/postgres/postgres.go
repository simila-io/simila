package postgres

import (
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"github.com/jmoiron/sqlx"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/pgroonga"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres/pgtrgm"
)

const (
	SearchModuleNone    = ""
	SearchModuleGroonga = "pgroonga"
	SearchModuleTrgm    = "pgtrgm"
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
	if err = rollbackNotCurrent(ctx, db, search); err != nil {
		return nil, fmt.Errorf("rollback failed: %w", err)
	}

	switch search {
	case SearchModuleNone:
		return getPgNonSpecificDb(ctx, db)
	case SearchModuleGroonga:
		return getPgGroongaDb(ctx, db)
	case SearchModuleTrgm:
		return getPgTrgmDb(ctx, db)
	}
	return nil, fmt.Errorf("unsupported postgres search module=%s: %w", search, errors.ErrInvalid)
}

func getPgNonSpecificDb(ctx context.Context, db *sqlx.DB) (*Db, error) {
	if err := migrateUpShared(ctx, db.DB); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return newDb(db, nil), nil
}

func getPgGroongaDb(ctx context.Context, db *sqlx.DB) (*Db, error) {
	if err := migrateUpSharedAndPgGroonga(ctx, db.DB); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	if err := setSessionParams(ctx, db, pgroonga.SessionParams()); err != nil {
		return nil, fmt.Errorf("session params set failed: %w", err)
	}
	return newDb(db, pgroonga.SearchFn), nil
}

func getPgTrgmDb(ctx context.Context, db *sqlx.DB) (*Db, error) {
	if err := migrateUpSharedAndPgTrgm(ctx, db.DB); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	if err := setSessionParams(ctx, db, pgtrgm.SessionParams()); err != nil {
		return nil, fmt.Errorf("session params set failed: %w", err)
	}
	return newDb(db, pgtrgm.SearchFn), nil
}

// rollbackNotCurrent rolls back migrations that DO NOT belong
// to curr, since the greatest migration found in DB must either
// belong to curr module migrations or be the "shared" one.
func rollbackNotCurrent(ctx context.Context, db *sqlx.DB, curr SearchModuleName) error {
	var err error
	switch curr {
	case SearchModuleGroonga:
		err = migrateDownPgTrgm(ctx, db.DB)
	case SearchModuleTrgm:
		err = migrateDownPgGroonga(ctx, db.DB)
	}
	return err
}

func setSessionParams(ctx context.Context, db *sqlx.DB, sessParams map[string]any) error {
	for k, v := range sessParams {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("set %s = %v ", k, v)); err != nil {
			return err
		}
	}
	return nil
}