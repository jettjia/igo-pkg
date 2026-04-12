package db

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/jettjia/igo-pkg/pkg/conf"
	"github.com/jettjia/igo-pkg/pkg/database/gormext"
)

type DB struct {
	Conf *DBConfig
	Conn *gorm.DB
}

var (
	dbOnce sync.Once
	dbImpl *DB
)

func NewDBClient(conf *conf.Config) *DB {

	dbOnce.Do(func() {
		cfg := tranCfg(conf)

		dbImpl = &DB{
			Conf: &cfg,
		}
		dbImpl.getConn()
	})

	return dbImpl
}

// Used to create non-singleton database connections
func NewDBClientWithDB(conf *conf.Config, dbName string) *DB {
	newConf := conf
	db := &DB{
		Conf: &DBConfig{
			DbType:        newConf.DB.DbType,
			Host:          newConf.DB.DbHost,
			Port:          newConf.DB.DbPort,
			User:          newConf.DB.Username,
			Password:      newConf.DB.Password,
			Db:            dbName,
			DbChar:        newConf.DB.Charset,
			MaxIdleConn:   newConf.DB.MaxIdleConn,
			MaxOpenConn:   newConf.DB.MaxOpenConn,
			MaxLifetime:   newConf.DB.ConnMaxLifetime,
			LogMode:       newConf.DB.LogMode,
			SlowThreshold: newConf.DB.SlowThreshold,
		},
	}

	db.getConn()
	return db
}

func tranCfg(conf *conf.Config) (cfg DBConfig) {
	cfg.DbType = conf.DB.DbType
	cfg.Host = conf.DB.DbHost
	cfg.Port = conf.DB.DbPort
	cfg.User = conf.DB.Username
	cfg.Password = conf.DB.Password
	cfg.Db = conf.DB.DbName
	cfg.DbChar = conf.DB.Charset
	cfg.MaxIdleConn = conf.DB.MaxIdleConn
	cfg.MaxOpenConn = conf.DB.MaxOpenConn
	cfg.MaxLifetime = conf.DB.ConnMaxLifetime
	cfg.LogMode = conf.DB.LogMode
	cfg.SlowThreshold = conf.DB.SlowThreshold

	return
}

type DBConfig struct {
	DbType        string // Database Type
	Host          string // Server Address
	Port          int    // port
	User          string // Database Username
	Password      string // Database Password
	Db            string // Database Name
	DbChar        string // Character Set
	MaxIdleConn   int    // Max Idle Connections
	MaxOpenConn   int    // Max Connections
	MaxLifetime   int    // Maximum Lifetime (s)
	LogMode       int    // Log Level(gorm; 1: Silent, 2:Error,3:Warn,4:Info)
	SlowThreshold int    // Slow SQL Judgment Time (s)
}

func (db *DB) getConn() *gorm.DB {
	loggerDefault := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Duration(db.Conf.SlowThreshold) * time.Second,
			LogLevel:                  logger.LogLevel(db.Conf.LogMode),
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
		},
	)

	cfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: loggerDefault,
	}

	var dialector gorm.Dialector
	switch db.Conf.DbType {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			db.Conf.User, db.Conf.Password, db.Conf.Host, db.Conf.Port, db.Conf.Db, db.Conf.DbChar)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=%s",
			db.Conf.Host, db.Conf.User, db.Conf.Password, db.Conf.Db, db.Conf.Port, getTimeZone())
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(db.Conf.Db)
	default:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			db.Conf.User, db.Conf.Password, db.Conf.Host, db.Conf.Port, db.Conf.Db, db.Conf.DbChar)
		dialector = gormext.Open(dsn)
	}

	conn, err := gorm.Open(dialector, cfg)
	if err != nil {
		panic(err)
	}
	db.Conn = conn

	sqlDB, _ := conn.DB()
	sqlDB.SetMaxIdleConns(db.Conf.MaxIdleConn)
	sqlDB.SetMaxOpenConns(db.Conf.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Duration(db.Conf.MaxLifetime) * time.Second)

	return conn
}

func getTimeZone() string {
	if tz := os.Getenv("TZ"); tz != "" {
		return tz
	}
	return "UTC"
}
