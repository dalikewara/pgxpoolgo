package pgxpoolgo_test

import (
	"context"
	"errors"
	"github.com/dalikewara/pgxpoolgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func poolExecInsertUser(ctx context.Context, pool pgxpoolgo.Pool, username, email string) error {
	commandTag, err := pool.Exec(ctx, `INSERT INTO users (username, email) VALUES ($1, $2)`, username, email)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() < 1 {
		return errors.New("no user was inserted")
	}

	return nil
}

func TestPoolExecInsertUser_OK(t *testing.T) {
	username := "johndoe"
	email := "johndoe@email.com"
	ctx := context.Background()
	mockPool := pgxpoolgo.NewMockPool(t)
	assert.Implements(t, (*pgxpoolgo.Pool)(nil), mockPool)

	mockCommandTag := pgxpoolgo.NewMockCommandTag("INSERT", int64(1))
	mockPool.On("Exec", ctx, `INSERT INTO users (username, email) VALUES ($1, $2)`, username, email).Return(mockCommandTag, nil).Once()

	err := poolExecInsertUser(ctx, mockPool, username, email)
	assert.Equal(t, true, mockPool.AssertCalled(t, "Exec", ctx, `INSERT INTO users (username, email) VALUES ($1, $2)`, username, email))
	assert.Equal(t, true, mockPool.AssertExpectations(t))
	assert.Nil(t, err)
}
