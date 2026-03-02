package xasynq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func getRedisConfig() RedisConfig {
	return RedisConfig{
		Addr: "localhost:6379",
		//Password: "123456",
	}
}

const (
	TypeEmailSend       = "email:send"
	TypeSMSSend         = "sms:send"
	TypeMsgNotification = "msg:notification"
	TypeUniqueTask      = "unique:task"
)

type EmailPayload struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

// handleEmailTask demonstrates a task handler that can fail and will be retried.
func handleEmailTask(ctx context.Context, p *EmailPayload) error {
	log.Printf(" [Email] Task for UserID %d completed successfully", p.UserID)
	return nil
}

type UniqueTaskPayload struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func handleUniqueTask(ctx context.Context, p *UniqueTaskPayload) error {
	log.Printf(" [Unique] Task for UserID %d completed successfully", p.UserID)
	return nil
}

type SMSPayload struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func (p *SMSPayload) ProcessTask(ctx context.Context, t *asynq.Task) error {
	if err := json.Unmarshal(t.Payload(), p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	log.Printf("[SMS] Task for UserID %d completed successfully", p.UserID)
	return nil
}

type MsgNotificationPayload struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func handleMsgNotificationTask(ctx context.Context, t *asynq.Task) error {
	var p MsgNotificationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	log.Printf("[MSG] Task for UserID %d completed successfully", p.UserID)
	return nil
}

func runConsumer(redisCfg RedisConfig) (*Server, error) {
	serverCfg := DefaultServerConfig(WithLogger(nil)) // Uses critical, default, low queues
	srv := NewServer(redisCfg, serverCfg)
	srv.Use(LoggingMiddleware(WithLogger(nil), WithMaxLength(200)))

	// register task handle function, there are three registration methods available
	RegisterTaskHandler(srv.Mux(), TypeEmailSend, HandleFunc(handleEmailTask)) // Method 1: use HandleFunc (Recommendation)
	srv.Register(TypeSMSSend, &SMSPayload{})                                   // Method 2: use struct as payload
	srv.RegisterFunc(TypeMsgNotification, handleMsgNotificationTask)           // Method 3: use function as handler

	RegisterTaskHandler(srv.Mux(), TypeUniqueTask, HandleFunc(handleUniqueTask))

	srv.Run()

	return srv, nil
}

func TestConsumer(t *testing.T) {
	srv, err := runConsumer(getRedisConfig())
	if err != nil {
		t.Log("run consumer error:", err)
		return
	}
	time.Sleep(10 * time.Second)
	srv.Shutdown()
}
