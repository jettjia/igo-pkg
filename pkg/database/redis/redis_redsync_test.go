package redis

import (
	"sync"
	"testing"
	"time"

	"github.com/jettjia/go-pkg/pkg/conf"
	"github.com/stretchr/testify/assert"
)

// go test -v -run Test_RedisLock ./
func Test_RedisLock(t *testing.T) {
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

	// 测试基本加锁和解锁
	t.Run("basic lock and unlock", func(t *testing.T) {
		lock := NewRedisLock(rdb.Conn, "test_lock")
		lock.SetExpire(1) // 设置1秒过期

		// 获取锁
		acquired, err := lock.Acquire()
		assert.NoError(t, err)
		assert.True(t, acquired)

		// 重复获取锁应该失败
		acquired2, err := lock.Acquire()
		assert.NoError(t, err)
		assert.False(t, acquired2)

		// 释放锁
		released, err := lock.Release()
		assert.NoError(t, err)
		assert.True(t, released)
	})

	// 测试锁过期
	t.Run("lock expiration", func(t *testing.T) {
		lock := NewRedisLock(rdb.Conn, "test_lock_expire")
		lock.SetExpire(1) // 设置1秒过期

		// 获取锁
		acquired, err := lock.Acquire()
		assert.NoError(t, err)
		assert.True(t, acquired)

		// 等待锁过期
		time.Sleep(2 * time.Second)

		// 另一个客户端应该能够获取锁
		lock2 := NewRedisLock(rdb.Conn, "test_lock_expire")
		acquired2, err := lock2.Acquire()
		assert.NoError(t, err)
		assert.True(t, acquired2)

		// 清理
		lock2.Release()
	})

	// 修改并发测试部分
	t.Run("concurrent lock acquisition", func(t *testing.T) {
		const concurrentClients = 10
		successCount := 0
		var wg sync.WaitGroup
		var mu sync.Mutex

		// 创建多个客户端同时尝试获取锁
		for i := 0; i < concurrentClients; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				lock := NewRedisLock(rdb.Conn, "test_concurrent_lock")
				lock.SetExpire(1)

				acquired, err := lock.Acquire()
				assert.NoError(t, err)
				if acquired {
					mu.Lock()
					successCount++
					mu.Unlock()
					time.Sleep(100 * time.Millisecond) // 模拟业务处理
					lock.Release()
				}
			}()
		}

		wg.Wait()
		// 确保只有一个客户端成功获取锁
		assert.Equal(t, 1, successCount)
	})

	// 测试随机字符串生成
	t.Run("random string generation", func(t *testing.T) {
		str1 := randomStr(randomLen)
		str2 := randomStr(randomLen)

		// 确保生成的字符串长度正确
		assert.Equal(t, randomLen, len(str1))
		assert.Equal(t, randomLen, len(str2))

		// 确保生成的字符串不同
		assert.NotEqual(t, str1, str2)
	})
}
