package cache

import (
	"database/sql"
	"time"
)

type PgsqlDLock struct {
	db *sql.DB
}

func NewPgsqlDLock(db *sql.DB) DLock {
	// 初始化表
	db.Exec(`CREATE TABLE IF NOT EXISTS dlock (
		key TEXT PRIMARY KEY,
		expire_at TIMESTAMPTZ
	);`)
	return &PgsqlDLock{db: db}
}

func (l *PgsqlDLock) Lock(key string, ttl time.Duration) (bool, error) {
	expireAt := time.Now().Add(ttl)
	res, err := l.db.Exec(`
		INSERT INTO dlock (key, expire_at) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET expire_at = EXCLUDED.expire_at
		WHERE dlock.expire_at < NOW()
	`, key, expireAt)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (l *PgsqlDLock) Unlock(key string) error {
	_, err := l.db.Exec(`DELETE FROM dlock WHERE key = $1`, key)
	return err
}
