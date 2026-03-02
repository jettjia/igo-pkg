package event

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Message
type Message struct {
	Data      []byte
	Timestamp time.Time
}

// MemoryConfig In-Memory Queue Configuration
type MemoryConfig struct {
	MaxQueueSize    int           // Maximum Queue Size per Topic
	MaxInFlight     int           // Maximum Concurrent Processing per Subscriber
	MessageTTL      time.Duration // Message Time-to-Live (TTL)
	CleanupInterval time.Duration // Cleanup Interval
	MaxMemoryMB     int64         // Maximum Memory Usage（MB）
}

// EventMemoryClient In-Memory Message Queue Client
type EventMemoryClient struct {
	subscribers   map[string][]chan Message
	messageQueues map[string]chan Message
	mu            sync.RWMutex
	done          chan struct{}
	closeOnce     sync.Once
	config        MemoryConfig
	stats         struct {
		messageCount  int64
		totalMemoryMB int64
		lastCleanup   time.Time
	}
}

// Default Configuration
var defaultConfig = MemoryConfig{
	MaxQueueSize:    10000,
	MaxInFlight:     100,
	MessageTTL:      time.Hour,
	CleanupInterval: time.Minute * 5,
	MaxMemoryMB:     1024, // 1GB
}

var (
	memoryClient *EventMemoryClient
	memoryOnce   sync.Once
)

// NewMemoryClient Create an in-memory message queue client
func NewMemoryClient(pubServer string, pubPort int, subServer string, subPort int) EventMQClient {
	return NewMemoryClientWithConfig(defaultConfig)
}

// NewMemoryClientWithConfig Create a client with custom configuration
func NewMemoryClientWithConfig(config MemoryConfig) EventMQClient {
	memoryOnce.Do(func() {
		memoryClient = &EventMemoryClient{
			subscribers:   make(map[string][]chan Message),
			messageQueues: make(map[string]chan Message),
			done:          make(chan struct{}),
			config:        config,
		}
		// Start Memory Monitoring
		go memoryClient.monitorMemoryUsage()
	})
	return memoryClient
}

// Pub publish a message to a topic
func (c *EventMemoryClient) Pub(topic string, msg []byte) error {
	message := Message{
		Data:      msg,
		Timestamp: time.Now(),
	}

	c.mu.Lock()
	// Ensure the message queue for the topic exists
	if _, exists := c.messageQueues[topic]; !exists {
		c.messageQueues[topic] = make(chan Message, c.config.MaxQueueSize)
		// Start Cleanup Goroutine
		go c.cleanExpiredMessages(topic)
	}
	c.mu.Unlock()

	// Check if it has been closed
	select {
	case <-c.done:
		return fmt.Errorf("client is closed")
	default:
		// Attempt to send a message
		select {
		case c.messageQueues[topic] <- message:
			atomic.AddInt64(&c.stats.messageCount, 1)
			return nil
		default:
			return fmt.Errorf("message queue for topic %s is full (max size: %d)", topic, c.config.MaxQueueSize)
		}
	}
}

// Sub subscribe to a specified topic
func (c *EventMemoryClient) Sub(topic string, channel string, handler MessageHandler, pollIntervalMilliseconds int64, maxInFlight int) error {
	// limit the number of concurrent processes
	if maxInFlight <= 0 || maxInFlight > c.config.MaxInFlight {
		maxInFlight = c.config.MaxInFlight
	}

	subscriber := make(chan Message, maxInFlight)

	c.mu.Lock()
	if _, exists := c.messageQueues[topic]; !exists {
		c.messageQueues[topic] = make(chan Message, c.config.MaxQueueSize)
		go c.cleanExpiredMessages(topic)
	}
	c.subscribers[topic] = append(c.subscribers[topic], subscriber)
	c.mu.Unlock()

	// start message distribution
	go c.dispatchMessages(topic)

	// create a work pool
	workers := make(chan struct{}, maxInFlight)
	for i := 0; i < maxInFlight; i++ {
		workers <- struct{}{}
	}

	log.Printf("Started subscription for topic: %s, channel: %s\n", topic, channel)

	// process messages
	for {
		select {
		case <-c.done:
			return nil
		case msg, ok := <-subscriber:
			if !ok {
				return nil
			}
			// check if the message is expired
			if time.Since(msg.Timestamp) > c.config.MessageTTL {
				continue
			}
			// get a work pool token
			<-workers
			go func(message Message) {
				defer func() { workers <- struct{}{} }()
				if err := handler(message.Data); err != nil {
					log.Printf("Error processing message: %v\n", err)
				}
			}(msg)
		}
	}
}

