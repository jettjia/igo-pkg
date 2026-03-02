package xasynq

import (
	"fmt"

	"github.com/hibiken/asynq"
)

// SchedulerOption set options.
type SchedulerOption func(*schedulerOptions)

type schedulerOptions struct {
	schedulerOptions *asynq.SchedulerOpts
	logger           asynq.Logger
	loggerLevel      asynq.LogLevel
}

func (o *schedulerOptions) apply(opts ...SchedulerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func defaultSchedulerOptions() *schedulerOptions {
	return &schedulerOptions{
		loggerLevel: asynq.InfoLevel,
	}
}

// WithSchedulerOptions sets the options for the scheduler.
func WithSchedulerOptions(opts *asynq.SchedulerOpts) SchedulerOption {
	return func(o *schedulerOptions) {
		o.schedulerOptions = opts
	}
}

// WithSchedulerLogLevel sets the log level for the scheduler.
func WithSchedulerLogLevel(level asynq.LogLevel) SchedulerOption {
	return func(o *schedulerOptions) {
		o.loggerLevel = level
	}
}

// WithSchedulerLogger sets the logger for the scheduler.
func WithSchedulerLogger(opts ...LoggerOption) SchedulerOption {
	opt := defaultLoggerOptions()
	opt.apply(opts...)
	return func(o *schedulerOptions) {
		o.logger = NewZapLogger(opt.logger, opt.zapSkip)
	}
}

// --------------------------------------------------------------------

// Scheduler is a wrapper around asynq.Scheduler.
type Scheduler struct {
	*asynq.Scheduler
}

// NewScheduler creates a new periodic task scheduler.
func NewScheduler(cfg RedisConfig, opts ...SchedulerOption) *Scheduler {
	o := defaultSchedulerOptions()
	o.apply(opts...)
	if o.logger != nil {
		if o.schedulerOptions == nil {
			o.schedulerOptions = &asynq.SchedulerOpts{
				Logger:   o.logger,
				LogLevel: o.loggerLevel,
			}
		} else {
			o.schedulerOptions.Logger = o.logger
			o.schedulerOptions.LogLevel = o.loggerLevel
		}
	}

	return &Scheduler{
		Scheduler: asynq.NewScheduler(cfg.GetAsynqRedisConnOpt(), o.schedulerOptions),
	}
}

// Register adds a new periodic task.
func (s *Scheduler) Register(cronSpec string, task *asynq.Task, opts ...asynq.Option) (entryID string, err error) {
	return s.Scheduler.Register(cronSpec, task, opts...)
}

// RegisterTask adds a new periodic task with a given type name.
func (s *Scheduler) RegisterTask(cronSpec string, typeName string, payload any, opts ...asynq.Option) (entryID string, err error) {
	task, err := NewTask(typeName, payload)
	if err != nil {
		return "", err
	}
	return s.Scheduler.Register(cronSpec, task, opts...)
}

// Unregister removes a periodic task, cancel task execution.
func (s *Scheduler) Unregister(entryID string) error {
	return s.Scheduler.Unregister(entryID)
}

// Run runs the asynq Scheduler in a separate goroutine
func (s *Scheduler) Run() {
	go func() {
		if err := s.Scheduler.Run(); err != nil {
			panic(fmt.Sprintf("could not run asynq scheduler: %v", err))
		}
	}()
}

// Shutdown the Scheduler.
func (s *Scheduler) Shutdown() {
	if s == nil || s.Scheduler == nil {
		return
	}
	s.Scheduler.Shutdown()
}
