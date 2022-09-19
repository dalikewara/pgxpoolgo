package pgxpoolgo

import (
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

const ErrDBCodeColumnNotExists = "42703"
const ErrDBCodeDuplicateKey = "23505"
const ErrDBCodeInvalidInputSyntax = "22P02"

type ErrDatabase struct {
	DBErr     error
	DBCode    string
	DBMessage string
}

func ErrDB(e error) *ErrDatabase {
	err := &ErrDatabase{
		DBErr: e,
	}
	err.extract()
	return err
}

func NewMockErrDB(code string) error {
	err := &ErrDatabase{
		DBErr:  errors.New(code),
		DBCode: code,
	}
	return err
}

func (e *ErrDatabase) Error() string {
	return e.DBErr.Error()
}

func (e *ErrDatabase) Code() string {
	return e.DBCode
}

func (e *ErrDatabase) Message() string {
	return e.DBMessage
}

func (e *ErrDatabase) IsNoRows() bool {
	return e.DBErr.Error() == pgx.ErrNoRows.Error()
}

func (e *ErrDatabase) IsColumnNotExists() bool {
	return e.DBCode == ErrDBCodeColumnNotExists
}

func (e *ErrDatabase) IsDuplicateKey() bool {
	return e.DBCode == ErrDBCodeDuplicateKey
}

func (e *ErrDatabase) IsInvalidInputSyntax() bool {
	return e.DBCode == ErrDBCodeInvalidInputSyntax
}

func (e *ErrDatabase) extract() {
	var pgErr *pgconn.PgError
	if errors.As(e.DBErr, &pgErr) {
		e.DBCode = pgErr.Code
		e.DBMessage = pgErr.Message
	} else {
		var errDB *ErrDatabase
		if errors.As(e.DBErr, &errDB) {
			e.DBCode = errDB.DBCode
			e.DBMessage = errDB.DBMessage
		}
	}
}
