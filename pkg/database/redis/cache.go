package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client redis.UniversalClient
	ctx    context.Context
}

func NewRedisCache(client redis.UniversalClient) *RedisCache {
	return &RedisCache{
		client: client,
		ctx:    context.Background(),
	}
}

// 注意：这里的Set、Get等方法需要根据redis的操作进行调整，例如使用SetEX进行带过期时间的设置。
func (r *RedisCache) Set(key string, value interface{}, ttl int) error {
	err := r.client.SetEx(r.ctx, key, value, time.Duration(ttl)*time.Second).Err()
	return err
}

func (r *RedisCache) Get(key string) (interface{}, bool) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		return nil, false
	}
	return val, true
}

func (r *RedisCache) Delete(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	return err
}

func (r *RedisCache) Exists(key string) bool {
	return r.client.Exists(r.ctx, key).Val() > 0
}
