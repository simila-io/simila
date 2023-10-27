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
	"io/ioutil"
	"strings"
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

// newModelTx returns the new ModelTx object
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
		err := t.execQuery(request)
		if err != nil {
			return fmt.Errorf("could not execute %s in ExecScript: %w", request, err)
		}
	}
	return nil
}

// ============================== modelTx ====================================
func (m *modelTx) CreateTestRecord(t TestRecord) (string, error) {
	t.ID = ulidutils.NewID()
	res, err := m.executor().Exec("INSERT INTO testtable (id) VALUES ($1)", t.ID)
	if err != nil {
		return "", mapError(err)
	}
	i, _ := res.RowsAffected()
	if i == 0 {
		return "", fmt.Errorf("already exists: %w", errors.ErrExist)
	}
	return t.ID, nil
}

const (
	PqUniqueViolationError = pq.ErrorCode("23505")
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
