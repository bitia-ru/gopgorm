package pgorm

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Db struct {
	ctx        context.Context
	connection *pgx.Conn
}

func Connect(ctx context.Context, url string) (*Db, error) {
	conn, err := pgx.Connect(ctx, url)

	if err != nil {
		return nil, err
	}

	return &Db{
		ctx:        ctx,
		connection: conn,
	}, nil
}

func (db *Db) Disconnect() error {
	return db.connection.Close(db.ctx)
}
