package xasynq

import (
	"github.com/hibiken/asynq"
)

// Inspector provides access to the Redis backend used by asynq.
type Inspector struct {
	*asynq.Inspector
}

// NewInspector creates a new Inspector instance.
// Note: A new Redis connection will be created, in actual use, only once
func NewInspector(cfg RedisConfig) *Inspector {
	return &Inspector{asynq.NewInspector(cfg.GetAsynqRedisConnOpt())}
}

// CancelTask cancels the processing of a task.
func (i *Inspector) CancelTask(queue string, taskID string) error {
	if queue == "" {
		return i.Inspector.CancelProcessing(taskID)
	}

	return i.Inspector.DeleteTask(queue, taskID)
}

// GetTaskInfo returns information about a task.
func (i *Inspector) GetTaskInfo(queue string, taskID string) (*asynq.TaskInfo, error) {
	return i.Inspector.GetTaskInfo(queue, taskID)
}

// Close closes the inspector.
func (i *Inspector) Close() error {
	if i.Inspector == nil {
		return nil
	}
	return i.Inspector.Close()
}
