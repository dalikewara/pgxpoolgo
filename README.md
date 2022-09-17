# pgxpoolgo

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/dalikewara/pgxpoolgo)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/dalikewara/pgxpoolgo)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/dalikewara/pgxpoolgo)
![GitHub license](https://img.shields.io/github/license/dalikewara/pgxpoolgo)

**pgxpoolgo** is based on **[pgxpool](https://github.com/jackc/pgx)**, but provides `Pool` interface
and ability to mock the `pgxpool` connection for unit testing.

This package also provides some mock function that based on **[pgxmock](https://github.com/pashagolub/pgxmock)**
like `NewMockRows`.

## Getting started

### Installation

You can use the `go get` method:

```bash
go get github.com/dalikewara/pgxpoolgo
```

### Features

- Mock support for these instance:
  - `pgxpool.Pool`
  - `pgx.Rows`
  - `pgx.Row`
  - `pgconn.CommandTag`
  - `pgx.Tx`

### Todo

- Add mock support for these instance:
  - `pgxpool.Conn`
  - `pgxpool.Config`
  - `pgxpool.Stat`
  - `pgx.BatchResults`
  - `pgx.CopyFromSource`
  - `pgx.Batch`
  - `pgx.QueryFuncRow`

### Usage

#### Pool.Query

```go
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
```

#### Pool.QueryRow

```go
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
```

#### Pool.Exec

```go
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
```

#### Pool.Begin

```go
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
```

## Release

### Changelog

Read at [CHANGELOG.md](https://github.com/dalikewara/pgxpoolgo/blob/master/CHANGELOG.md)

### Credits

The original `pgxpool` package belongs to [https://github.com/jackc/pgx](https://github.com/jackc/pgx)

The original `pgxmock` package belongs to [https://github.com/pashagolub/pgxmock](https://github.com/pashagolub/pgxmock)

### License

[MIT License](https://github.com/dalikewara/pgxpoolgo/blob/master/LICENSE)
