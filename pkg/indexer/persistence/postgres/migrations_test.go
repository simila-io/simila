package postgres

import (
	"context"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/stretchr/testify/assert"
	"time"
)

func (ts *pgSharedTestSuite) TestMigrations() {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	// down
	assert.NoError(ts.T(), migrateDownShared(ctx, ts.db.db.DB))
	count, err := persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)

	// up
	assert.NoError(ts.T(), migrateUpShared(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(2), count)

	// down
	assert.NoError(ts.T(), migrateDownShared(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)
}

func (ts *pgGroongaTestSuite) TestMigrations() {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	// down
	assert.NoError(ts.T(), migrateDownSharedAndPgGroonga(ctx, ts.db.db.DB))
	count, err := persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)

	// up
	assert.NoError(ts.T(), migrateUpSharedAndPgGroonga(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(4), count)

	// down
	assert.NoError(ts.T(), migrateDownSharedAndPgGroonga(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)
}

func (ts *pgTrgmTestSuite) TestMigrations() {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	// down
	assert.NoError(ts.T(), migrateDownSharedAndPgTrgm(ctx, ts.db.db.DB))
	count, err := persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)

	// up
	assert.NoError(ts.T(), migrateUpSharedAndPgTrgm(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(4), count)

	// down
	assert.NoError(ts.T(), migrateDownSharedAndPgTrgm(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)
}
