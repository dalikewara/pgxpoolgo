package pgxpoolgo_test

import (
	"context"
	"errors"
	"github.com/dalikewara/pgxpoolgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func poolBeginInsertUser(ctx context.Context, pool pgxpoolgo.Pool, username, email string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil
	}

	commandTag, err := tx.Exec(ctx, `INSERT INTO users (username, email) VALUES ($1, $2)`, username, email)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() < 1 {
		return errors.New("no user was inserted")
	}

	commandTag, err = tx.Exec(ctx, `INSERT INTO profiles (username, email) VALUES ($1, $2)`, username, email)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() < 1 {
		return errors.New("no profile was inserted")
	}

	if err = tx.Commit(ctx); err != nil {
		return nil
	}

	return nil
}

func TestPoolBeginInsertUser_OK(t *testing.T) {
	username := "johndoe"
	email := "johndoe@email.com"
	ctx := context.Background()
	mockPool := pgxpoolgo.NewMockPool(t)
	assert.Implements(t, (*pgxpoolgo.Pool)(nil), mockPool)

	mockTx := pgxpoolgo.NewMockTx(t)
	mockPool.On("Begin", ctx).Return(mockTx, nil).Once()

	mockCommandTag := pgxpoolgo.NewMockCommandTag("INSERT", int64(1))
	mockTx.On("Exec", ctx, `INSERT INTO users (username, email) VALUES ($1, $2)`, username, email).Return(mockCommandTag, nil)

	mockCommandTag = pgxpoolgo.NewMockCommandTag("INSERT", int64(1))
	mockTx.On("Exec", ctx, `INSERT INTO profiles (username, email) VALUES ($1, $2)`, username, email).Return(mockCommandTag, nil)

	mockTx.On("Commit", ctx).Return(nil).Once()

	err := poolBeginInsertUser(ctx, mockPool, username, email)
	assert.Equal(t, true, mockPool.AssertCalled(t, "Begin", ctx))
	assert.Equal(t, true, mockPool.AssertExpectations(t))
	assert.Equal(t, true, mockTx.AssertCalled(t, "Exec", ctx, `INSERT INTO users (username, email) VALUES ($1, $2)`, username, email))
	assert.Equal(t, true, mockTx.AssertCalled(t, "Exec", ctx, `INSERT INTO profiles (username, email) VALUES ($1, $2)`, username, email))
	assert.Equal(t, true, mockTx.AssertExpectations(t))
	assert.Nil(t, err)
}
