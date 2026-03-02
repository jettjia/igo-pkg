package redis

import (
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/jettjia/igo-pkg/pkg/conf"
)

type RedisDB struct {
	Conn redis.UniversalClient
}

type RedisConfig struct {
	RedisType  string // redis使用模式:alone, sentinel,cluster
	Addrs      []string
	Password   string
	MasterName string
	PoolSize   int
}

var (
	once   sync.Once
	rdConn *RedisDB
)

// NewRedisClient 获取redis的链接
func NewRedisClient(conf *conf.Config) *RedisDB {

	once.Do(func() {
		rdConn = &RedisDB{}

		// 转换配置
		cfg := tranCfg(conf)

		rdConn.Conn = getConn(&cfg)
	})

	return rdConn
}

// 转换数据库配置
func tranCfg(conf *conf.Config) (redisConf RedisConfig) {
	redisConf.Addrs = []string{
		conf.Cache.Addr,
	}
	redisConf.Password = conf.Cache.Password
	redisConf.MasterName = conf.Cache.MasterName
	redisConf.PoolSize = conf.Cache.PoolSize
	redisConf.RedisType = conf.Cache.RedisType

	return
}

func getConn(cfg *RedisConfig) redis.UniversalClient {
	var (
		rdb redis.UniversalClient
		typ string
	)

	typ = cfg.RedisType

	switch typ {
	case "alone":
		rdb = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:    cfg.Addrs,
			Password: cfg.Password,
			PoolSize: cfg.PoolSize,
		})
	case "sentinel":
		rdb = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      cfg.Addrs,
			MasterName: cfg.MasterName,
			Password:   cfg.Password,
			// To route commands by latency or randomly, enable one of the following.
			RouteByLatency: true,
			RouteRandomly:  true,
			PoolSize:       cfg.PoolSize,
		})
	case "cluster":
		rdb = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:    cfg.Addrs,
			Password: cfg.Password,
			// To route commands by latency or randomly, enable one of the following.
			RouteByLatency: true,
			RouteRandomly:  true,
			PoolSize:       cfg.PoolSize,
		})
	default:
		panic("redis link type error, Link type must be:alone,sentinel,cluster")
	}

	return rdb
}
