package pgxpoolgo_test

import (
	"context"
	"github.com/dalikewara/pgxpoolgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func poolQueryGetUserIDs(ctx context.Context, pool pgxpoolgo.Pool) ([]uint32, error) {
	var ids []uint32

	rows, err := pool.Query(ctx, `SELECT id FROM users`)
	if err != nil {
		return ids, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint32
		if err = rows.Scan(&id); err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func TestPoolQueryGetUsersIDs_OK(t *testing.T) {
	ctx := context.Background()
	mockPool := pgxpoolgo.NewMockPool(t)
	assert.Implements(t, (*pgxpoolgo.Pool)(nil), mockPool)

	mockRows := pgxpoolgo.NewMockRows([]string{"id"}).AddRow(uint32(1)).AddRow(uint32(2)).AddRow(uint32(3)).Compose()
	mockPool.On("Query", ctx, `SELECT id FROM users`).Return(mockRows, nil).Once()

	ids, err := poolQueryGetUserIDs(ctx, mockPool)
	assert.Equal(t, true, mockPool.AssertCalled(t, "Query", ctx, `SELECT id FROM users`))
	assert.Equal(t, true, mockPool.AssertExpectations(t))
	assert.Nil(t, err)
	assert.Equal(t, []uint32{1, 2, 3}, ids)
}
