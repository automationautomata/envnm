package postgres

import (
	"context"
	"envmn/internal/repository/postgres/queries"

	"github.com/jackc/pgx/v5/pgxpool"
)

type connection struct {
	db *pgxpool.Pool
	*queries.Queries
}

func NewConnection(db *pgxpool.Pool) *connection {
	return &connection{db: db, Queries: queries.New(db)}
}

func (c *connection) transaction(ctx context.Context, fn func(*queries.Queries) error) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(c.Queries.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
