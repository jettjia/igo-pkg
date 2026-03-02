package v1

import (
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/jettjia/go-pkg/pkg/log/enum"
	"github.com/jettjia/go-pkg/pkg/log/sdk"
)

var (
	once sync.Once

	l *logrus.Logger
)

// DoLogger get a log handle
func (lcfg *Client) NewLogger(options ...func(*sdk.Config)) *logrus.Logger {
	once.Do(func() {
		cfg := sdk.Config{
			LogFilePath: "/var/log/",
			MaxSize:     128,
			MaxBackup:   255,
			MaxAge:      7,
			LogLevel:    "error",
			LogSendType: enum.EnumType(enum.CONSOLE),
		}
		for _, option := range options {
			option(&cfg)
		}
		l = lcfg.initLogger(&cfg)
	})

	return l
}

func (lcfg *Client) initLogger(cfg *sdk.Config) *logrus.Logger {

	logHandle := logrus.New()
	logHandle.SetLevel(logLevel(cfg.LogLevel))
	logHandle.SetFormatter(&logrus.JSONFormatter{})

	var output io.Writer
	var sendType int
	sendType = cfg.LogSendType.Index()
	switch sendType {
	case enum.CONSOLE.Index():
		output = os.Stdout
		logHandle.SetOutput(output)
		return logHandle
	case enum.FILE.Index():
		output = &lumberjack.Logger{
			Filename:   logFileNamePath(cfg.LogFilePath),
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackup,
			MaxAge:     cfg.MaxAge,
			Compress:   true,
		}
		logHandle.SetOutput(output)
		return logHandle
	case enum.MQ.Index():
		mqHook, err := newMqHook(lcfg.Config.LogName, lcfg.MQ.MqType, lcfg.MQ.ProducerHost, lcfg.MQ.ProducerPort)
		if err == nil {
			logHandle.Hooks.Add(mqHook)
		}
		return logHandle
	case enum.ES.Index():
		esh, err := newEsHook(lcfg.Config.LogName, lcfg.ConfigES.EsAddrs, lcfg.ConfigES.EsUser, lcfg.ConfigES.EsPassword)
		if err == nil {
			logHandle.Hooks.Add(esh)
		}
		return logHandle
	case enum.ZS.Index():
		zsh, err := newZsHook(lcfg.Config.LogName, lcfg.ConfigZS.ZsAddrs, lcfg.ConfigZS.ZsUser, lcfg.ConfigZS.ZsPassword)
		if err == nil {
			logHandle.Hooks.Add(zsh)
		}
		return logHandle
	case enum.OTEL.Index():
		otelHook, err := newOtelHook(lcfg.Config.LogName, lcfg.ConfigOTEL.Endpoint, lcfg.ConfigOTEL.User, lcfg.ConfigOTEL.Pwd)
		if err != nil {
			logHandle.SetOutput(output)
			return logHandle
		}
		logHandle.Hooks.Add(otelHook)
		return logHandle
	default:
		output = os.Stdout
		logHandle.SetOutput(output)
		return logHandle
	}
}

func logLevel(logLevel string) (level logrus.Level) {
	switch logLevel {
	case "panic":
		return logrus.PanicLevel
	case "fatal":
		return logrus.FatalLevel
	case "error":
		return logrus.ErrorLevel
	case "warn":
		return logrus.WarnLevel
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	case "trace":
		return logrus.TraceLevel
	default:
		return logrus.ErrorLevel
	}
}

func logFileNamePath(settingPath string) string {
	var (
		logFilePath string
	)
	logFilePath = settingPath
	if logFilePath == "" {
		logFilePath = "/var/log/"
	}

	if err := os.MkdirAll(logFilePath, 0o777); err != nil {
		panic(err)
	}

	// Set filename to date
	logFileName := time.Now().Format("2006-01-02") + ".log"
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			panic(err)
		}
	}

	return fileName
}
