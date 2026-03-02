package xasynq

import (
	"time"

	"github.com/hibiken/asynq"
)

// Client is a wrapper around asynq.Client providing more convenient APIs.
type Client struct {
	*asynq.Client
}

// NewClient creates a new producer client.
func NewClient(cfg RedisConfig) *Client {
	return &Client{
		Client: asynq.NewClient(cfg.GetAsynqRedisConnOpt()),
	}
}

// NewFromClient creates a new producer client from an existing asynq.Client.
func NewFromClient(c *asynq.Client) *Client {
	return &Client{
		Client: c,
	}
}

// Enqueue enqueues the given task to a queue.
func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.Client.Enqueue(task, opts...)
}

// EnqueueNow enqueues a task for immediate processing, parameter payload should be supported json.Marshal
func (c *Client) EnqueueNow(typeName string, payload any, opts ...asynq.Option) (*asynq.Task, *asynq.TaskInfo, error) {
	task, err := NewTask(typeName, payload)
	if err != nil {
		return nil, nil, err
	}

	info, err := c.Client.Enqueue(task, opts...)
	return task, info, err
}

// EnqueueIn enqueues a task to be processed after a specified delay.
func (c *Client) EnqueueIn(delay time.Duration, typeName string, payload any, opts ...asynq.Option) (*asynq.Task, *asynq.TaskInfo, error) {
	task, err := NewTask(typeName, payload)
	if err != nil {
		return nil, nil, err
	}

	opts = append(opts, asynq.ProcessIn(delay))
	info, err := c.Client.Enqueue(task, opts...)
	return task, info, err
}

// EnqueueAt enqueues a task to be processed at a specific time.
func (c *Client) EnqueueAt(t time.Time, typeName string, payload any, opts ...asynq.Option) (*asynq.Task, *asynq.TaskInfo, error) {
	task, err := NewTask(typeName, payload)
	if err != nil {
		return nil, nil, err
	}

	opts = append(opts, asynq.ProcessAt(t))
	info, err := c.Client.Enqueue(task, opts...)
	return task, info, err
}

// EnqueueUnique enqueues a task with unique in the queue for a specified duration.
func (c *Client) EnqueueUnique(keepTime time.Duration, typeName string, payload any, opts ...asynq.Option) (*asynq.Task, *asynq.TaskInfo, error) {
	task, err := NewTask(typeName, payload)
	if err != nil {
		return nil, nil, err
	}
	opts = append(opts, asynq.Unique(keepTime))
	info, err := c.Client.Enqueue(task, opts...)
	return task, info, err
}
