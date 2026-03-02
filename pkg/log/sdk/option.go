package sdk

import "github.com/jettjia/go-pkg/pkg/log/enum"

type Option func(p *Config)

func WithLogPath(logPath string) Option {
	return func(c *Config) {
		c.LogFilePath = logPath
	}
}

func WithLogName(LogName string) Option {
	return func(c *Config) {
		c.LogName = LogName
	}
}

func WithMaxSize(MaxSize int) Option {
	return func(c *Config) {
		c.MaxSize = MaxSize
	}
}

func WithMaxBackup(MaxBackup int) Option {
	return func(c *Config) {
		c.MaxBackup = MaxBackup
	}
}

func WithMaxAge(MaxAge int) Option {
	return func(c *Config) {
		c.MaxAge = MaxAge
	}
}

func WithLogLevel(LogLevel string) Option {
	return func(c *Config) {
		c.LogLevel = LogLevel
	}
}

func WithLogSendType(LogSendType enum.EnumType) Option {
	return func(c *Config) {
		c.LogSendType = LogSendType
	}
}
