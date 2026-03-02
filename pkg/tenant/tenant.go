package tenant

import (
	"context"

	"github.com/gin-gonic/gin"
)

var TenantIDKey = struct{}{}

func GetTenantID(ctx context.Context) string {
	if gc, ok := ctx.(*gin.Context); ok {
		if tid, exists := gc.Get("tenant_id"); exists {
			if tenantID, ok := tid.(string); ok {
				return tenantID
			}
		}
	}

	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}

	return ""
}

func NewTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}
