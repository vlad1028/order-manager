package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) RunSerializable(ctx context.Context, fn func(tx pgx.Tx) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}
	return m.pool.BeginTxFunc(ctx, opts, fn)
}

func (m *TxManager) RunRepeatableRead(ctx context.Context, fn func(tx pgx.Tx) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	}
	return m.pool.BeginTxFunc(ctx, opts, fn)
}

func (m *TxManager) RunReadUncommitted(ctx context.Context, fn func(tx pgx.Tx) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.ReadUncommitted,
		AccessMode: pgx.ReadOnly,
	}
	return m.pool.BeginTxFunc(ctx, opts, fn)
}

func (m *TxManager) Run(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return m.pool.BeginFunc(ctx, fn)
}
