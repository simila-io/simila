package persistence

import (
	"database/sql"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	PqForeignKeyViolationError = pq.ErrorCode("23503")
	PqUniqueViolationError     = pq.ErrorCode("23505")
)

func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return errors.ErrNotExist
	}
	return mapPqError(err)
}

func mapPqError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case PqForeignKeyViolationError:
			return fmt.Errorf("%v: %w", pqErr.Message, errors.ErrConflict)
		case PqUniqueViolationError:
			return fmt.Errorf("%v: %w", pqErr.Message, errors.ErrExist)
		}
	}
	return nil
}

func ScanRows[T any](rows *sqlx.Rows) ([]T, error) {
	var res []T
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {
			return nil, MapError(err)
		}
		res = append(res, t)
	}
	return res, nil
}

func ScanRowsQueryResult[T any](rows *sqlx.Rows) ([]T, error) {
	var res []T
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {
			return nil, MapError(err)
		}
		res = append(res, t)
	}
	return res, nil
}

func ScanRowsQueryResultAndMap[T, K any](rows *sqlx.Rows, mapFn func(entity T) K) ([]K, error) {
	var res []K
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {
			return nil, MapError(err)
		}
		res = append(res, mapFn(t))
	}
	return res, nil
}
