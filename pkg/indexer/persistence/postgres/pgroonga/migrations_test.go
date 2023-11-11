package pgroonga

import (
	"context"
	"github.com/stretchr/testify/assert"
)

func (ts *pgTestSuite) TestMigrations() {
	assert.NoError(ts.T(), migrateUp(context.Background(), ts.db.db.DB))
	assert.NoError(ts.T(), migrateDown(context.Background(), ts.db.db.DB))
}
