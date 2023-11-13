package persistence

import (
	"database/sql"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
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
	return err
}
