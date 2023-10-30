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
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"github.com/acquirecloud/golibs/logging"
	"github.com/acquirecloud/golibs/ulidutils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"io/ioutil"
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
	file, err := ioutil.ReadFile(sqlScript)
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
	format.ID = newID()
	format.CreatedAt = time.Now()
	format.UpdatedAt = format.CreatedAt
	_, err := m.executor().Exec("insert into format (id, name, basis, created_at, updated_at) values ($1, $2, $3, $4, $5)",
		format.ID, format.Name, format.Basis, format.CreatedAt, format.UpdatedAt)
	if err != nil {
		return "", mapError(err)
	}
	return format.ID, nil
}

func (m *modelTx) CreateIndex(index Index) (string, error) {
	id, err := newIndexID(index)
	if err != nil {
		return "", err
	}
	index.ID = id
	index.CreatedAt = time.Now()
	index.UpdatedAt = index.CreatedAt
	_, err = m.executor().Exec("insert into index (id, format, tags, created_at, updated_at) values ($1, $2, $3, $4, $5)",
		index.ID, index.Format, index.Tags, index.CreatedAt, index.UpdatedAt)
	if err != nil {
		return "", mapError(err)
	}
	return index.ID, nil
}

func (m *modelTx) CreateIndexRecord(record IndexRecord) (string, error) {
	id, err := newRecordID(record)
	if err != nil {
		return "", err
	}
	record.ID = id
	record.CreatedAt = time.Now()
	record.UpdatedAt = record.CreatedAt
	_, err = m.executor().Exec("insert into index_record (id, index_id, segment, vector, created_at, updated_at) values ($1, $2, $3, $4, $5, $6)",
		record.ID, record.IndexID, record.Segment, record.Vector, record.CreatedAt, record.UpdatedAt)
	if err != nil {
		return "", mapError(err)
	}
	return record.ID, nil
}

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
			return errors.ErrConflict
		case PqUniqueViolationError:
			return errors.ErrExist
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

func scanRowsQueryResult[T any](rows *sqlx.Rows, total int64) (QueryResult[T], error) {
	var res []T
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {
			return QueryResult[T]{}, mapError(err)
		}
		res = append(res, t)
	}
	return QueryResult[T]{Items: res, Total: total}, nil
}

func mustEncode(v any) []byte {
	var bb bytes.Buffer
	enc := gob.NewEncoder(&bb)
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return bb.Bytes()
}

func newID() string {
	return fmt.Sprintf("%x", ulidutils.New().Bytes())
}

func newIndexID(index Index) (string, error) {
	if len(index.ID) == 0 {
		return "", fmt.Errorf("new index ID must must be specified: %w", errors.ErrInvalid)
	}
	hSum := sha1.Sum([]byte(index.ID))
	return fmt.Sprintf("%x", hSum), nil
}

func newRecordID(record IndexRecord) (string, error) {
	if len(record.IndexID) == 0 || len(record.Vector) == 0 {
		return "", fmt.Errorf("new record indexID and vector must be specified: %w", errors.ErrInvalid)
	}
	var bb bytes.Buffer
	bb.WriteString(record.IndexID)
	bb.Write(mustEncode(record.Vector))
	hSum := sha1.Sum(bb.Bytes())
	return fmt.Sprintf("%x", hSum), nil
}
