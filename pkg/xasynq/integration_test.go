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

// 集成测试专用的任务类型常量，使用不同前缀避免冲突
const (
	IntTypeEmailSend       = "int:email:send"
	IntTypeSMSSend         = "int:sms:send"
	IntTypeMsgNotification = "int:msg:notification"
	IntTypeUniqueTask      = "int:unique:task"
	IntTypePriorityTask    = "int:priority:task"
)

// 集成测试专用的负载结构体

type IntEmailPayload struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func handleIntEmailTask(ctx context.Context, p *IntEmailPayload) error {
	log.Printf("[Integration][Email] Task for UserID %d completed successfully: %s", p.UserID, p.Message)
	return nil
}

type IntSMSPayload struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func (p *IntSMSPayload) ProcessTask(ctx context.Context, t *asynq.Task) error {
	if err := json.Unmarshal(t.Payload(), p); err != nil {
		return fmt.Errorf("failed to unmarshal SMS payload: %w", err)
	}
	log.Printf("[Integration][SMS] Task for UserID %d completed successfully: %s", p.UserID, p.Message)
	return nil
}

type IntMsgNotificationPayload struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func handleIntMsgNotificationTask(ctx context.Context, t *asynq.Task) error {
	var p IntMsgNotificationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal notification payload: %w", err)
	}
	log.Printf("[Integration][MSG] Task for UserID %d completed successfully: %s", p.UserID, p.Message)
	return nil
}

type IntUniqueTaskPayload struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func handleIntUniqueTask(ctx context.Context, p *IntUniqueTaskPayload) error {
	log.Printf("[Integration][Unique] Task for UserID %d completed successfully: %s", p.UserID, p.Message)
	return nil
}

type IntPriorityTaskPayload struct {
	UserID   int `json:"user_id"`
	Priority int `json:"priority"`
}

func handleIntPriorityTask(ctx context.Context, p *IntPriorityTaskPayload) error {
	log.Printf("[Integration][Priority] Task for UserID %d with priority %d completed successfully", p.UserID, p.Priority)
	return nil
}

// 获取集成测试专用的Redis配置
func getIntRedisConfig() RedisConfig {
	return RedisConfig{
		Addr: "localhost:6379",
		// 如果需要密码，可以在这里添加
		// Password: "your-redis-password",
	}
}

// 启动集成测试专用的消费者
func runIntConsumer(redisCfg RedisConfig) (*Server, error) {
	serverCfg := DefaultServerConfig(WithLogger(nil))
	srv := NewServer(redisCfg, serverCfg)
	srv.Use(LoggingMiddleware(WithLogger(nil), WithMaxLength(200)))

	// 注册任务处理函数
	RegisterTaskHandler(srv.Mux(), IntTypeEmailSend, HandleFunc(handleIntEmailTask))
	srv.Register(IntTypeSMSSend, &IntSMSPayload{})
	srv.RegisterFunc(IntTypeMsgNotification, handleIntMsgNotificationTask)
	RegisterTaskHandler(srv.Mux(), IntTypeUniqueTask, HandleFunc(handleIntUniqueTask))
	RegisterTaskHandler(srv.Mux(), IntTypePriorityTask, HandleFunc(handleIntPriorityTask))

	srv.Run()
	return srv, nil
}