// dispatchMessages distribute messages to all subscribers
func (c *EventMemoryClient) dispatchMessages(topic string) {
	for {
		select {
		case <-c.done:
			return
		default:
			c.mu.RLock()
			queue, exists := c.messageQueues[topic]
			if !exists {
				c.mu.RUnlock()
				return
			}
			subscribers := c.subscribers[topic]
			c.mu.RUnlock()

			select {
			case msg, ok := <-queue:
				if !ok {
					return
				}
				// check if the message is expired
				if time.Since(msg.Timestamp) > c.config.MessageTTL {
					continue
				}
				// send the message to all subscribers
				for _, subscriber := range subscribers {
					select {
					case subscriber <- msg:
						// the message was sent successfully
					default:
						log.Printf("Subscriber buffer full, message dropped for topic %s\n", topic)
					}
				}
			case <-c.done:
				return
			}
		}
	}
}

// cleanExpiredMessages clean up expired messages
func (c *EventMemoryClient) cleanExpiredMessages(topic string) {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			if queue, exists := c.messageQueues[topic]; exists {
				newQueue := make(chan Message, c.config.MaxQueueSize)
				close(queue)

				// only non expired messages are kept
				for msg := range queue {
					if time.Since(msg.Timestamp) < c.config.MessageTTL {
						newQueue <- msg
					} else {
						atomic.AddInt64(&c.stats.messageCount, -1)
					}
				}
				c.messageQueues[topic] = newQueue
			}
			c.mu.Unlock()
			c.stats.lastCleanup = time.Now()
		}
	}
}

// monitorMemoryUsage monitor memory usage
func (c *EventMemoryClient) monitorMemoryUsage() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			currentMemoryMB := int64(m.Alloc / 1024 / 1024)
			atomic.StoreInt64(&c.stats.totalMemoryMB, currentMemoryMB)

			// if the memory usage exceeds the limit a cleanup is triggered
			if currentMemoryMB > c.config.MaxMemoryMB {
				log.Printf("Memory usage high (%dMB), triggering cleanup\n", currentMemoryMB)
				c.cleanupOldMessages()
			}
		}
	}
}

// cleanupOldMessages clean up old messages
func (c *EventMemoryClient) cleanupOldMessages() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for topic, queue := range c.messageQueues {
		// if the queue is too large clean up half of the messages
		if len(queue) > c.config.MaxQueueSize/2 {
			newQueue := make(chan Message, c.config.MaxQueueSize)
			close(queue)

			count := 0
			for msg := range queue {
				if count > len(queue)/2 {
					newQueue <- msg
				} else {
					atomic.AddInt64(&c.stats.messageCount, -1)
				}
				count++
			}
			c.messageQueues[topic] = newQueue
		}
	}
}

// Close close the client
func (c *EventMemoryClient) Close() {
	c.closeOnce.Do(func() {
		close(c.done)

		c.mu.Lock()
		defer c.mu.Unlock()

		// close all message queues
		for topic, queue := range c.messageQueues {
			close(queue)
			delete(c.messageQueues, topic)
		}

		// turn off all subscribers
		for topic, subscribers := range c.subscribers {
			for _, subscriber := range subscribers {
				close(subscriber)
			}
			delete(c.subscribers, topic)
		}

		log.Printf("Memory client closed. Final stats: Messages: %d, Memory: %dMB\n",
			atomic.LoadInt64(&c.stats.messageCount),
			atomic.LoadInt64(&c.stats.totalMemoryMB))
	})
}

// GetStats get statistics
func (c *EventMemoryClient) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"message_count":   atomic.LoadInt64(&c.stats.messageCount),
		"memory_usage_mb": atomic.LoadInt64(&c.stats.totalMemoryMB),
		"last_cleanup":    c.stats.lastCleanup,
	}
}
