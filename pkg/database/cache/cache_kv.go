package cache

import "fmt"

// KVCache 是通用 KV 接口
type KVCache interface {
	Set(key string, value []byte, ttlSeconds int64) error
	Get(key string) ([]byte, error)
	Del(key string) error
	Close()
}

// Option 是可选参数类型（可扩展）
type Option func(*Options)

type Options struct {
	Addr     string // redis/pgsql 地址
	DSN      string // pgsql dsn
	Password string // redis/pgsql 密码
}

// NewKVCache 工厂方法
func NewKVCache(cacheType string, opts ...Option) (KVCache, error) {
	switch cacheType {
	case "redis":
		return NewRedisCache(opts...)
	case "pgsql":
		return NewPgsqlCache(opts...)
	default:
		return nil, fmt.Errorf("unknown cache type: %s", cacheType)
	}
}

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

func WithDSN(dsn string) Option {
	return func(o *Options) {
		o.DSN = dsn
	}
}

func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}
