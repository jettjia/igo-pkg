package event

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type EventRedisClient struct {
	client    *redis.Client
	ctx       context.Context
	pubServer string
	pubPort   int
	subServer string
	subPort   int
}

// NewRedisClient create a redis client
func NewRedisClient(pubServer string, pubPort int, subServer string, subPort int) EventMQClient {
	redisPwd := os.Getenv("REDIS_PWD")
	if redisPwd == "" {
		redisPwd = "admin123"
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", parseHost(pubServer), pubPort),
		Password:     redisPwd,
		DB:           9,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	return &EventRedisClient{
		client:    client,
		ctx:       context.Background(),
		pubServer: pubServer,
		pubPort:   pubPort,
		subServer: subServer,
		subPort:   subPort,
	}
}

// Pub send a message to the specified topic
func (c *EventRedisClient) Pub(topic string, msg []byte) error {
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		result := c.client.Publish(c.ctx, topic, msg)
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

// Sub subscribe to messages for a specified topic
func (c *EventRedisClient) Sub(topic string, channel string, handler MessageHandler, pollIntervalMilliseconds int64, maxInFlight int) error {
	log.Printf("Starting Redis subscription for topic: %s, channel: %s\n", topic, channel)

	if pollIntervalMilliseconds < 1 {
		pollIntervalMilliseconds = 1
	}
	if pollIntervalMilliseconds > 1000 {
		pollIntervalMilliseconds = 1000
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	concurrency := maxInFlight
	if concurrency > 256 {
		concurrency = 256
	}
	if concurrency < 1 {
		concurrency = 1
	}

	return c.createSubscription(topic, channel, handler, pollIntervalMilliseconds, concurrency, sigChan)
}

func (c *EventRedisClient) createSubscription(topic, channel string, handler MessageHandler, pollInterval int64, concurrency int, sigChan chan os.Signal) error {
	for {
		select {
		case <-sigChan:
			log.Println("Received shutdown signal, closing subscription...")
			return nil
		default:
			if err := c.runSubscription(topic, channel, handler, pollInterval, concurrency); err != nil {
				log.Printf("Subscription error: %v, retrying...\n", err)
				time.Sleep(time.Second)
				continue
			}
		}
	}
}

func (c *EventRedisClient) runSubscription(topic, channel string, handler MessageHandler, pollInterval int64, concurrency int) error {
	// create a subscription
	pubsub := c.client.Subscribe(c.ctx, topic)
	defer pubsub.Close()

	// verify the subscription connection
	if _, err := pubsub.Receive(c.ctx); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// get the message channel
	ch := pubsub.Channel()

	// create a work pool channel
	workChan := make(chan *redis.Message, concurrency)
	defer close(workChan)

	// start the work pool
	for i := 0; i < concurrency; i++ {
		go c.worker(i, workChan, handler)
	}

	// message processing loops
	for msg := range ch {
		// send a message to the working pool
		select {
		case workChan <- msg:
			// the message has been sent to the working pool
		default:
			// the working pool is full wait for some time and try again
			time.Sleep(time.Duration(pollInterval) * time.Millisecond)
			workChan <- msg
		}
	}

	return nil
}

func (c *EventRedisClient) worker(id int, msgChan <-chan *redis.Message, handler MessageHandler) {
	for msg := range msgChan {
		// create a timeout context for message processing
		ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)

		// create a completion channel
		done := make(chan struct{})

		go func() {
			defer close(done)
			if err := handler([]byte(msg.Payload)); err != nil {
				log.Printf("Worker %d failed to process message: %v\n", id, err)
			}
		}()

		// wait for the processing to complete or time out
		select {
		case <-done:
			// processing is complete
		case <-ctx.Done():
			log.Printf("Worker %d: message processing timed out\n", id)
		}

		cancel()
	}
}

// Close close the redis connection
func (c *EventRedisClient) Close() {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v\n", err)
		}
	}
}
