package pgxpoolgo

import (
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type ErrDatabase struct {
	err     error
	code    string
	message string
}

func ErrDB(e error) *ErrDatabase {
	err := &ErrDatabase{
		err: e,
	}
	err.extract()
	return err
}

func (e *ErrDatabase) Error() string {
	return e.err.Error()
}

func (e *ErrDatabase) Code() string {
	return e.code
}

func (e *ErrDatabase) Message() string {
	return e.message
}

func (e *ErrDatabase) NoRows() bool {
	return e.err.Error() == pgx.ErrNoRows.Error()
}

func (e *ErrDatabase) ColumnNotExists() bool {
	return e.code == "42703"
}

func (e *ErrDatabase) DuplicateKey() bool {
	return e.code == "23505"
}

func (e *ErrDatabase) InvalidInputSyntax() bool {
	return e.code == "22P02"
}

func (e *ErrDatabase) extract() {
	var pgErr *pgconn.PgError
	if errors.As(e.err, &pgErr) {
		e.code = pgErr.Code
		e.message = pgErr.Message
	}
}
