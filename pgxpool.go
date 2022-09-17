package pgxpoolgo

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Connect runs pgxpool.Connect.
func Connect(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	return pgxpool.Connect(ctx, connString)
}

// ConnectConfig runs pgxpool.ConnectConfig.
func ConnectConfig(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
	return pgxpool.ConnectConfig(ctx, config)
}

// ParseConfig runs pgxpool.ParseConfig.
func ParseConfig(connString string) (*pgxpool.Config, error) {
	return pgxpool.ParseConfig(connString)
}
