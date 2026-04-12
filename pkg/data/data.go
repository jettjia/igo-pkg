package data

import (
	"github.com/jettjia/igo-pkg/pkg/database/dbresolver"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Data struct {
	Mysql     *gorm.DB
	RedisCli  redis.UniversalClient
	DBManager *dbresolver.DBManager
	// Dynamic Tenant Database Manager
	DBManagerDynamic *dbresolver.DBManagerDynamic
}

func NewData(mysqlDB *gorm.DB, redisCli redis.UniversalClient) *Data {
	return &Data{
		Mysql:    mysqlDB,
		RedisCli: redisCli,
	}
}

type Option func(*Data)

func NewDataOption(options ...Option) *Data {
	data := &Data{}

	for _, option := range options {
		option(data)
	}
	return data
}

// WithMysql Set up database connection
func WithMysql(mysqlDB *gorm.DB) Option {
	return func(d *Data) {
		d.Mysql = mysqlDB
	}
}

// WithRedis Set up redis connection
func WithRedis(redisCli redis.UniversalClient) Option {
	return func(d *Data) {
		d.RedisCli = redisCli
	}
}

// WithDBManager Set up DBManager
func WithDBManager(DBManager *dbresolver.DBManager) Option {
	return func(d *Data) {
		d.DBManager = DBManager
	}
}

// WithDBManagerDynamic Set up DBManagerDynamic
func WithDBManagerDynamic(DBManagerDynamic *dbresolver.DBManagerDynamic) Option {
	return func(d *Data) {
		d.DBManagerDynamic = DBManagerDynamic
	}
}
