package db

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"time"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type PgClient struct {
	pool   *pgxpool.Pool
	ctx    context.Context
	cancel context.CancelFunc
}

func CloseStorageDb(db PgClient) {
	defer db.cancel()
	defer func() {
		db.pool.Close()
	}()
}

func InitializePgClient() (PgClient, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	// TODO: replace with configs
	storageDbUrl := os.Getenv("DATABASE_URL")

	config, err := pgxpool.ParseConfig(storageDbUrl)
	if err != nil {
		return PgClient{}, err
	}

	// TODO: handle TLS

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return PgClient{}, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return PgClient{}, err
	}

	return PgClient{pool, ctx, cancelFunc}, nil
}
