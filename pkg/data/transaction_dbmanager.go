package data

import (
	"context"

	"gorm.io/gorm"
)

// Transaction Transaction Management
type TransactionForDBManager interface {
	ExecTxForDBManager(context.Context, string, func(ctx context.Context) error) error
}

// Transaction Context
type contextTxKeyForDBManager struct{}

// NewTransaction .
func NewTransactionForDBManager(d *Data) TransactionForDBManager {
	return d
}

// ExecTx gorm Transaction
func (d *Data) ExecTxForDBManager(ctx context.Context, sourceName string, fn func(ctx context.Context) error) error {
	return d.DBManager.Sources[sourceName].WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKeyForDBManager{}, tx)
		return fn(ctx)
	})
}

func (d *Data) DBForDBManager(ctx context.Context, sourceName string) *gorm.DB {
	tx, ok := ctx.Value(contextTxKeyForDBManager{}).(*gorm.DB)
	if ok {
		return tx
	}
	return d.DBManager.Sources[sourceName]
}
