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

package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"github.com/acquirecloud/golibs/logging"
	"github.com/acquirecloud/golibs/ulidutils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"os"
	"strings"
	"time"
)

type (
	db struct {
		dn     string // driver name "postgres", "sqlite3" ...
		ds     string // data source name "user=foo dbname=bar sslmode=disable"
		logger logging.Logger
		db     *sqlx.DB
	}

	// exec is a helper interface to provide joined functionality of sqlx.DB and sqlx.Tx
	// it is used by the tx.executor()
	exec interface {
		sqlx.Queryer
		sqlx.Ext
		Get(dest interface{}, query string, args ...interface{}) error
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}

	// tx implements the Tx interface
	tx struct {
		db *sqlx.DB // never nil
		tx *sqlx.Tx // keeps active transaction, if it exists. It can be nil, if not started.
	}

	// modelTx is a helper to persist persistence objects ModelTx
	modelTx struct {
		*tx // the active transaction, never nil for the object
	}
)

// NewDb creates new db object
func NewDb(driverName, dataSourceName string) Db {
	return &db{dn: driverName, ds: dataSourceName, logger: logging.NewLogger("db." + driverName)}
}

// Init implements linker.Initializer interface
func (d *db) Init(ctx context.Context) error {
	d.logger.Infof("Initializing...")
	sdb, err := sqlx.Connect(d.dn, d.ds)
	if err != nil {
		return fmt.Errorf("could not connect to the database %s: %w", d.dn, err)
	}
	d.db = sdb
	return nil
}

