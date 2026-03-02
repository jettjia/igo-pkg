package xasynq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// TaskHandler is a generic interface for handling a task with a specific payload type.
type TaskHandler[T any] interface {
	Handle(ctx context.Context, payload T) error
}

// TaskHandleFunc is a function adapter for TaskHandler.
type TaskHandleFunc[T any] func(ctx context.Context, payload T) error

// Handle calls the wrapped function.
func (f TaskHandleFunc[T]) Handle(ctx context.Context, payload T) error {
	return f(ctx, payload)
}

// NewTask creates a new asynq.Task with a typed payload.
// It automatically marshals the payload into JSON.
func NewTask[P any](typeName string, payload P, opts ...asynq.Option) (*asynq.Task, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task payload: %w", err)
	}
	return asynq.NewTask(typeName, payloadBytes, opts...), nil
}

// HandleFunc creates a TaskHandler from a function.
func HandleFunc[T any](f func(ctx context.Context, payloadType T) error) TaskHandler[T] {
	return TaskHandleFunc[T](f)
}

// RegisterTaskHandler registers a generic, type-safe task handler with the server's mux.
// It automatically unmarshals the JSON payload into the specified type.
func RegisterTaskHandler[T any](mux *asynq.ServeMux, typeName string, handler TaskHandler[T]) {
	mux.HandleFunc(typeName, func(ctx context.Context, t *asynq.Task) error {
		var payload T
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal task payload: %w", err)
		}
		return handler.Handle(ctx, payload)
	})
}

// --- Task Option Helpers ---

// WithMaxRetry specifies the max number of times the task will be retried.
func WithMaxRetry(maxRetry int) asynq.Option {
	return asynq.MaxRetry(maxRetry)
}

// WithTimeout specifies the timeout duration for the task.
func WithTimeout(timeout time.Duration) asynq.Option {
	return asynq.Timeout(timeout)
}

// WithDeadline specifies the deadline for the task.
func WithDeadline(t time.Time) asynq.Option {
	return asynq.Deadline(t)
}

// WithTaskID specifies the ID for the task, if another task with the same ID already exists, it will be rejeceted.
func WithTaskID(id string) asynq.Option {
	return asynq.TaskID(id)
}

// WithQueue specifies which queue the task should be sent to.
func WithQueue(name string) asynq.Option {
	return asynq.Queue(name)
}
