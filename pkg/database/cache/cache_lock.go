package cache

import (
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"
)

type DLock interface {
	Lock(key string, ttl time.Duration) (bool, error) // 加锁，返回是否成功
	Unlock(key string) error                          // 解锁
}

func NewDLock(lockType string, backend interface{}) DLock {
	switch lockType {
	case "redis":
		return NewRedisDLock(backend.(*redis.Client))
	case "pgsql":
		return NewPgsqlDLock(backend.(*sql.DB))
	default:
		return nil
	}
}