// Shutdown implements linker.Shutdowner interface
func (d *db) Shutdown() {
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
func (d *db) NewModelTx() ModelTx {
	return &modelTx{tx: d.NewTx().(*tx)}
}

// NewTx returns the new Tx object
func (d *db) NewTx() Tx {
	return &tx{db: d.db}
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
	t.tx = t.db.MustBegin()
}

// MustBeginSerializable is a part of the Tx interface
func (t *tx) MustBeginSerializable(ctx context.Context) {
	tx := t.db.MustBeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
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
	_, err := t.executor().Exec(sqlQuery, params...)
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

func (m *modelTx) CreateFormat(format Format) (string, error) {
	if len(format.Basis) == 0 {
		format.Basis = []byte("{}")
	}
	format.ID = ulidutils.NewID()
	format.CreatedAt = time.Now()
	format.UpdatedAt = format.CreatedAt
	_, err := m.executor().Exec("insert into format (id, name, basis, created_at, updated_at) values ($1, $2, $3, $4, $5)",
		format.ID, format.Name, format.Basis, format.CreatedAt, format.UpdatedAt)
	if err != nil {
		return "", mapError(err)
	}
	return format.ID, nil
}

func (m *modelTx) GetFormat(name string) (Format, error) {
	var f Format
	return f, mapError(m.executor().Get(&f, "select * from format where name=$1", name))
}

func (m *modelTx) DeleteFormat(name string) error {
	res, err := m.executor().Exec("delete from format where name=$1", name)
	if err != nil {
		return mapError(err)
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return errors.ErrNotExist
	}
	return nil
}

func (m *modelTx) ListFormats() ([]Format, error) {
	rows, err := m.executor().Queryx("select * from format order by name")
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanRows[Format](rows)
}

func (m *modelTx) CreateIndex(index Index) (string, error) {
	if len(index.ID) == 0 {
		return "", fmt.Errorf("index ID must be specified: %w", errors.ErrInvalid)
	}
	if index.Tags == nil {
		index.Tags = make(Tags)
	}
	index.CreatedAt = time.Now()
	index.UpdatedAt = index.CreatedAt
	_, err := m.executor().Exec("insert into index (id, format, tags, created_at, updated_at) values ($1, $2, $3, $4, $5)",
		index.ID, index.Format, index.Tags, index.CreatedAt, index.UpdatedAt)
	if err != nil {
		return "", mapError(err)
	}
	return index.ID, nil
}

func (m *modelTx) GetIndex(ID string) (Index, error) {
	var idx Index
	return idx, mapError(m.executor().Get(&idx, "select * from index where id=$1", ID))
}

func (m *modelTx) UpdateIndex(index Index) error {
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

	res, err := m.executor().Exec(sqlx.Rebind(sqlx.DOLLAR, sb.String()), args...)
	if err != nil {
		return mapError(err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.ErrNotExist
	}
	return nil
}

func (m *modelTx) DeleteIndex(ID string) error {
	res, err := m.executor().Exec("delete from index where id=$1", ID)
	if err != nil {
		return mapError(err)
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return errors.ErrNotExist
	}
	return nil
}

func (m *modelTx) QueryIndexes(query IndexQuery) (QueryResult[Index, string], error) {
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
		for k, v := range query.Tags {
			if tb.Len() > 2 {
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

	// count
	where := sqlx.Rebind(sqlx.DOLLAR, sb.String())
	total, err := m.getCount(fmt.Sprintf("select count(*) from index where %s", where), args...)
	if err != nil {
		return QueryResult[Index, string]{}, mapError(err)
	}

	// query
	if query.Limit <= 0 {
		return QueryResult[Index, string]{Total: total}, nil
	}
	args = append(args, query.Limit+1)
	rows, err := m.executor().Queryx(fmt.Sprintf("select * from index where %s order by id limit $%d", where, len(args)), args...)
	if err != nil {
		return QueryResult[Index, string]{Total: total}, mapError(err)
	}

	// results
	res, err := scanRowsQueryResult[Index](rows)
	if err != nil {
		return QueryResult[Index, string]{}, mapError(err)
	}
	var nextID string
	if len(res) > query.Limit {
		nextID = res[len(res)-1].ID
		res = res[:query.Limit]
	}
	return QueryResult[Index, string]{Items: res, NextID: nextID, Total: total}, nil
}

func (m *modelTx) CreateIndexRecords(records []IndexRecord) error {
	var sb strings.Builder
	params := []any{}
	firstIdx := 1
	sb.WriteString("insert into index_record (id, index_id, segment, vector, created_at, updated_at) values ")
	now := time.Now()
	for i, r := range records {
		if r.ID == "" {
			return fmt.Errorf("index record for record %d ID must be specified: %w", i, errors.ErrInvalid)
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
	_, err := m.executor().Exec(sb.String(), params...)

	if err != nil {
		return mapError(err)
	}
	return nil
}

func (m *modelTx) GetIndexRecord(ID string) (IndexRecord, error) {
	var r IndexRecord
	return r, mapError(m.executor().Get(&r, "select * FROM index_record WHERE id=$1", ID))
}

func (m *modelTx) UpdateIndexRecord(record IndexRecord) error {
	if len(record.ID) == 0 {
		return fmt.Errorf("index record ID must be specified: %w", errors.ErrInvalid)
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

	sb.WriteString(", updated_at = ? where id = ?")
	args = append(args, time.Now(), record.ID)

	res, err := m.executor().Exec(sqlx.Rebind(sqlx.DOLLAR, sb.String()), args...)
	if err != nil {
		return mapError(err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.ErrNotExist
	}
	return nil
}

func (m *modelTx) DeleteIndexRecord(ID string) error {
	res, err := m.executor().Exec("delete from index_record where id=$1", ID)
	if err != nil {
		return mapError(err)
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return errors.ErrNotExist
	}
	return nil
}

func (m *modelTx) QueryIndexRecords(query IndexRecordQuery) (QueryResult[IndexRecord, string], error) {
	sb := strings.Builder{}
	args := make([]any, 0)

	if len(query.FromID) > 0 {
		var fromID IndexRecordID
		if err := fromID.Decode(query.FromID); err != nil {
			return QueryResult[IndexRecord, string]{}, fmt.Errorf("invalid FromID: %w", errors.ErrInvalid)
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

	// count
	where := sqlx.Rebind(sqlx.DOLLAR, sb.String())
	total, err := m.getCount(fmt.Sprintf("select count(*) from index_record where %s ", where), args...)
	if err != nil {
		return QueryResult[IndexRecord, string]{}, mapError(err)
	}

	// query
	if query.Limit <= 0 {
		return QueryResult[IndexRecord, string]{Total: total}, nil
	}
	args = append(args, query.Limit+1)
	rows, err := m.executor().Queryx(fmt.Sprintf("select * from index_record where %s order by index_id asc, id asc limit $%d", where, len(args)), args...)
	if err != nil {
		return QueryResult[IndexRecord, string]{Total: total}, mapError(err)
	}

	// results
	res, err := scanRowsQueryResult[IndexRecord](rows)
	if err != nil {
		return QueryResult[IndexRecord, string]{}, mapError(err)
	}
	var nextID IndexRecordID
	if len(res) > query.Limit {
		nextID = IndexRecordID{IndexID: res[len(res)-1].IndexID, RecordID: res[len(res)-1].ID}
		res = res[:query.Limit]
	}
	return QueryResult[IndexRecord, string]{Items: res, NextID: nextID.Encode(), Total: total}, nil
}

func (m *modelTx) Search(query SearchQuery) (QueryResult[SearchQueryResultItem, string], error) {
	if len(query.Query) == 0 {
		return QueryResult[SearchQueryResultItem, string]{}, fmt.Errorf("search query must be non-empty: %w", errors.ErrInvalid)
	}
	sb := strings.Builder{}
	args := make([]any, 0)

	if len(query.FromID) > 0 {
		var fromID IndexRecordID
		if err := fromID.Decode(query.FromID); err != nil {
			return QueryResult[SearchQueryResultItem, string]{}, fmt.Errorf("invalid FromID: %w", errors.ErrInvalid)
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
		for k, v := range query.Tags {
			if tb.Len() > 2 {
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
		sb.WriteString(" index_record.segment &@~ ? ")
		args = append(args, query.Query)
	}
	cntDistinct, qryDistinct := "*", ""
	if query.Distinct {
		cntDistinct = "distinct index_record.index_id"
		qryDistinct = "distinct on(index_record.index_id)"
	}

	// count
	where := sqlx.Rebind(sqlx.DOLLAR, sb.String())
	total, err := m.getCount(fmt.Sprintf("select count(%s) from index_record inner join index on index.id = index_record.index_id where %s ", cntDistinct, where), args...)
	if err != nil {
		return QueryResult[SearchQueryResultItem, string]{}, mapError(err)
	}

	// query
	if query.Limit <= 0 {
		return QueryResult[SearchQueryResultItem, string]{Total: total}, nil
	}
	args = append(args, query.Limit+1)
	rows, err := m.executor().Queryx(fmt.Sprintf("select %s index_record.*, pgroonga_score(index_record.tableoid, index_record.ctid) as score from index_record "+
		"inner join index on index.id = index_record.index_id where %s order by index_id asc, id asc limit $%d", qryDistinct, where, len(args)), args...)
	if err != nil {
		return QueryResult[SearchQueryResultItem, string]{}, mapError(err)
	}

	// results
	res, err := scanRowsQueryResult[SearchQueryResultItem](rows)
	if err != nil {
		return QueryResult[SearchQueryResultItem, string]{}, mapError(err)
	}
	var nextID IndexRecordID
	if len(res) > query.Limit {
		nextID = IndexRecordID{IndexID: res[len(res)-1].IndexID, RecordID: res[len(res)-1].ID}
		res = res[:query.Limit]
	}
	return QueryResult[SearchQueryResultItem, string]{Items: res, NextID: nextID.Encode(), Total: total}, nil
}

func (m *modelTx) getCount(query string, params ...any) (int64, error) {
	rows, err := m.executor().Query(query, params...)
	if err != nil {
		return -1, mapError(err)
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

// ============================== helpers ====================================

const (
	PqForeignKeyViolationError = pq.ErrorCode("23503")
	PqUniqueViolationError     = pq.ErrorCode("23505")
)

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return errors.ErrNotExist
	}
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case PqForeignKeyViolationError:
			return fmt.Errorf("%v: %w", pqErr.Message, errors.ErrConflict)
		case PqUniqueViolationError:
			return fmt.Errorf("%v: %w", pqErr.Message, errors.ErrExist)
		}
	}
	return err
}

func scanRows[T any](rows *sqlx.Rows) ([]T, error) {
	var res []T
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {
			return nil, mapError(err)
		}
		res = append(res, t)
	}
	return res, nil
}

func scanRowsQueryResult[T any](rows *sqlx.Rows) ([]T, error) {
	var res []T
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {
			return nil, mapError(err)
		}
		res = append(res, t)
	}
	return res, nil
}
