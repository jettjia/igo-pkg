package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jettjia/igo-pkg/pkg/conf"
)

// go test -v -run Test_RedisClient ./
func Test_RedisClient(t *testing.T) {
	var pkgConf = conf.Config{
		Cache: conf.CacheConf{
			RedisType:  "sentinel",
			Addr:       "127.0.0.1:6379",
			Password:   "admin123",
			MasterName: "",
			PoolSize:   10,
		},
	}
	rdb := NewRedisClient(&pkgConf)

	rdb.Conn.Set(context.TODO(), "test", "testValue", 5*time.Second)

	data := rdb.Conn.Get(context.TODO(), "test")
	fmt.Println(data.Val())
}

// go test -v -run Test_RedisClient ./
func Test_NewRedisCache(t *testing.T) {
	var pkgConf = conf.Config{
		Cache: conf.CacheConf{
			RedisType:  "sentinel",
			Addr:       "127.0.0.1:6379",
			Password:   "admin123",
			MasterName: "",
			PoolSize:   10,
		},
	}
	rdb := NewRedisClient(&pkgConf)

	rdCache := NewRedisCache(rdb.Conn)
	rdCache.Set("test", "test", 10)

	data, exist := rdCache.Get("test")
	if !exist {
		t.Error("data not exist")
	}

	fmt.Println(data)
}
