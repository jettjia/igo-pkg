package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDLock struct {
	client *redis.Client
}

func NewRedisDLock(client *redis.Client) DLock {
	return &RedisDLock{client: client}
}

func (l *RedisDLock) Lock(key string, ttl time.Duration) (bool, error) {
	ok, err := l.client.SetNX(context.Background(), key, "1", ttl).Result()
	return ok, err
}

func (l *RedisDLock) Unlock(key string) error {
	return l.client.Del(context.Background(), key).Err()
}
