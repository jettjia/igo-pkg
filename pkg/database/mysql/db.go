package mysql

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/jettjia/go-pkg/pkg/conf"
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

	if db.Conf.DbType == "mysql" {
		return db.getConnMysql()
	}

	return db.getConnPostgres()
}

func (db *DB) getConnPostgres() *gorm.DB {
	var (
		dsn string
		err error
	)

	dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=%s",
		db.Conf.Host, db.Conf.User, db.Conf.Password, db.Conf.Db, db.Conf.Port, getTimeZone())

	// sql_gorm logger
	loggerDefault := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Duration(db.Conf.SlowThreshold) * time.Second, // Slow SQL
			LogLevel:                  logger.LogLevel(db.Conf.LogMode),                   // Log level
			Colorful:                  true,                                               // Color Printing
			IgnoreRecordNotFoundError: true,                                               // close not found error
		},
	)

	cfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: loggerDefault,
	}

	if db.Conn, err = gorm.Open(postgres.Open(dsn), cfg); err != nil {
		panic(err)
	}

	sqlDB, _ := db.Conn.DB()
	sqlDB.SetMaxIdleConns(db.Conf.MaxIdleConn)
	sqlDB.SetMaxOpenConns(db.Conf.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Duration(db.Conf.MaxLifetime) * time.Second)

	return db.Conn
}

func getTimeZone() string {
	if tz := os.Getenv("TZ"); tz != "" {
		return tz
	}
	return "UTC"
}

func (db *DB) getConnMysql() *gorm.DB {
	var (
		dsn string
		err error
	)
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", db.Conf.User, db.Conf.Password, db.Conf.Host, db.Conf.Port, db.Conf.Db, db.Conf.DbChar)

	// sql_gorm logger 配置
	loggerDefault := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
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

	if db.Conn, err = gorm.Open(mysql.Open(dsn), cfg); err != nil {
		panic(err)
	}

	sqlDB, _ := db.Conn.DB()
	sqlDB.SetMaxIdleConns(db.Conf.MaxIdleConn)
	sqlDB.SetMaxOpenConns(db.Conf.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Duration(db.Conf.MaxLifetime) * time.Second)

	return db.Conn
}
