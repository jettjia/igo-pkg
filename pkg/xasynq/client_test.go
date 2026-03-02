package xasynq

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func runProducer(client *Client) error {
	userPayload1 := &EmailPayload{UserID: 101, Message: "Critical Update"}
	_, info1, err := client.EnqueueNow(TypeEmailSend, userPayload1, WithQueue("critical"), WithMaxRetry(5), asynq.Retention(60*time.Second))
	if err != nil {
		return err
	}
	log.Printf("enqueued task: type=%s, id=%s queue=%s", TypeEmailSend, info1.ID, info1.Queue)

	userPayload2 := &SMSPayload{UserID: 202, Message: "Weekly Newsletter"}
	_, info2, err := client.EnqueueIn(time.Second*5, TypeSMSSend, userPayload2, WithQueue("default"), WithMaxRetry(3), asynq.Retention(60*time.Second))
	if err != nil {
		return err
	}
	log.Printf("enqueued task: type=%s, id=%s queue=%s", TypeSMSSend, info2.ID, info2.Queue)
	cancelTask("default", info2.ID, true) // cancel task will succeed

	userPayload3 := &MsgNotificationPayload{UserID: 303, Message: "Promotional Offer"}
	_, info3, err := client.EnqueueAt(time.Now().Add(time.Second*10), TypeMsgNotification, userPayload3, WithQueue("low"), WithMaxRetry(1), asynq.Retention(60*time.Second))
	if err != nil {
		return err
	}
	log.Printf("enqueued task: type=%s, id=%s queue=%s", TypeMsgNotification, info3.ID, info3.Queue)
	cancelTask("low", info3.ID, true) // cancel task will succeed

	userPayload4 := &UniqueTaskPayload{UserID: 404, Message: "unique task"}
	_, info4, err := client.EnqueueUnique(time.Minute, TypeUniqueTask, userPayload4, WithQueue("default"), WithMaxRetry(2))
	if err != nil {
		return err
	}
	log.Printf("enqueued task: type=%s, id=%s queue=%s", TypeUniqueTask, info4.ID, info4.Queue)
	_, _, err = client.EnqueueUnique(time.Minute, TypeUniqueTask, userPayload4, WithQueue("default"), WithMaxRetry(2))
	if err != nil {
		log.Printf("triggered duplicate task error:%v", err)
	}

	// Equivalent EnqueueNow function
	userPayload5 := &EmailPayload{UserID: 505, Message: "Important Notification"}
	task, err := NewTask(TypeEmailSend, userPayload5)
	if err != nil {
		return err
	}
	info5, err := client.Enqueue(task, WithQueue("low"), WithMaxRetry(3), WithDeadline(time.Now().Add(time.Second*15)), WithTaskID(fmt.Sprintf("unique-%d", rand.Int63n(1e10))), asynq.Retention(60*time.Second))
	if err != nil {
		return err
	}
	log.Printf("enqueued task: type=%s, id=%s queue=%s", TypeEmailSend, info5.ID, info5.Queue)

	return nil
}

func cancelTask(queue string, taskID string, isScheduled bool) {
	fmt.Println()
	defer fmt.Println()
	time.Sleep(time.Second)

	inspector := NewInspector(getRedisConfig())

	info, err := inspector.GetTaskInfo(queue, taskID)
	if err != nil {
		log.Printf("get task info failed: %s, queue=%s, taskID=%s", err, queue, taskID)
		return
	}
	log.Printf("task status: type=%s, id=%s queue=%s, status=%s", info.Type, info.ID, info.Queue, info.State.String())
	if info.State == asynq.TaskStateCompleted {
		return
	}

	time.Sleep(time.Millisecond * 100)
	if isScheduled {
		err = inspector.CancelTask(queue, info.ID)
	} else {
		err = inspector.CancelTask("", info.ID) // queue is empty string for non-scheduled tasks
	}

	if err != nil {
		log.Printf("cancel task failed: %s, queue=%s, taskID=%s", err, queue, info.ID)
		return
	}
	log.Printf("cancel task succeeded: type=%s, id=%s queue=%s", info.Type, info.ID, info.Queue)

	time.Sleep(time.Millisecond * 100)
	info2, err := inspector.GetTaskInfo(queue, info.ID)
	if err != nil {
		log.Printf("get task info after cancel failed: %s, queue=%s, taskID=%s", err, queue, info.ID)
		return
	}
	log.Printf("get task status after cancel: type=%s, id=%s queue=%s, status=%s", info2.Type, info2.ID, info2.Queue, info2.State.String())
}

func TestProducer(t *testing.T) {
	client := NewClient(getRedisConfig())

	err := runProducer(client)
	if err != nil {
		t.Log("run producer failed:", err)
		return
	}
	defer client.Close()

	log.Println("all tasks enqueued")
}
