package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(opts ...Option) (KVCache, error) {
	options := &Options{
		Addr: "127.0.0.1:6379",
	}
	for _, opt := range opts {
		opt(options)
	}
	client := redis.NewClient(&redis.Options{
		Addr:     options.Addr,
		Password: options.Password,
	})
	return &RedisCache{client: client}, nil
}

func (r *RedisCache) Set(key string, value []byte, ttlSeconds int64) error {
	return r.client.Set(context.Background(), key, value, time.Duration(ttlSeconds)*time.Second).Err()
}

func (r *RedisCache) Get(key string) ([]byte, error) {
	return r.client.Get(context.Background(), key).Bytes()
}

func (r *RedisCache) Del(key string) error {
	return r.client.Del(context.Background(), key).Err()
}

func (r *RedisCache) Close() {
	r.client.Close()
}
