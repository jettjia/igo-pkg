package cache

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PgsqlCache struct {
	db *sql.DB
}

func NewPgsqlCache(opts ...Option) (KVCache, error) {
	options := &Options{
		Addr:     "127.0.0.1:5432",
		DSN:      "",
		Password: "",
	}
	for _, opt := range opts {
		opt(options)
	}
	// 优先用 DSN，否则拼接
	dsn := options.DSN
	if dsn == "" {
		dsn = fmt.Sprintf("postgres://root:%s@%s/xtext?sslmode=disable", options.Password, options.Addr)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS kv_cache (
			key TEXT PRIMARY KEY,
			value BYTEA,
			expire_at TIMESTAMPTZ
		);
		CREATE INDEX IF NOT EXISTS idx_kv_cache_expire_at ON kv_cache(expire_at);
	`)
	if err != nil {
		return nil, err
	}
	return &PgsqlCache{db: db}, nil
}

func (p *PgsqlCache) Set(key string, value []byte, ttlSeconds int64) error {
	expireAt := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	_, err := p.db.Exec(`
		INSERT INTO kv_cache (key, value, expire_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE SET value = $2, expire_at = $3
	`, key, value, expireAt)
	return err
}

func (p *PgsqlCache) Get(key string) ([]byte, error) {
	var value []byte
	var expireAt time.Time
	err := p.db.QueryRow(`
		SELECT value, expire_at FROM kv_cache WHERE key = $1
	`, key).Scan(&value, &expireAt)
	if err != nil {
		return nil, err
	}
	if time.Now().After(expireAt) {
		_ = p.Del(key)
		return nil, fmt.Errorf("key expired")
	}
	return value, nil
}

func (p *PgsqlCache) Del(key string) error {
	_, err := p.db.Exec(`DELETE FROM kv_cache WHERE key = $1`, key)
	return err
}

func (p *PgsqlCache) Close() {
	p.db.Close()
}
