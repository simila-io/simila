package persistence

import (
	"context"
	"github.com/acquirecloud/golibs/strutil"
	"github.com/jmoiron/sqlx"
	"strings"
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

func MapKeywordsToListFn(startMark, endMark string) func(item SearchQueryResultItem) SearchQueryResultItem {
	return func(item SearchQueryResultItem) SearchQueryResultItem {
		kwArr := strings.Split(item.MatchedKeywords, startMark)
		if len(kwArr) == 0 {
			return item
		}
		kwArr = kwArr[1:]
		for i := 0; i < len(kwArr); i++ {
			kwArr[i] = strings.TrimSpace(strings.Split(kwArr[i], endMark)[0])
		}
		item.MatchedKeywordsList = strutil.RemoveDups(kwArr)
		return item
	}
}
