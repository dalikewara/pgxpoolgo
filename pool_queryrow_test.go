package pgxpoolgo_test

import (
	"context"
	"github.com/dalikewara/pgxpoolgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func poolQueryRowGetUserID(ctx context.Context, pool pgxpoolgo.Pool) (uint32, error) {
	var id uint32

	err := pool.QueryRow(ctx, `SELECT id FROM users`).Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func TestPoolQueryRowGetUsersID_OK(t *testing.T) {
	ctx := context.Background()
	mockPool := pgxpoolgo.NewMockPool(t)
	assert.Implements(t, (*pgxpoolgo.Pool)(nil), mockPool)

	mockRow := pgxpoolgo.NewMockRow([]string{"id"}).AddRow(uint32(1)).Compose()
	mockPool.On("QueryRow", ctx, `SELECT id FROM users`).Return(mockRow, nil).Once()

	id, err := poolQueryRowGetUserID(ctx, mockPool)
	assert.Equal(t, true, mockPool.AssertCalled(t, "QueryRow", ctx, `SELECT id FROM users`))
	assert.Equal(t, true, mockPool.AssertExpectations(t))
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), id)
}
