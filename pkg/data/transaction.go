package data

import (
	"context"

	"github.com/jettjia/go-pkg/pkg/tenant"
	"gorm.io/gorm"
)

// Transaction Transaction Management
type Transaction interface {
	ExecTx(context.Context, func(ctx context.Context) error) error
}

// Transaction Context
type contextTxKey struct{}

// NewTransaction .
func NewTransaction(d *Data) Transaction {
	return d
}

// ExecTx gorm Transaction
func (d *Data) ExecTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.Mysql.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctx)
	})
}

func (d *Data) DB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB); ok {
		return tx
	}

	tenantID := tenant.GetTenantID(ctx)
	// fmt.Printf("DB method - Context values: %+v\n", ctx)
	// fmt.Printf("DB method - TenantID: %s\n", tenantID)

	if tenantID != "" && d.DBManagerDynamic != nil {
		if db := d.DBManagerDynamic.GetDB(tenantID); db != nil {
			return db
		}
	}

	return d.Mysql
}
