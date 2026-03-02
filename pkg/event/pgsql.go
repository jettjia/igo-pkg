package event

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createQueueTableSQL = `
CREATE TABLE IF NOT EXISTS event_queue (
    id BIGSERIAL PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    payload BYTEA,
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_until TIMESTAMPTZ,
    processing_attempts INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_event_queue_topic_status_created_at ON event_queue (topic, status, created_at);
`
)

// EventPgClient implements the EventMQClient interface for PostgreSQL.
type EventPgClient struct {
	pool   *pgxpool.Pool
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewPgClient creates a new PostgreSQL client for message queuing.
func NewPgClient(pubServer string, pubPort int, subServer string, subPort int) EventMQClient {
	// For simplicity, we use the pubServer/pubPort for connection.
	pgUser := os.Getenv("PG_USER")
	if pgUser == "" {
		pgUser = "root"
	}
	pgPassword := os.Getenv("PG_PASSWORD")
	if pgPassword == "" {
		pgPassword = "admin123"
	}
	// No default password for security
	pgDB := os.Getenv("PG_DATABASE")
	if pgDB == "" {
		pgDB = "xtext"
	}
	sslMode := os.Getenv("PG_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		pgUser, pgPassword, parseHost(pubServer), pubPort, pgDB, sslMode)

	ctx, cancel := context.WithCancel(context.Background())

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		cancel()
		log.Fatalf("Unable to parse connection string: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		cancel()
		log.Fatalf("Unable to connect to database: %v", err)
	}

	// Initialize schema
	if _, err := pool.Exec(ctx, createQueueTableSQL); err != nil {
		cancel()
		pool.Close()
		log.Fatalf("Failed to create queue table: %v", err)
	}

	return &EventPgClient{
		pool:   pool,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Pub sends a message to the specified topic.
func (c *EventPgClient) Pub(topic string, msg []byte) error {
	tx, err := c.pool.Begin(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(c.ctx)

	// Insert the message into the queue table
	insertSQL := "INSERT INTO event_queue (topic, payload) VALUES ($1, $2)"
	_, err = tx.Exec(c.ctx, insertSQL, topic, msg)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}

	// Notify listeners on the topic channel
	notifySQL := fmt.Sprintf("NOTIFY %s", pgx.Identifier{topic}.Sanitize())
	_, err = tx.Exec(c.ctx, notifySQL)
	if err != nil {
		return fmt.Errorf("failed to notify channel: %w", err)
	}

	return tx.Commit(c.ctx)
}

// Sub subscribes to messages for a specified topic.
func (c *EventPgClient) Sub(topic string, channel string, handler MessageHandler, pollIntervalMilliseconds int64, maxInFlight int) error {
	log.Printf("Starting PostgreSQL subscription for topic: %s, channel: %s\n", topic, channel)

	if maxInFlight < 1 {
		maxInFlight = 1
	}
	if pollIntervalMilliseconds < 100 {
		pollIntervalMilliseconds = 100 // Minimum poll interval
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.listenAndProcess(topic, handler, maxInFlight, time.Duration(pollIntervalMilliseconds)*time.Millisecond)
	}()

	return nil
}

func (c *EventPgClient) listenAndProcess(topic string, handler MessageHandler, maxInFlight int, pollInterval time.Duration) {
	conn, err := c.pool.Acquire(c.ctx)
	if err != nil {
		log.Printf("[PG MQ] Error acquiring connection for listener: %v", err)
		return
	}
	defer conn.Release()

	// Listen on the channel for notifications
	listenSQL := fmt.Sprintf("LISTEN %s", pgx.Identifier{topic}.Sanitize())
	_, err = conn.Exec(c.ctx, listenSQL)
	if err != nil {
		log.Printf("[PG MQ] Error listening to channel %s: %v", topic, err)
		return
	}

	semaphore := make(chan struct{}, maxInFlight)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("[PG MQ] Closing subscription for topic %s", topic)
			return
		case <-ticker.C:
			c.processAvailableJobs(topic, handler, semaphore, maxInFlight)
		default:
			notification := conn.Conn().PgConn().WaitForNotification(c.ctx)
			if notification != nil {
				c.processAvailableJobs(topic, handler, semaphore, maxInFlight)
			}
		}
	}
}

func (c *EventPgClient) processAvailableJobs(topic string, handler MessageHandler, semaphore chan struct{}, maxInFlight int) {
	// Try to fill the worker pool
	for i := 0; i < maxInFlight; i++ {
		select {
		case semaphore <- struct{}{}:
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				defer func() { <-semaphore }()

				job, err := c.fetchAndLockJob(topic)
				if err != nil {
					if err != pgx.ErrNoRows {
						log.Printf("[PG MQ] Error fetching job: %v", err)
					}
					return
				}

				if err := handler(job.Payload); err != nil {
					log.Printf("[PG MQ] Handler error for job %d: %v. Attempts: %d", job.ID, err, job.Attempts)
					c.releaseOrMarkFailed(job.ID, job.Attempts)
				} else {
					c.deleteJob(job.ID)
				}
			}()
		default:
			// All workers are busy
			return
		}
	}
}

type job struct {
	ID       int64
	Payload  []byte
	Attempts int
}

func (c *EventPgClient) fetchAndLockJob(topic string) (*job, error) {
	sql := `
        WITH next_job AS (
            SELECT id
            FROM event_queue
            WHERE topic = $1 AND status = 'available'
            ORDER BY created_at
            LIMIT 1
            FOR UPDATE SKIP LOCKED
        )
        UPDATE event_queue
        SET status = 'locked',
            locked_until = NOW() + INTERVAL '5 minutes',
            processing_attempts = processing_attempts + 1
        WHERE id = (SELECT id FROM next_job)
        RETURNING id, payload, processing_attempts;
    `
	j := &job{}
	err := c.pool.QueryRow(c.ctx, sql, topic).Scan(&j.ID, &j.Payload, &j.Attempts)
	return j, err
}

func (c *EventPgClient) deleteJob(jobID int64) {
	sql := "DELETE FROM event_queue WHERE id = $1"
	_, err := c.pool.Exec(c.ctx, sql, jobID)
	if err != nil {
		log.Printf("[PG MQ] Error deleting job %d: %v", jobID, err)
	}
}

func (c *EventPgClient) releaseOrMarkFailed(jobID int64, attempts int) {
	var sql string
	maxAttempts := 5 // configurable
	if attempts >= maxAttempts {
		sql = "UPDATE event_queue SET status = 'failed' WHERE id = $1"
	} else {
		sql = "UPDATE event_queue SET status = 'available', locked_until = NULL WHERE id = $1"
	}

	_, err := c.pool.Exec(c.ctx, sql, jobID)
	if err != nil {
		log.Printf("[PG MQ] Error updating job %d status: %v", jobID, err)
	}
}

// Close closes the PostgreSQL connection pool.
func (c *EventPgClient) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	if c.pool != nil {
		c.pool.Close()
	}
}
