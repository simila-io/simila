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

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"github.com/acquirecloud/golibs/logging"
	"github.com/jmoiron/sqlx"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"os"
	"strings"
	"time"
)

type (
	// Db implements persistence.Db
	Db struct {
		logger   logging.Logger
		searchFn SearchFn
		db       *sqlx.DB
	}

	// SearchFn is used to provide different search implementations
	SearchFn func(ctx context.Context, qx sqlx.QueryerContext, q persistence.SearchQuery) (persistence.SearchQueryResult, error)

	// exec is a helper interface to provide joined functionality of sqlx.DB and sqlx.Tx
	// it is used by the tx.executor()
	exec interface {
		sqlx.QueryerContext
		sqlx.Ext
		GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}

	// tx implements the Tx interface
	tx struct {
		ctx context.Context // context for all the operations within the tx
		db  *sqlx.DB        // never nil
		tx  *sqlx.Tx        // keeps active transaction, if it exists. It can be nil, if not started.
	}

	// modelTx is a helper to persist persistence objects ModelTx
	modelTx struct {
		searchFn SearchFn
		*tx      // the active transaction, never nil for the object
	}
)

func newDb(sdb *sqlx.DB, searchFn SearchFn) *Db {
	return &Db{db: sdb, searchFn: searchFn, logger: logging.NewLogger("db.postgres")}
}

// Init implements linker.Initializer interface
func (d *Db) Init(ctx context.Context) error {
	d.logger.Infof("Initializing...")
	return nil
}

// Shutdown implements linker.Shutdowner interface
func (d *Db) Shutdown() {
	d.logger.Infof("Shutdown")
	if d.db == nil {
		d.logger.Errorf("not initialized, but shutting down")
		return
	}
	err := d.db.Close()
	if err != nil {
		d.logger.Warnf("could not close the DB connection: %v", err)
	}
}

// NewModelTx returns the new ModelTx object
func (d *Db) NewModelTx(ctx context.Context) persistence.ModelTx {
	return &modelTx{tx: d.NewTx(ctx).(*tx), searchFn: d.searchFn}
}

// NewTx returns the new Tx object
func (d *Db) NewTx(ctx context.Context) persistence.Tx {
	return &tx{ctx: ctx, db: d.db}
}

// ============================== tx ====================================

func (t *tx) executor() exec {
	if t.tx == nil {
		return t.db
	}
	return t.tx
}

// MustBegin is a part of the Tx interface
func (t *tx) MustBegin() {
	t.Commit()
	t.tx = t.db.MustBeginTx(t.ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})
}

