package xasynq

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

var (
	// Print payload max length
	defaultMaxLength = 300
	// default zap log
	defaultLogger, _ = zap.NewProduction()
)

// LoggerOption set options.
type LoggerOption func(*loggerOptions)

type loggerOptions struct {
	logger    *zap.Logger
	zapSkip   int // default is 2
	maxLength int // default is 300
}

func (o *loggerOptions) apply(opts ...LoggerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func defaultLoggerOptions() *loggerOptions {
	return &loggerOptions{
		logger:    defaultLogger,
		maxLength: defaultMaxLength,
		zapSkip:   2,
	}
}

// WithLogger sets the logger to use for logging.
func WithLogger(l *zap.Logger) LoggerOption {
	return func(o *loggerOptions) {
		if l != nil {
			o.logger = l
		}
	}
}

// WithMaxLength sets the maximum length of the payload to log.
func WithMaxLength(l int) LoggerOption {
	return func(o *loggerOptions) {
		if l > 0 {
			o.maxLength = l
		}
	}
}

// WithZapSkip sets the number of callers to skip when logging.
func WithZapSkip(s int) LoggerOption {
	return func(o *loggerOptions) {
		if s >= 0 {
			o.zapSkip = s
		}
	}
}

// LoggingMiddleware logs information about each processed task.
func LoggingMiddleware(opts ...LoggerOption) func(next asynq.Handler) asynq.Handler {
	o := defaultLoggerOptions()
	o.apply(opts...)

	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			start := time.Now()

			id := getTaskID(ctx)
			o.logger.Info("[asynq] <<<< starting task",
				zap.String("type", t.Type()),
				zap.String("id", id),
				getPayload(t, o.maxLength),
			)
			err := next.ProcessTask(ctx, t)
			if err != nil {
				o.logger.Error("[asynq] >>>> task failed",
					zap.Error(err),
					zap.String("task_type", t.Type()),
					zap.String("task_id", getTaskID(ctx)),
				)
				return err
			}
			o.logger.Info("[asynq] >>>> task completed successfully",
				zap.String("type", t.Type()),
				zap.String("id", id),
				zap.Int64("time_us", time.Since(start).Microseconds()),
			)
			return nil
		})
	}
}

func getTaskID(ctx context.Context) string {
	id, ok := asynq.GetTaskID(ctx)
	if !ok {
		id = "unknown"
	}
	return id
}

func getPayload(t *asynq.Task, maxLength int) zap.Field {
	payloadField := zap.Skip()
	sizeField := len(t.Payload())
	if sizeField > 0 {
		if sizeField > maxLength {
			payloadField = zap.String("payload", string(t.Payload()[:maxLength])+" ...... ")
		} else {
			payloadField = zap.String("payload", string(t.Payload()))
		}
	}
	return payloadField
}

// ------------------------------------------------------------------------------------------

type ZapLogger struct {
	zLog *zap.Logger
}

func NewZapLogger(l *zap.Logger, skip int) asynq.Logger {
	zLog := l.WithOptions(zap.AddCallerSkip(skip)).With(zap.String("asynq", "true"))
	return &ZapLogger{
		zLog: zLog,
	}
}

func (l *ZapLogger) Debug(args ...interface{}) {
	l.zLog.Debug(fmt.Sprint(args...))
}

func (l *ZapLogger) Info(args ...interface{}) {
	l.zLog.Info(fmt.Sprint(args...))
}

func (l *ZapLogger) Warn(args ...interface{}) {
	l.zLog.Warn(fmt.Sprint(args...))
}

func (l *ZapLogger) Error(args ...interface{}) {
	l.zLog.Error(fmt.Sprint(args...))
}

func (l *ZapLogger) Fatal(args ...interface{}) {
	l.zLog.Fatal(fmt.Sprint(args...))
}
