package cache

import (
	"testing"
	"time"
)

// go test -v -run TestRedisCache ./
func TestRedisCache(t *testing.T) {
	cache, err := NewKVCache("redis",
		WithAddr("127.0.0.1:6379"),
		WithPassword("admin123"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	key := "foo"
	val := []byte("bar")
	if err := cache.Set(key, val, 2); err != nil {
		t.Fatal(err)
	}
	got, err := cache.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "bar" {
		t.Fatalf("expected bar, got %s", got)
	}
	time.Sleep(3 * time.Second)
	_, err = cache.Get(key)
	if err == nil {
		t.Fatal("expected key expired")
	}
}

// go test -v -run TestPgsqlCache ./
func TestPgsqlCache(t *testing.T) {
	cache, err := NewKVCache("pgsql",
		WithAddr("127.0.0.1:5432"),
		WithPassword("admin123"))
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	key := "foo"
	val := []byte("bar")
	if err := cache.Set(key, val, 2); err != nil {
		t.Fatal(err)
	}
	got, err := cache.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "bar" {
		t.Fatalf("expected bar, got %s", got)
	}
	time.Sleep(3 * time.Second)
	_, err = cache.Get(key)
	if err == nil {
		t.Fatal("expected key expired")
	}
}
