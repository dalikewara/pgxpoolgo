package pgxpoolgo

import (
	"github.com/jackc/pgconn"
	"github.com/pashagolub/pgxmock"
)

// NewMockCommandTag mocks pgconn.CommandTag.
func NewMockCommandTag(op string, rowsAffected int64) pgconn.CommandTag {
	return pgxmock.NewResult(op, rowsAffected)
}

// NewMockCommandTagError mocks pgconn.CommandTag error.
func NewMockCommandTagError(err error) pgconn.CommandTag {
	return pgxmock.NewErrorResult(err)
}
