package persistence

import (
	"context"
	"github.com/jmoiron/sqlx"
)

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

func Scan[T any](rows *sqlx.Rows) (T, error) {
	var res T
	if rows.Next() {
		_ = rows.Scan(&res)
	}
	return res, nil
}

func Count(ctx context.Context, q sqlx.QueryerContext, query string, params ...any) (int64, error) {
	rows, err := q.QueryxContext(ctx, query, params...)
	if err != nil {
		return -1, MapError(err)
	}
	defer func() {
		_ = rows.Close()
	}()
	return Scan[int64](rows)
}
