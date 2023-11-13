package postgres

import (
	"context"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/stretchr/testify/assert"
	"time"
)

// common

func (ts *pgCommonTestSuite) TestMigrations() {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	// down
	assert.NoError(ts.T(), migrateCommonDown(ctx, ts.db.db.DB))
	count, err := persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)

	// up
	assert.NoError(ts.T(), migrateCommonUp(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(2), count)

	// down
	assert.NoError(ts.T(), migrateCommonDown(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)
}

// groonga

func (ts *pgGroongaTestSuite) TestMigrations() {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	// down
	assert.NoError(ts.T(), migrateGroongaDown(ctx, ts.db.db.DB))
	count, err := persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)

	// up
	assert.NoError(ts.T(), migrateGroongaUp(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(4), count)

	// down
	assert.NoError(ts.T(), migrateGroongaDown(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)
}

//  trigram

func (ts *pgTrigramTestSuite) TestMigrations() {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	// down
	assert.NoError(ts.T(), migrateTrigramDown(ctx, ts.db.db.DB))
	count, err := persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)

	// up
	assert.NoError(ts.T(), migrateTrigramUp(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(4), count)

	// down
	assert.NoError(ts.T(), migrateTrigramDown(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)
}

// fts

func (ts *pgFtsTestSuite) TestMigrations() {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	// down
	assert.NoError(ts.T(), migrateFtsDown(ctx, ts.db.db.DB))
	count, err := persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)

	// up
	assert.NoError(ts.T(), migrateFtsUp(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(3), count)

	// down
	assert.NoError(ts.T(), migrateFtsDown(ctx, ts.db.db.DB))
	count, err = persistence.Count(ctx, ts.db.db, "select count(*) from gorp_migrations")
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)
}