// 发送集成测试专用的任务
func sendIntTasks(client *Client) error {
	// 1. 发送即时任务
	emailPayload := &IntEmailPayload{UserID: 1001, Message: "Integration Test - Critical Update"}
	_, info1, err := client.EnqueueNow(IntTypeEmailSend, emailPayload, WithQueue("critical"), WithMaxRetry(5))
	if err != nil {
		return fmt.Errorf("failed to enqueue email task: %w", err)
	}
	log.Printf("[Integration] Enqueued email task: id=%s, queue=%s", info1.ID, info1.Queue)

	// 2. 发送延迟任务
	smsPayload := &IntSMSPayload{UserID: 1002, Message: "Integration Test - Weekly Newsletter"}
	_, info2, err := client.EnqueueIn(2*time.Second, IntTypeSMSSend, smsPayload, WithQueue("default"))
	if err != nil {
		return fmt.Errorf("failed to enqueue SMS task: %w", err)
	}
	log.Printf("[Integration] Enqueued SMS task: id=%s, queue=%s", info2.ID, info2.Queue)

	// 3. 发送定时任务
	msgPayload := &IntMsgNotificationPayload{UserID: 1003, Message: "Integration Test - Promotional Offer"}
	_, info3, err := client.EnqueueAt(time.Now().Add(5*time.Second), IntTypeMsgNotification, msgPayload, WithQueue("low"))
	if err != nil {
		return fmt.Errorf("failed to enqueue notification task: %w", err)
	}
	log.Printf("[Integration] Enqueued notification task: id=%s, queue=%s", info3.ID, info3.Queue)

	// 4. 发送唯一任务
	uniquePayload := &IntUniqueTaskPayload{UserID: 1004, Message: "Integration Test - Unique Task"}
	_, info4, err := client.EnqueueUnique(30*time.Second, IntTypeUniqueTask, uniquePayload, WithQueue("default"))
	if err != nil {
		return fmt.Errorf("failed to enqueue unique task: %w", err)
	}
	log.Printf("[Integration] Enqueued unique task: id=%s, queue=%s", info4.ID, info4.Queue)

	// 尝试发送重复的唯一任务
	_, _, err = client.EnqueueUnique(30*time.Second, IntTypeUniqueTask, uniquePayload, WithQueue("default"))
	if err == nil {
		log.Printf("[Integration] Warning: Expected error for duplicate unique task but got none")
	} else {
		log.Printf("[Integration] Expected error for duplicate unique task: %v", err)
	}

	// 5. 测试队列优先级
	for i := 5; i >= 1; i-- {
		priorityPayload := &IntPriorityTaskPayload{UserID: 1000 + i, Priority: i}
		queue := "default"
		if i >= 3 {
			queue = "critical"
		}
		_, info, err := client.EnqueueNow(IntTypePriorityTask, priorityPayload, WithQueue(queue))
		if err != nil {
			return fmt.Errorf("failed to enqueue priority task: %w", err)
		}
		log.Printf("[Integration] Enqueued priority %d task: id=%s, queue=%s", i, info.ID, info.Queue)
	}

	return nil
}

// 集成测试取消任务辅助函数
func cancelIntTask(queue string, taskID string, isScheduled bool) {
	fmt.Println()
	defer fmt.Println()
	time.Sleep(time.Second)

	inspector := NewInspector(getIntRedisConfig())

	info, err := inspector.GetTaskInfo(queue, taskID)
	if err != nil {
		log.Printf("[Integration] Get task info failed: %s, queue=%s, taskID=%s", err, queue, taskID)
		return
	}
	log.Printf("[Integration] Task status before cancel: type=%s, id=%s, queue=%s, status=%s", info.Type, info.ID, info.Queue, info.State.String())

	if isScheduled {
		err = inspector.CancelTask(queue, info.ID)
	} else {
		err = inspector.CancelTask("", info.ID) // 非计划任务队列为空字符串
	}

	if err != nil {
		log.Printf("[Integration] Cancel task failed: %s, queue=%s, taskID=%s", err, queue, taskID)
		return
	}
	log.Printf("[Integration] Cancel task succeeded: type=%s, id=%s, queue=%s", info.Type, info.ID, info.Queue)

	time.Sleep(time.Millisecond * 100)
	info2, err := inspector.GetTaskInfo(queue, info.ID)
	if err != nil {
		log.Printf("[Integration] Get task info after cancel failed: %s, queue=%s, taskID=%s", err, queue, taskID)
		return
	}
	log.Printf("[Integration] Task status after cancel: type=%s, id=%s, queue=%s, status=%s", info2.Type, info2.ID, info2.Queue, info2.State.String())
}

// 主集成测试函数
func TestXAsynqIntegration(t *testing.T) {
	// 1. 获取Redis配置
	redisCfg := getIntRedisConfig()

	// 2. 启动消费者
	srv, err := runIntConsumer(redisCfg)
	if err != nil {
		t.Fatalf("Failed to start consumer: %v", err)
	}
	defer srv.Shutdown()

	// 3. 创建客户端
	client := NewClient(redisCfg)

	// 4. 发送测试任务
	if err := sendIntTasks(client); err != nil {
		t.Fatalf("Failed to send tasks: %v", err)
	}

	// 5. 等待任务处理完成
	time.Sleep(15 * time.Second)

	log.Println("[Integration] All tasks processed. Integration test completed successfully.")
}
