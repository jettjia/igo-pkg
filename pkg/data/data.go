package data

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jettjia/igo-pkg/pkg/database/dbresolver"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Data struct {
	Mysql          *gorm.DB
	RedisCli       redis.UniversalClient
	SearchCli      *elasticsearch.Client
	SearchCliTyped *elasticsearch.TypedClient
	DBManager      *dbresolver.DBManager
	// Dynamic Tenant Database Manager
	DBManagerDynamic *dbresolver.DBManagerDynamic
}

func NewData(mysqlDB *gorm.DB, redisCli redis.UniversalClient, searchCli *elasticsearch.Client) *Data {
	return &Data{
		Mysql:     mysqlDB,
		RedisCli:  redisCli,
		SearchCli: searchCli,
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

// WithSearch Set up elasticsearch connection
func WithSearch(searchCli *elasticsearch.Client) Option {
	return func(d *Data) {
		d.SearchCli = searchCli
	}
}

// WithSearch Set up elasticsearchTyped connection
func WithSearchTyped(searchCli *elasticsearch.TypedClient) Option {
	return func(d *Data) {
		d.SearchCliTyped = searchCli
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
