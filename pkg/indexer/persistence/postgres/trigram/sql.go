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

package trigram

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
	Db struct {
		dsName string // data source name "user=foo dbname=bar sslmode=disable"
		logger logging.Logger
		db     *sqlx.DB
	}

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
		*tx // the active transaction, never nil for the object
	}
)

// NewDb creates new db object
func NewDb(dsName string) *Db {
	return &Db{dsName: dsName, logger: logging.NewLogger("db.postgres")}
}

// Init implements linker.Initializer interface
func (d *Db) Init(ctx context.Context) error {
	d.logger.Infof("Initializing...")
	sdb, err := sqlx.ConnectContext(ctx, "postgres", d.dsName)
	if err != nil {
		return fmt.Errorf("could not connect to the database: %w", err)
	}
	if err = setSessionParams(ctx, sdb); err != nil {
		return fmt.Errorf("could not set database session params: %w", err)
	}
	d.logger.Infof("Migrating...")
	if err = migrateUp(ctx, sdb.DB); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	d.db = sdb
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
	return &modelTx{tx: d.NewTx(ctx).(*tx)}
}

// NewTx returns the new Tx object
func (d *Db) NewTx(ctx context.Context) persistence.Tx {
	return &tx{ctx: ctx, db: d.db}
}

// Better to set these parameters via migration for the whole DB or system,
// but unfortunately controlled environments like AWS won't allow to do that.
func setSessionParams(ctx context.Context, db *sqlx.DB) error {
	if _, err := db.ExecContext(ctx, "set pg_trgm.word_similarity_threshold = 0.3;"); err != nil {
		return fmt.Errorf("failed to set 'pg_trgm.word_similarity_threshold': %w", err)
	}
	return nil
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

func (m *modelTx) CreateIndex(index persistence.Index) (persistence.Index, error) {
	if len(index.ID) == 0 {
		return persistence.Index{}, fmt.Errorf("index ID must be specified: %w", errors.ErrInvalid)
	}
	if index.Tags == nil {
		index.Tags = make(persistence.Tags)
	}
	index.CreatedAt = time.Now()
	index.UpdatedAt = index.CreatedAt
	_, err := m.executor().ExecContext(m.ctx, "insert into index (id, format, tags, created_at, updated_at) values ($1, $2, $3, $4, $5)",
		index.ID, index.Format, index.Tags, index.CreatedAt, index.UpdatedAt)
	if err != nil {
		return persistence.Index{}, persistence.MapError(err)
	}
	return index, nil
}

func (m *modelTx) GetIndex(ID string) (persistence.Index, error) {
	var idx persistence.Index
	return idx, persistence.MapError(m.executor().GetContext(m.ctx, &idx, "select * from index where id=$1", ID))
}

func (m *modelTx) UpdateIndex(index persistence.Index) error {
	if len(index.ID) == 0 {
		return fmt.Errorf("index ID must be specified: %w", errors.ErrInvalid)
	}

	sb := strings.Builder{}
	sb.WriteString("update index set")

	args := make([]any, 0)
	if len(index.Tags) > 0 {
		sb.WriteString(" tags = ?")
		args = append(args, index.Tags)
	}
	if len(args) == 0 {
		return nil
	}

	sb.WriteString(", updated_at = ? where id = ?")
	args = append(args, time.Now(), index.ID)

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

func (m *modelTx) DeleteIndex(ID string) error {
	res, err := m.executor().ExecContext(m.ctx, "delete from index where id=$1", ID)
	if err != nil {
		return persistence.MapError(err)
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return errors.ErrNotExist
	}
	return nil
}

func (m *modelTx) QueryIndexes(query persistence.IndexQuery) (persistence.QueryResult[persistence.Index, string], error) {
	sb := strings.Builder{}
	args := make([]any, 0)

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
	if len(query.Tags) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		var tb strings.Builder
		tb.WriteString(" {")
		oldLen := tb.Len()
		for k, v := range query.Tags {
			if tb.Len() > oldLen {
				tb.WriteByte(',')
			}
			tb.WriteString(fmt.Sprintf("%q:%q", k, v))
		}
		tb.WriteString("}")
		sb.WriteString(" tags @> ?")
		args = append(args, tb.String())
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

	where := sqlx.Rebind(sqlx.DOLLAR, sb.String())
	if len(where) > 0 {
		where = " where " + where
	}

	// count
	total, err := m.getCount(fmt.Sprintf("select count(*) from index %s", where), args...)
	if err != nil {
		return persistence.QueryResult[persistence.Index, string]{}, persistence.MapError(err)
	}

	// query
	if query.Limit <= 0 {
		return persistence.QueryResult[persistence.Index, string]{Total: total}, nil
	}
	args = append(args, query.Limit+1)
	rows, err := m.executor().QueryxContext(m.ctx, fmt.Sprintf("select * from index %s order by id limit $%d", where, len(args)), args...)
	if err != nil {
		return persistence.QueryResult[persistence.Index, string]{Total: total}, persistence.MapError(err)
	}

	// results
	res, err := persistence.ScanRowsQueryResult[persistence.Index](rows)
	if err != nil {
		return persistence.QueryResult[persistence.Index, string]{}, persistence.MapError(err)
	}
	var nextID string
	if len(res) > query.Limit {
		nextID = res[len(res)-1].ID
		res = res[:query.Limit]
	}
	return persistence.QueryResult[persistence.Index, string]{Items: res, NextID: nextID, Total: total}, nil
}

func (m *modelTx) UpsertIndexRecords(records ...persistence.IndexRecord) error {
	if len(records) == 0 {
		return nil
	}

	var sb strings.Builder
	var params []any

	firstIdx := 1
	sb.WriteString("insert into index_record (id, index_id, segment, vector, created_at, updated_at) values ")
	now := time.Now()
	for i, r := range records {
		if len(r.ID) == 0 {
			return fmt.Errorf("record ID for item=%d  must be specified: %w", i, errors.ErrInvalid)
		}
		if len(r.IndexID) == 0 {
			return fmt.Errorf("record index ID for item=%d must be specified: %w", i, errors.ErrInvalid)
		}
		if len(r.Vector) == 0 {
			r.Vector = []byte("{}")
		}
		if i > 0 {
			sb.WriteString(",")
		}

		sb.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", firstIdx, firstIdx+1, firstIdx+2, firstIdx+3, firstIdx+4, firstIdx+5))
		firstIdx += 6

		params = append(params, r.ID)
		params = append(params, r.IndexID)
		params = append(params, r.Segment)
		params = append(params, r.Vector)
		params = append(params, now)
		params = append(params, now)
	}
	sb.WriteString(" on conflict (index_id,id) do update set (segment, vector, updated_at) = (excluded.segment, excluded.vector, excluded.updated_at)")
	if _, err := m.executor().ExecContext(m.ctx, sb.String(), params...); err != nil {
		return persistence.MapError(err)
	}
	return nil
}

func (m *modelTx) GetIndexRecord(ID, indexID string) (persistence.IndexRecord, error) {
	if len(ID) == 0 {
		return persistence.IndexRecord{}, fmt.Errorf("record ID must be specified: %w", errors.ErrInvalid)
	}
	if len(indexID) == 0 {
		return persistence.IndexRecord{}, fmt.Errorf("record index ID must be specified: %w", errors.ErrInvalid)
	}

	var r persistence.IndexRecord
	return r, persistence.MapError(m.executor().GetContext(m.ctx, &r, "select * FROM index_record WHERE index_id=$1 and id=$2", indexID, ID))
}

func (m *modelTx) UpdateIndexRecord(record persistence.IndexRecord) error {
	if len(record.ID) == 0 {
		return fmt.Errorf("record ID must be specified: %w", errors.ErrInvalid)
	}
	if len(record.IndexID) == 0 {
		return fmt.Errorf("record index ID must be specified: %w", errors.ErrInvalid)
	}

	sb := strings.Builder{}
	sb.WriteString("update index_record set")

	args := make([]interface{}, 0)
	if len(record.Segment) > 0 {
		if len(args) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(" segment = ?")
		args = append(args, record.Segment)
	}
	if len(record.Vector) > 0 {
		if len(args) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(" vector = ?")
		args = append(args, record.Vector)
	}
	if len(args) == 0 {
		return nil
	}

	sb.WriteString(", updated_at = ? where index_id = ? and id = ?")
	args = append(args, time.Now(), record.IndexID, record.ID)

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

func (m *modelTx) DeleteIndexRecords(records ...persistence.IndexRecord) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	recIDs := make([]string, len(records))
	idxIDs := make([]string, len(records))

	for i := 0; i < len(records); i++ {
		if len(records[i].ID) == 0 {
			return 0, fmt.Errorf("record ID for item=%d  must be specified: %w", i, errors.ErrInvalid)
		}
		if len(records[i].IndexID) == 0 {
			return 0, fmt.Errorf("record index ID for item=%d must be specified: %w", i, errors.ErrInvalid)
		}
		recIDs[i] = records[i].ID
		idxIDs[i] = records[i].IndexID
	}

	idsList, idsArgs, _ := sqlx.In("?", recIDs)
	idxIDsList, idxIDsArgs, _ := sqlx.In("?", idxIDs)

	qry := fmt.Sprintf("delete from index_record where index_id in (%s) and id in (%s)", idxIDsList, idsList)
	res, err := m.executor().ExecContext(m.ctx, sqlx.Rebind(sqlx.DOLLAR, qry), append(idxIDsArgs, idsArgs...)...)
	if err != nil {
		return 0, persistence.MapError(err)
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return 0, errors.ErrNotExist
	}
	return int(cnt), nil
}

func (m *modelTx) QueryIndexRecords(query persistence.IndexRecordQuery) (persistence.QueryResult[persistence.IndexRecord, string], error) {
	sb := strings.Builder{}
	args := make([]any, 0)

	if len(query.FromID) > 0 {
		var fromID persistence.IndexRecordID
		if err := fromID.Decode(query.FromID); err != nil {
			return persistence.QueryResult[persistence.IndexRecord, string]{}, fmt.Errorf("invalid FromID: %w", errors.ErrInvalid)
		}
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" index_record.index_id >= ? and index_record.id >= ? ")
		args = append(args, fromID.IndexID, fromID.RecordID)
	}
	if len(query.IndexIDs) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		oldLen := len(args)
		sb.WriteString(" index_id in ( ")
		for _, id := range query.IndexIDs {
			if len(args) > oldLen {
				sb.WriteString(", ")
			}
			sb.WriteString("?")
			args = append(args, id)
		}
		sb.WriteString(")")
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

	where := sqlx.Rebind(sqlx.DOLLAR, sb.String())
	if len(where) > 0 {
		where = " where " + where
	}

	// count
	total, err := m.getCount(fmt.Sprintf("select count(*) from index_record %s ", where), args...)
	if err != nil {
		return persistence.QueryResult[persistence.IndexRecord, string]{}, persistence.MapError(err)
	}

	// query
	if query.Limit <= 0 {
		return persistence.QueryResult[persistence.IndexRecord, string]{Total: total}, nil
	}
	args = append(args, query.Limit+1)
	rows, err := m.executor().QueryxContext(m.ctx, fmt.Sprintf("select * from index_record %s order by index_id asc, id asc limit $%d", where, len(args)), args...)
	if err != nil {
		return persistence.QueryResult[persistence.IndexRecord, string]{Total: total}, persistence.MapError(err)
	}

	// results
	res, err := persistence.ScanRowsQueryResult[persistence.IndexRecord](rows)
	if err != nil {
		return persistence.QueryResult[persistence.IndexRecord, string]{}, persistence.MapError(err)
	}
	var nextID persistence.IndexRecordID
	if len(res) > query.Limit {
		nextID = persistence.IndexRecordID{IndexID: res[len(res)-1].IndexID, RecordID: res[len(res)-1].ID}
		res = res[:query.Limit]
	}
	return persistence.QueryResult[persistence.IndexRecord, string]{Items: res, NextID: nextID.Encode(), Total: total}, nil
}

func (m *modelTx) Search(query persistence.SearchQuery) (persistence.QueryResult[persistence.SearchQueryResultItem, string], error) {
	if len(query.Query) == 0 {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, fmt.Errorf("search query must be non-empty: %w", errors.ErrInvalid)
	}
	sb := strings.Builder{}
	args := make([]any, 0)

	if len(query.FromID) > 0 {
		var fromID persistence.IndexRecordID
		if err := fromID.Decode(query.FromID); err != nil {
			return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, fmt.Errorf("invalid FromID: %w", errors.ErrInvalid)
		}
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" index_record.index_id >= ? and index_record.id >= ? ")
		args = append(args, fromID.IndexID, fromID.RecordID)
	}
	if len(query.IndexIDs) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		oldLen := len(args)
		sb.WriteString(" index_record.index_id in ( ")
		for _, id := range query.IndexIDs {
			if len(args) > oldLen {
				sb.WriteString(", ")
			}
			sb.WriteString("?")
			args = append(args, id)
		}
		sb.WriteString(")")
	}
	if len(query.Tags) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		var tb strings.Builder
		tb.WriteString(" {")
		oldLen := tb.Len()
		for k, v := range query.Tags {
			if tb.Len() > oldLen {
				tb.WriteByte(',')
			}
			tb.WriteString(fmt.Sprintf("%q:%q", k, v))
		}
		tb.WriteString("}")
		sb.WriteString(" index.tags @> ?")
		args = append(args, tb.String())
	}
	if len(query.Query) > 0 {
		if len(args) > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(" index_record.segment %> ? ")
		args = append(args, query.Query)
	}

	where := sqlx.Rebind(sqlx.DOLLAR, sb.String())
	if len(where) > 0 {
		where = " where " + where
	}

	distinct := ""
	if query.Distinct {
		if query.OrderByScore {
			distinct = "distinct on(score, index_record.index_id)"
		} else {
			distinct = "distinct on(index_record.index_id)"
		}
	}

	orderBy, limit := "", 0
	if query.OrderByScore {
		orderBy = "order by score desc, index_record.index_id asc, index_record.id asc"
		limit = query.Limit // no +1, since no pagination
	} else {
		orderBy = "order by index_record.index_id asc, index_record.id asc"
		limit = query.Limit + 1
	}

	// count
	args = append(args, query.Query)
	total, err := m.getCount(fmt.Sprintf("select count(*) "+
		"from (select %s index_record.*, 1.0 - (segment <->> $%d) as score from index_record "+
		"inner join index on index.id = index_record.index_id %s %s)", distinct, len(args), where, orderBy), args...)
	if err != nil {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, persistence.MapError(err)
	}

	// query
	if query.Limit <= 0 {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{Total: total}, nil
	}

	args = append(args, query.Offset, limit)
	rows, err := m.executor().QueryxContext(m.ctx, fmt.Sprintf("select %s index_record.*, "+
		"1 - (segment <->> $%d) as score from index_record "+
		"inner join index on index.id = index_record.index_id %s %s offset $%d limit $%d",
		distinct, len(args)-2, where, orderBy, len(args)-1, len(args)), args...)
	if err != nil {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, persistence.MapError(err)
	}

	// results
	res, err := persistence.ScanRowsQueryResult[persistence.SearchQueryResultItem](rows)
	if err != nil {
		return persistence.QueryResult[persistence.SearchQueryResultItem, string]{}, persistence.MapError(err)
	}
	var nextID persistence.IndexRecordID
	if len(res) > query.Limit {
		nextID = persistence.IndexRecordID{IndexID: res[len(res)-1].IndexID, RecordID: res[len(res)-1].ID}
		res = res[:query.Limit]
	}
	return persistence.QueryResult[persistence.SearchQueryResultItem, string]{Items: res, NextID: nextID.Encode(), Total: total}, nil
}

func (m *modelTx) getCount(query string, params ...any) (int64, error) {
	rows, err := m.executor().QueryContext(m.ctx, query, params...)
	if err != nil {
		return -1, persistence.MapError(err)
	}
	defer func() {
		_ = rows.Close()
	}()
	var count int64
	if rows.Next() {
		_ = rows.Scan(&count)
	}
	return count, nil
}
