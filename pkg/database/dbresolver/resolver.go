package dbresolver

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/gookit/color"
	"github.com/jettjia/igo-pkg/pkg/conf"
	"github.com/jettjia/igo-pkg/pkg/database/gormext"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

type DB struct {
	Conf    []DataSourceCfg
	Manager *DBManager
}

var (
	dbOnce sync.Once
	dbImpl *DB
)

func NewDBManagerClient(conf *conf.Config) *DB {

	dbOnce.Do(func() {
		cfg := tranCfg(conf)

		dbImpl = &DB{
			Conf: cfg,
		}
		dbImpl.Manager = NewDBManager(cfg)
	})

	return dbImpl
}

func tranCfg(conf *conf.Config) (cfgs []DataSourceCfg) {
	for _, val := range conf.DBManager.DataSources {
		var cfgInfo DataSourceCfg
		cfgInfo.DbType = val.DbType
		cfgInfo.Name = val.Name
		cfgInfo.MasterDSN = val.MasterDSN
		cfgInfo.SlaveDSNs = val.SlaveDSNs
		cfgInfo.MaxIdleConn = val.MaxIdleConn
		cfgInfo.MaxOpenConn = val.MaxOpenConn
		cfgInfo.MaxLifetime = val.MaxLifetime
		cfgInfo.LogMode = val.LogMode
		cfgInfo.SlowThreshold = val.SlowThreshold

		cfgs = append(cfgs, cfgInfo)
	}

	return
}

type DataSourceCfg struct {
	DbType       string            `yaml:"db_type"`
	Name         string            `yaml:"name"`
	MasterDSN    string            `yaml:"master_dsn"`
	SlaveDSNs    []string          `yaml:"slave_dsns"`
	ResolverConf dbresolver.Config `yaml:"resolver_conf"`

	MaxIdleConn   int
	MaxOpenConn   int
	MaxLifetime   int
	LogMode       int
	SlowThreshold int
}

type DBManager struct {
	Sources map[string]*gorm.DB
}

func NewDBManager(cfgs []DataSourceCfg) *DBManager {
	manager := &DBManager{Sources: make(map[string]*gorm.DB)}

	for _, cfg := range cfgs {
		db, err := initDB(cfg.DbType, cfg.MasterDSN, cfg.SlaveDSNs, cfg.ResolverConf, cfg)
		if err != nil {
			color.Red.Println("NewDBManager.initDB.err:", err)
			panic(err)
		}
		manager.Sources[cfg.Name] = db
	}

	return manager
}

func initDB(dbType string, masterDSN string, slaveDSNs []string, resolverConf dbresolver.Config, cfg DataSourceCfg) (*gorm.DB, error) {
	if cfg.MaxIdleConn == 0 {
		cfg.MaxIdleConn = 10
	}
	if cfg.MaxOpenConn == 0 {
		cfg.MaxOpenConn = 100
	}
	if cfg.MaxLifetime == 0 {
		cfg.MaxLifetime = 60
	}
	if cfg.LogMode == 0 {
		cfg.LogMode = 4
	}
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = 100
	}

	loggerDefault := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Duration(cfg.SlowThreshold) * time.Second,
			LogLevel:                  logger.LogLevel(cfg.LogMode),
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
		},
	)

	dsn := masterDSN

	db, err := gorm.Open(dialectorFor(dbType, dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		SkipDefaultTransaction: true,
		Logger:                 loggerDefault,
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	masterDialector := dialectorFor(dbType, dsn)

	replicaDialectors := make([]gorm.Dialector, len(slaveDSNs))
	for i, dsn := range slaveDSNs {
		replicaDialectors[i] = dialectorFor(dbType, dsn)
	}

	resolverConf.Sources = append(resolverConf.Sources, masterDialector)
	resolverConf.Replicas = append(resolverConf.Replicas, replicaDialectors...)

	err = db.Use(dbresolver.Register(resolverConf))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func dialectorFor(dbType string, dsn string) gorm.Dialector {
	switch dbType {
	case "mysql":
		return mysql.Open(dsn)
	case "postgres":
		return postgres.Open(dsn)
	default:
		return gormext.Open(dsn)
	}
}
