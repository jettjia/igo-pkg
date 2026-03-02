package event

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/cast"
	"golang.org/x/exp/rand"
)

func Test_NewEventMQClient_nsq_pub(t *testing.T) {
	client := NewMQClient("127.0.0.1", 4151, "127.0.0.1", 4161, "nsq")

	topic := "test-nsq"

	err := client.Pub(topic, []byte("i am nsq"))
	if err != nil {
		t.Error(err)
	}
}

func Test_NewEventMQClient_nsq_sub(t *testing.T) {
	client := NewMQClient("127.0.0.1", 4151, "127.0.0.1", 4161, "nsq")

	topic := "test-nsq"
	channel := "test"

	err := client.Sub(topic, channel, func(msg []byte) error {
		fmt.Println("sub.msg:", cast.ToString(msg))
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// go test -v -run Test_NewEventMQClient_redis_pub_sub ./
func Test_NewEventMQClient_redis_pub_sub(t *testing.T) {
	// set a redis password
	os.Setenv("REDIS_PWD", "admin123")

	topic := "test-redis"
	channel := "test-channel"

	// create subscribers
	subClient := NewMQClient("127.0.0.1", 6379, "127.0.0.1", 6379, "redis")

	// create a publisher
	pubClient := NewMQClient("127.0.0.1", 6379, "127.0.0.1", 6379, "redis")
	defer pubClient.Close()

	// it is used to count the number of messages received
	var receivedCount int32

	// start subscribers
	go func() {
		err := subClient.Sub(topic, channel, func(msg []byte) error {
			atomic.AddInt32(&receivedCount, 1)
			fmt.Printf("Received message %d: %s\n", atomic.LoadInt32(&receivedCount), string(msg))
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	}()

	// wait for the subscriber to start
	time.Sleep(time.Second)

	// start the publisher and send a message at random intervals
	go func() {
		for i := 1; i <= 10; i++ {
			msg := struct {
				ID      int       `json:"id"`
				Message string    `json:"message"`
				Time    time.Time `json:"time"`
				Random  float64   `json:"random"`
			}{
				ID:      i,
				Message: fmt.Sprintf("Test message %d", i),
				Time:    time.Now(),
				Random:  rand.Float64(),
			}

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				t.Errorf("Failed to marshal message: %v", err)
				continue
			}

			err = pubClient.Pub(topic, msgBytes)
			if err != nil {
				t.Errorf("Failed to publish message %d: %v", i, err)
			} else {
				fmt.Printf("Published message %d: %s\n", i, string(msgBytes))
			}

			sleepTime := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
			time.Sleep(sleepTime)
		}
	}()

	testDuration := 30 * time.Second

	timer := time.NewTimer(testDuration)

	// wait for the test to complete or time out
	<-timer.C
	fmt.Printf("\nTest completed. Total messages received: %d\n", atomic.LoadInt32(&receivedCount))
}

// go test -v -run Test_NewEventMQClient_redis_stream ./
func Test_NewEventMQClient_redis_stream(t *testing.T) {
	// create a client
	client := NewMQClient("127.0.0.1", 6379, "127.0.0.1", 6379, "redis-stream")
	defer client.Close()

	stream := "test-stream"
	group := "test-group"

	// it is used to count the number of messages received
	var receivedCount int32

	// kickstart the consumer
	go func() {
		fmt.Printf("Starting consumer for stream: %s, group: %s\n", stream, group)
		err := client.Sub(stream, group, func(msg []byte) error {
			count := atomic.AddInt32(&receivedCount, 1)
			fmt.Printf("Received message %d: %s\n", count, string(msg))
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(2 * time.Second)
	fmt.Println("Consumer started, beginning to publish messages...")

	for i := 1; i <= 5; i++ {
		msg := struct {
			ID      int       `json:"id"`
			Message string    `json:"message"`
			Time    time.Time `json:"time"`
			Random  string    `json:"random"`
		}{
			ID:      i,
			Message: fmt.Sprintf("test message %d", i),
			Time:    time.Now(),
			Random:  fmt.Sprintf("rand-%d", time.Now().UnixNano()),
		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal message: %v", err)
			continue
		}

		fmt.Printf("Publishing message %d...\n", i)
		err = client.Pub(stream, msgBytes)
		if err != nil {
			t.Errorf("Failed to publish message %d: %v", i, err)
		} else {
			fmt.Printf("Successfully published message %d: %s\n", i, string(msgBytes))
		}

		time.Sleep(time.Second)
	}

	time.Sleep(5 * time.Second)

	// output final statistics
	finalCount := atomic.LoadInt32(&receivedCount)
	fmt.Printf("\nTest completed. Total messages received: %d\n", finalCount)
	if finalCount != 5 {
		t.Errorf("Expected to receive 5 messages, but got %d", finalCount)
	}
}

// go test -v -run Test_NewEventMQClient_memory ./
func Test_NewEventMQClient_memory(t *testing.T) {
	client := NewMQClient("", 0, "", 0, "memory")
	defer client.Close()

	topic := "test-memory"
	channel := "test-channel"

	var receivedCount int32
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		fmt.Printf("Starting consumer for topic: %s, channel: %s\n", topic, channel)
		err := client.Sub(topic, channel, func(msg []byte) error {
			count := atomic.AddInt32(&receivedCount, 1)
			fmt.Printf("Received message %d: %s\n", count, string(msg))
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(time.Second)

	for i := 1; i <= 5; i++ {
		msg := struct {
			ID      int       `json:"id"`
			Message string    `json:"message"`
			Time    time.Time `json:"time"`
			Random  string    `json:"random"`
		}{
			ID:      i,
			Message: fmt.Sprintf("test message %d", i),
			Time:    time.Now(),
			Random:  fmt.Sprintf("rand-%d", time.Now().UnixNano()),
		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal message: %v", err)
			continue
		}

		err = client.Pub(topic, msgBytes)
		if err != nil {
			t.Errorf("Failed to publish message %d: %v", i, err)
		} else {
			fmt.Printf("Published message %d: %s\n", i, string(msgBytes))
		}

		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(time.Second)

	client.Close()

	wg.Wait()

	finalCount := atomic.LoadInt32(&receivedCount)
	fmt.Printf("\nTest completed. Total messages received: %d\n", finalCount)
	if finalCount != 5 {
		t.Errorf("Expected to receive 5 messages, but got %d", finalCount)
	}
}

// go test -v -run Test_NewEventMQClient_pgsql_pub_sub ./
func Test_NewEventMQClient_pgsql_pub_sub(t *testing.T) {
	// 你可以通过环境变量设置 PG 连接参数
	os.Setenv("PG_USER", "root")
	os.Setenv("PG_PASSWORD", "admin123")
	os.Setenv("PG_DATABASE", "xtext")
	os.Setenv("PG_SSLMODE", "disable")

	topic := "test-pgsql"
	channel := "test-channel"

	subClient := NewMQClient("127.0.0.1", 5432, "127.0.0.1", 5432, "pgsql")
	pubClient := NewMQClient("127.0.0.1", 5432, "127.0.0.1", 5432, "pgsql")
	defer pubClient.Close()

	var receivedCount int32

	go func() {
		err := subClient.Sub(topic, channel, func(msg []byte) error {
			atomic.AddInt32(&receivedCount, 1)
			fmt.Printf("PGSQL Received message %d: %s\n", atomic.LoadInt32(&receivedCount), string(msg))
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(time.Second)

	for i := 1; i <= 5; i++ {
		msg := struct {
			ID      int       `json:"id"`
			Message string    `json:"message"`
			Time    time.Time `json:"time"`
		}{
			ID:      i,
			Message: fmt.Sprintf("Test message %d", i),
			Time:    time.Now(),
		}
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal message: %v", err)
			continue
		}
		err = pubClient.Pub(topic, msgBytes)
		if err != nil {
			t.Errorf("Failed to publish message %d: %v", i, err)
		} else {
			fmt.Printf("PGSQL Published message %d: %s\n", i, string(msgBytes))
		}
		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(2 * time.Second)
	finalCount := atomic.LoadInt32(&receivedCount)
	fmt.Printf("\nPGSQL Test completed. Total messages received: %d\n", finalCount)
	if finalCount != 5 {
		t.Errorf("Expected to receive 5 messages, but got %d", finalCount)
	}
}
