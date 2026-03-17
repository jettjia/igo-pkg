package dbresolver

import (
	"fmt"
	"sync"

	"github.com/jettjia/igo-pkg/pkg/conf"
	db "github.com/jettjia/igo-pkg/pkg/database/db"
	"gorm.io/gorm"
)

// DBManager Managing Database Connections for Multi-Tenancy
type DBManagerDynamic struct {
	mu        sync.RWMutex
	defaultDB *gorm.DB
	tenantDBs map[string]*gorm.DB
	conf      *conf.Config
}

// NewDBManager Create a new DBManager instance
func NewDBManagerDynamic(defaultDB *gorm.DB, conf *conf.Config) *DBManagerDynamic {
	return &DBManagerDynamic{
		defaultDB: defaultDB,
		tenantDBs: make(map[string]*gorm.DB),
		conf:      conf,
	}
}

// GetDB Retrieve the database connection corresponding to the tenant
func (m *DBManagerDynamic) GetDB(tenantID string) *gorm.DB {
	if tenantID == "" {
		return m.defaultDB
	}

	m.mu.RLock()
	if db, exists := m.tenantDBs[tenantID]; exists {
		m.mu.RUnlock()
		return db
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if db, exists := m.tenantDBs[tenantID]; exists {
		return db
	}

	dbName := fmt.Sprintf("%s_xtext", tenantID)
	newDB := db.NewDBClientWithDB(m.conf, dbName).Conn
	m.tenantDBs[tenantID] = newDB

	return newDB
}
