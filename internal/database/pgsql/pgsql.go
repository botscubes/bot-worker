package pgsql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Db struct {
	Pool *pgxpool.Pool
}

const prefixSchema = "bot_"

func OpenConnection(url string) (*Db, error) {
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, err
	}

	return &Db{Pool: pool}, nil
}

func (db *Db) CloseConnection() {
	db.Pool.Close()
}
