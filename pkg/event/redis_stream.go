package event

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type EventRedisStreamClient struct {
	client    *redis.Client
	ctx       context.Context
	pubServer string
	pubPort   int
	subServer string
	subPort   int
}

// NewRedisStreamClient create a redis stream client
func NewRedisStreamClient(pubServer string, pubPort int, subServer string, subPort int) EventMQClient {
	redisPwd := os.Getenv("REDIS_PWD")
	if redisPwd == "" {
		redisPwd = "admin123"
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", parseHost(pubServer), pubPort),
		Password:     redisPwd,
		DB:           0,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	return &EventRedisStreamClient{
		client:    client,
		ctx:       context.Background(),
		pubServer: pubServer,
		pubPort:   pubPort,
		subServer: subServer,
		subPort:   subPort,
	}
}

// Pub send a message to the specified stream
func (c *EventRedisStreamClient) Pub(stream string, msg []byte) error {
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		// add a message to the stream
		result := c.client.XAdd(c.ctx, &redis.XAddArgs{
			Stream: stream,
			Values: map[string]interface{}{
				"data": msg,
				"time": time.Now().UnixNano(),
			},
		})

		if result.Err() == nil {
			return nil
		}

		log.Printf("Failed to publish message (attempt %d/%d): %v\n", i+1, maxRetries, result.Err())

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	return fmt.Errorf("failed to publish message after %d attempts", maxRetries)
}

// Sub subscribe to a specified stream
func (c *EventRedisStreamClient) Sub(stream string, group string, handler MessageHandler, pollIntervalMilliseconds int64, maxInFlight int) error {
	log.Printf("Starting Redis Stream subscription for stream: %s, group: %s\n", stream, group)

	if pollIntervalMilliseconds < 1 {
		pollIntervalMilliseconds = 1
	}
	if pollIntervalMilliseconds > 1000 {
		pollIntervalMilliseconds = 1000
	}

	// create a consumer group if it doesn t exist
	err := c.createConsumerGroup(stream, group)
	if err != nil {
		log.Printf("Warning: Consumer group creation: %v\n", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// create a work pool
	concurrency := maxInFlight
	if concurrency > 256 {
		concurrency = 256
	}
	if concurrency < 1 {
		concurrency = 1
	}

	// start the consumer
	return c.startConsumer(stream, group, handler, pollIntervalMilliseconds, concurrency, sigChan)
}

// createConsumerGroup create a consumer group
func (c *EventRedisStreamClient) createConsumerGroup(stream, group string) error {
	// try creating a stream first if it doesn t exist
	err := c.client.XGroupCreateMkStream(c.ctx, stream, group, "0").Err()
	if err != nil {
		// if the group already exists this is not an error
		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			return nil
		}
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	return nil
}

// startConsumer initiate consumer processing
func (c *EventRedisStreamClient) startConsumer(stream, group string, handler MessageHandler, pollInterval int64, concurrency int, sigChan chan os.Signal) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	// create a unique consumer id for each worker
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		consumerID := fmt.Sprintf("consumer-%d", i)

		go func(id string) {
			defer wg.Done()
			if err := c.consume(stream, group, id, handler, pollInterval); err != nil {
				select {
				case errChan <- err:
				default:
				}
			}
		}(consumerID)
	}

	// wait for a signal or error
	select {
	case <-sigChan:
		log.Println("Received shutdown signal, closing subscription...")
		return nil
	case err := <-errChan:
		return fmt.Errorf("consumer error: %w", err)
	}
}

// consume consume messages and delete processed messages immediately
func (c *EventRedisStreamClient) consume(stream, group, consumer string, handler MessageHandler, pollInterval int64) error {
	for {
		// read a new message
		streams, err := c.client.XReadGroup(c.ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"}, // ">" indicates that only new messages are read
			Count:    10,
			Block:    time.Duration(pollInterval) * time.Millisecond,
		}).Result()

		if err != nil {
			if err == redis.Nil { // no new news
				continue
			}
			// check if it s an error where the group doesn t exist
			if strings.Contains(err.Error(), "NOGROUP") {
				// try recreating the group
				if createErr := c.createConsumerGroup(stream, group); createErr != nil {
					log.Printf("Failed to recreate consumer group: %v\n", createErr)
				}
				continue
			}
			return fmt.Errorf("failed to read from stream: %w", err)
		}

		// process messages
		for _, stream := range streams {
			for _, message := range stream.Messages {
				// extract message data
				if data, ok := message.Values["data"].(string); ok {
					// process messages
					if err := handler([]byte(data)); err != nil {
						log.Printf("Error processing message %s: %v\n", message.ID, err)
						continue
					}

					// acknowledgment message
					if err := c.client.XAck(c.ctx, stream.Stream, group, message.ID).Err(); err != nil {
						log.Printf("Error acknowledging message %s: %v\n", message.ID, err)
					}

					// delete the processed message immediately
					if err := c.client.XDel(c.ctx, stream.Stream, message.ID).Err(); err != nil {
						log.Printf("Error deleting message %s: %v\n", message.ID, err)
					} else {
						log.Printf("Successfully deleted message %s\n", message.ID)
					}
				}
			}
		}
	}
}

// Close close the redis connection
func (c *EventRedisStreamClient) Close() {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v\n", err)
		}
	}
}

// CleanupStream clean up all messages in the specified stream
func (c *EventRedisStreamClient) CleanupStream(stream string) error {
	err := c.client.XTrimMaxLen(c.ctx, stream, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to cleanup stream %s: %w", stream, err)
	}
	return nil
}