// MustBeginSerializable is a part of the Tx interface
func (t *tx) MustBeginSerializable() {
	tx := t.db.MustBeginTx(t.ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	t.Commit()
	t.tx = tx
}

// Rollback rolls the transaction bock (if started)
func (t *tx) Rollback() error {
	var err error
	if t.tx != nil {
		err = t.tx.Rollback()
		t.tx = nil
	}
	return err
}

func (t *tx) Commit() error {
	var err error
	if t.tx != nil {
		err = t.tx.Commit()
		t.tx = nil
	}
	return err
}

// ExecQuery executes a query with params within the transaction
func (t *tx) execQuery(sqlQuery string, params ...interface{}) error {
	_, err := t.executor().ExecContext(t.ctx, sqlQuery, params...)
	return err
}

// ExecScript runs the sqlScript (file name)
func (t *tx) ExecScript(sqlScript string) error {
	file, err := os.ReadFile(sqlScript)
	if err != nil {
		return fmt.Errorf("could not read %s in ExecScript: %w", sqlScript, err)
	}

	requests := strings.Split(string(file), ";")
	for _, request := range requests {
		if strings.Trim(request, " ") == "" {
			continue
		}
		if err = t.execQuery(request); err != nil {
			return fmt.Errorf("could not execute %s in ExecScript: %w", request, err)
		}
	}
	return nil
}

// ============================== modelTx ====================================

func (m *modelTx) CreateFormat(format persistence.Format) (persistence.Format, error) {
	if len(format.ID) == 0 {
		return persistence.Format{}, fmt.Errorf("format ID must be non-empty: %w", errors.ErrInvalid)
	}
	if len(format.Basis) == 0 {
		format.Basis = []byte("{}")
	}
	format.CreatedAt = time.Now()
	format.UpdatedAt = format.CreatedAt
	_, err := m.executor().ExecContext(m.ctx, "insert into format (id, basis, created_at, updated_at) values ($1, $2, $3, $4)",
		format.ID, format.Basis, format.CreatedAt, format.UpdatedAt)
	if err != nil {
		return persistence.Format{}, persistence.MapError(err)
	}
	return format, nil
}

func (m *modelTx) GetFormat(ID string) (persistence.Format, error) {
	var f persistence.Format
	return f, persistence.MapError(m.executor().GetContext(m.ctx, &f, "select * from format where id=$1", ID))
}

func (m *modelTx) DeleteFormat(ID string) error {
	res, err := m.executor().ExecContext(m.ctx, "delete from format where id=$1", ID)
	if err != nil {
		return persistence.MapError(err)
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return errors.ErrNotExist
	}
	return nil
}

func (m *modelTx) ListFormats() ([]persistence.Format, error) {
	rows, err := m.executor().QueryxContext(m.ctx, "select * from format order by id")
	if err != nil {
		return nil, persistence.MapError(err)
	}
	defer rows.Close()
	return persistence.ScanRows[persistence.Format](rows)
}

func (m *modelTx) CreateNodes(nodes ...persistence.Node) ([]persistence.Node, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	var sb strings.Builder
	var params []any

	firstIdx := 1
	sb.WriteString("insert into node (path, name, tags, flags, created_at, updated_at) values ")
	now := time.Now()

	for i, n := range nodes {
		if len(n.Path) == 0 {
			return nil, fmt.Errorf("node path for item=%d must be specified: %w", i, errors.ErrInvalid)
		}
		if len(n.Name) == 0 {
			return nil, fmt.Errorf("node name for item=%d must be specified: %w", i, errors.ErrInvalid)
		}
		if n.Tags == nil {
			n.Tags = make(persistence.Tags)
		}
		if i > 0 {
			sb.WriteString(",")
		}

		sb.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", firstIdx, firstIdx+1, firstIdx+2, firstIdx+3, firstIdx+4, firstIdx+5))
		firstIdx += 6

		params = append(params, persistence.ToNodePath(n.Path))
		params = append(params, strings.TrimSpace(n.Name))
		params = append(params, n.Tags)
		params = append(params, n.Flags)
		params = append(params, now)
		params = append(params, now)
	}
	sb.WriteString(" returning *")
	rows, err := m.executor().QueryxContext(m.ctx, sb.String(), params...)
	if err != nil {
		return nil, persistence.MapError(err)
	}
	defer rows.Close()
	return persistence.ScanRows[persistence.Node](rows)
}

func (m *modelTx) ListNodes(path string) ([]persistence.Node, error) {
	var sb strings.Builder
	var args []any

	for _, p := range persistence.ToNodePathNamePairs(path) {
		if sb.Len() > 0 {
			sb.WriteString(" or ")
		}
		sb.WriteString("(path = ? and name = ?)")
		args = append(args, p[0], p[1])
	}
	if sb.Len() == 0 {
		return nil, nil
	}

	where := sqlx.Rebind(sqlx.DOLLAR, sb.String())
	rows, err := m.executor().QueryxContext(m.ctx, fmt.Sprintf("select * from node where %s order by path, name", where), args...)
	if err != nil {
		return nil, persistence.MapError(err)
	}
	defer rows.Close()
	return persistence.ScanRows[persistence.Node](rows)
}

func (m *modelTx) ListChildren(path string) ([]persistence.Node, error) {
	rows, err := m.executor().QueryxContext(m.ctx, "select * from node where path = $1 order by path, name", persistence.ToNodePath(path))
	if err != nil {
		return nil, persistence.MapError(err)
	}
	defer rows.Close()
	return persistence.ScanRows[persistence.Node](rows)
}

func (m *modelTx) GetNode(fqnp string) (persistence.Node, error) {
	path, name := persistence.ToNodePathName(fqnp)
	var node persistence.Node
	if err := m.executor().GetContext(m.ctx, &node, "select * from node where path=$1 and name = $2", path, name); err != nil {
		return persistence.Node{}, persistence.MapError(err)
	}
	return node, nil
}

func (m *modelTx) UpdateNode(node persistence.Node) error {
	if node.ID == 0 {
		return fmt.Errorf("node ID must be specified: %w", errors.ErrInvalid)
	}

	sb := strings.Builder{}
	sb.WriteString("update node set")

	var args []any
	if len(node.Tags) > 0 {
		sb.WriteString(" tags = ?")
		args = append(args, node.Tags.JSON())
	}
	if len(args) == 0 {
		return nil
	}

	sb.WriteString(", updated_at = ? where id = ?")
	args = append(args, time.Now(), node.ID)

	res, err := m.executor().ExecContext(m.ctx, sqlx.Rebind(sqlx.DOLLAR, sb.String()), args...)
	if err != nil {
		return persistence.MapError(err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.ErrNotExist
	}
	return nil
}

func (m *modelTx) DeleteNode(nID int64, force bool) error {
	if !force {
		childCount, err := persistence.Count(m.ctx, m.executor(),
			"select count (*) "+
				"from node, (select * from node where id = $1) as n "+
				"where n.flags = $2 and node.path like concat(n.path, n.name, '/', '%%')",
			nID, persistence.NodeFlagFolder)
		if err != nil {
			return persistence.MapError(err)
		}
		if childCount > 0 {
			return fmt.Errorf("delete node with ID=%d failed (force=%t), "+
				"the node has children: %w", nID, force, errors.ErrConflict)
		}
	}
	res, err := m.executor().ExecContext(m.ctx,
		"delete from node "+
			"using (select * from node where id = $1) as n "+
			"where node.id = n.id or (n.flags = $2 and node.path like concat(n.path, n.name, '/', '%%'))",
		nID, persistence.NodeFlagFolder)
	if err != nil {
		return persistence.MapError(err)
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return errors.ErrNotExist
	}
	return nil
}

func (m *modelTx) UpsertIndexRecords(records ...persistence.IndexRecord) (int64, error) {
	if len(records) == 0 {
		return 0, nil
	}

	var sb strings.Builder
	var params []any

	firstIdx := 1
	sb.WriteString("insert into index_record (id, node_id, segment, vector, format, rank_multiplier, created_at, updated_at) values ")
	now := time.Now()
	for i, r := range records {
		if len(r.ID) == 0 {
			return 0, fmt.Errorf("record ID for item=%d  must be specified: %w", i, errors.ErrInvalid)
		}
		if r.NodeID == 0 {
			return 0, fmt.Errorf("record node ID for item=%d must be specified: %w", i, errors.ErrInvalid)
		}
		if len(r.Format) == 0 {
			return 0, fmt.Errorf("record format for item=%d must be specified: %w", i, errors.ErrInvalid)
		}
		if r.RankMult <= 0 {
			r.RankMult = 1.0
		}
		if len(r.Vector) == 0 {
			r.Vector = []byte("{}")
		}
		if i > 0 {
			sb.WriteString(",")
		}

		sb.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", firstIdx, firstIdx+1, firstIdx+2, firstIdx+3, firstIdx+4, firstIdx+5, firstIdx+6, firstIdx+7))
		firstIdx += 8

		params = append(params, r.ID)
		params = append(params, r.NodeID)
		params = append(params, r.Segment)
		params = append(params, r.Vector)
		params = append(params, r.Format)
		params = append(params, r.RankMult)
		params = append(params, now)
		params = append(params, now)
	}
	sb.WriteString(" on conflict (node_id,id) " +
		"do update set (segment, vector, format, rank_multiplier, updated_at) = " +
		"(excluded.segment, excluded.vector, excluded.format, excluded.rank_multiplier, excluded.updated_at)")
	res, err := m.executor().ExecContext(m.ctx, sb.String(), params...)
	if err != nil {
		return 0, persistence.MapError(err)
	}
	cnt, _ := res.RowsAffected()
	return cnt, nil
}

func (m *modelTx) DeleteIndexRecords(records ...persistence.IndexRecord) (int64, error) {
	if len(records) == 0 {
		return 0, nil
	}

	var sb strings.Builder
	var args []any

	for _, r := range records {
		if sb.Len() > 0 {
			sb.WriteString(" or ")
		}
		sb.WriteString("(node_id = ? and id = ?)")
		args = append(args, r.NodeID, r.ID)
	}
	if sb.Len() == 0 {
		return 0, nil
	}

	where := sqlx.Rebind(sqlx.DOLLAR, sb.String())
	res, err := m.executor().ExecContext(m.ctx, fmt.Sprintf("delete from index_record where %s", where), args...)
	if err != nil {
		return 0, persistence.MapError(err)
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return 0, errors.ErrNotExist
	}
	return cnt, nil
}

func (m *modelTx) QueryIndexRecords(query persistence.IndexRecordQuery) (persistence.QueryResult[persistence.IndexRecord, string], error) {
	if query.NodeID == 0 {
		return persistence.QueryResult[persistence.IndexRecord, string]{}, fmt.Errorf("node ID must be specified: %w", errors.ErrInvalid)
	}

	sb := strings.Builder{}
	sb.WriteString(" node_id = ? ")

	args := make([]any, 0)
	args = append(args, query.NodeID)

	if len(query.FromID) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" id >= ? ")
		args = append(args, query.FromID)
	}
	if len(query.Format) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" format = ? ")
		args = append(args, query.Format)
	}
	if !query.CreatedBefore.IsZero() {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" created_at < ? ")
		args = append(args, query.CreatedBefore)
	}
	if !query.CreatedAfter.IsZero() {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" created_at > ? ")
		args = append(args, query.CreatedAfter)
	}

	var where string
	if sb.Len() > 0 {
		where = " where " + sqlx.Rebind(sqlx.DOLLAR, sb.String())
	}

	// count
	total, err := persistence.Count(m.ctx, m.executor(), fmt.Sprintf("select count(*) from index_record %s ", where), args...)
	if err != nil {
		return persistence.QueryResult[persistence.IndexRecord, string]{}, persistence.MapError(err)
	}

	// query
	if query.Limit <= 0 {
		return persistence.QueryResult[persistence.IndexRecord, string]{Total: total}, nil
	}
	args = append(args, query.Limit+1)
	rows, err := m.executor().QueryxContext(m.ctx, fmt.Sprintf("select * from index_record %s order by id limit $%d", where, len(args)), args...)
	if err != nil {
		return persistence.QueryResult[persistence.IndexRecord, string]{Total: total}, persistence.MapError(err)
	}

	// results
	res, err := persistence.ScanRowsQueryResult[persistence.IndexRecord](rows)
	if err != nil {
		return persistence.QueryResult[persistence.IndexRecord, string]{}, persistence.MapError(err)
	}
	var nextID string
	if len(res) > query.Limit {
		nextID = res[len(res)-1].ID
		res = res[:query.Limit]
	}
	return persistence.QueryResult[persistence.IndexRecord, string]{Items: res, NextID: nextID, Total: total}, nil
}

func (m *modelTx) Search(query persistence.SearchQuery) (persistence.SearchQueryResult, error) {
	if len(query.Query) == 0 {
		return persistence.SearchQueryResult{}, fmt.Errorf("search query must be non-empty: %w", errors.ErrInvalid)
	}
	if m.searchFn != nil {
		return m.searchFn(m.ctx, m.executor(), query)
	}
	return persistence.SearchQueryResult{}, errors.ErrUnimplemented
}
