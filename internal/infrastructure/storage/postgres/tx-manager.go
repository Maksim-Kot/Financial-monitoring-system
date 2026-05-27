package postgres

import (
	"context"

	"fms-project/internal/domain/transaction"

	"github.com/uptrace/bun"
)

type txContextKey struct{}

func withTx(ctx context.Context, tx bun.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func CheckTx(ctx context.Context, client *Client) bun.IDB {
	tx, ok := ctx.Value(txContextKey{}).(bun.Tx)
	if ok {
		return tx
	}
	return client.db
}

type TxManager struct {
	db *bun.DB
}

func NewTxManager(client *Client) transaction.TxManager {
	return &TxManager{db: client.db}
}

func (m *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(withTx(ctx, tx))
	})
}
