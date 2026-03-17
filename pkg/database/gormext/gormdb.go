package gormext

import (
	"database/sql"
	"os"
	"strings"

	"github.com/jettjia/igo-pkg/pkg/database/gormext/dm"
	_ "github.com/kweaver-ai/proton-rds-sdk-go/driver"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(dsn string) gorm.Dialector {

	dbType := strings.ToUpper(os.Getenv("DB_TYPE"))

	open, err := sql.Open("proton-rds", dsn)
	if err != nil {
		panic(err)
	}

	if dbType == "DM8" {
		return dm.New(dm.Config{Conn: open})
	}

	if dbType == "KDB9" {
		return postgres.New(postgres.Config{Conn: open})
	}

	return mysql.New(mysql.Config{Conn: open})
}
