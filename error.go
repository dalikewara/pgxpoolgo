package pgxpoolgo

import (
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

const DBCodeColumnNotExists = "42703"
const DBCodeDuplicateKey = "23505"
const DBCodeInvalidInputSyntax = "22P02"

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

func NewMockErrDB(code, message string) error {
	err := &ErrDatabase{
		DBErr:     errors.New(code + "||" + message),
		DBCode:    code,
		DBMessage: message,
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
	return e.DBCode == DBCodeColumnNotExists
}

func (e *ErrDatabase) IsDuplicateKey() bool {
	return e.DBCode == DBCodeDuplicateKey
}

func (e *ErrDatabase) IsInvalidInputSyntax() bool {
	return e.DBCode == DBCodeInvalidInputSyntax
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
