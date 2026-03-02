package cache

import (
	"database/sql"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// go test -v -run TestRedisDLock ./
func TestRedisDLock(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "admin123",
	})
	lock := NewRedisDLock(client)
	key := "dlock-redis-demo"
	ok, err := lock.Lock(key, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("redis lock failed")
	}
	// 再次加锁应失败
	ok2, _ := lock.Lock(key, 2*time.Second)
	if ok2 {
		t.Fatal("should not get lock again before expire")
	}
	// 解锁
	if err := lock.Unlock(key); err != nil {
		t.Fatal(err)
	}
	// 解锁后可再次加锁
	ok3, _ := lock.Lock(key, 2*time.Second)
	if !ok3 {
		t.Fatal("should get lock after unlock")
	}
	lock.Unlock(key)
}

// go test -v -run TestPgsqlDLock ./
func TestPgsqlDLock(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://root:admin123@127.0.0.1:5432/xtext?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	lock := NewPgsqlDLock(db)
	key := "dlock-pgsql-demo"
	ok, err := lock.Lock(key, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("pgsql lock failed")
	}
	// 再次加锁应失败
	ok2, _ := lock.Lock(key, 2*time.Second)
	if ok2 {
		t.Fatal("should not get lock again before expire")
	}
	// 解锁
	if err := lock.Unlock(key); err != nil {
		t.Fatal(err)
	}
	// 解锁后可再次加锁
	ok3, _ := lock.Lock(key, 2*time.Second)
	if !ok3 {
		t.Fatal("should get lock after unlock")
	}
	lock.Unlock(key)
}
