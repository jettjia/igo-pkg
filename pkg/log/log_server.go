package log

import (
	"github.com/sirupsen/logrus"

	"github.com/jettjia/igo-pkg/pkg/conf"
	"github.com/jettjia/igo-pkg/pkg/log/enum"
	"github.com/jettjia/igo-pkg/pkg/log/sdk"
	logv1 "github.com/jettjia/igo-pkg/pkg/log/service/log/v1"
)

func NewLoggerServer(conf *conf.Config) *logrus.Logger {
	config := &sdk.Config{
		LogFilePath: conf.Log.LogFileDir,
		LogName:     conf.Server.ServerName,
		MaxSize:     conf.Log.MaxSize,
		MaxBackup:   conf.Log.MaxBackups,
		MaxAge:      conf.Log.MaxAge,
		LogLevel:    conf.Log.LogLevel,
	}

	if conf.Log.LogOut == int(enum.CONSOLE) {
		logClient, err := logv1.NewClientCfg(config)
		if err != nil {
			goto defaultCli
		}

		return logClient.NewLogger(sdk.WithLogLevel(conf.Log.LogLevel), sdk.WithLogSendType(enum.CONSOLE))
	}

	if conf.Log.LogOut == int(enum.FILE) {
		logClient, err := logv1.NewClientCfg(config)
		if err != nil {
			goto defaultCli
		}

		return logClient.NewLogger(sdk.WithLogLevel(conf.Log.LogLevel), sdk.WithLogSendType(enum.FILE))
	}

	if conf.Log.LogOut == int(enum.MQ) {
		mqConfig := &sdk.MQ{
			MqType:       conf.Mq.MqType,
			ProducerHost: conf.Mq.MqProducerHost,
			ProducerPort: conf.Mq.MqProducerPort,
		}
		logClient, err := logv1.NewClientMQ(config, mqConfig)
		if err != nil {
			goto defaultCli
		}
		return logClient.NewLogger(sdk.WithLogLevel(conf.Log.LogLevel), sdk.WithLogSendType(enum.MQ))
	}

	if conf.Log.LogOut == int(enum.ZS) {
		zsConfig := &sdk.ZS{
			ZsAddrs:    conf.Search.Addr,
			ZsUser:     conf.Search.Username,
			ZsPassword: conf.Search.Password,
		}
		logClient, err := logv1.NewClientZs(config, zsConfig)
		if err != nil {
			goto defaultCli
		}
		cli := logClient.NewLogger(sdk.WithLogLevel(conf.Log.LogLevel), sdk.WithLogSendType(enum.ZS))
		if cli == nil {
			goto defaultCli
		}
	}

	if conf.Log.LogOut == int(enum.OTEL) {
		otelConfig := &sdk.OTEL{
			Endpoint: "http://" + conf.Otel.ExportEndpoint,
			User:     conf.Otel.Username,
			Pwd:      conf.Otel.Password,
		}
		logClient, err := logv1.NewClientOTEL(config, otelConfig)
		if err != nil {
			goto defaultCli
		}
		cli := logClient.NewLogger(sdk.WithLogLevel(conf.Log.LogLevel), sdk.WithLogSendType(enum.OTEL))
		if cli == nil {
			goto defaultCli
		}
	}

defaultCli:
	// default
	logClient, err := logv1.NewClient()
	if err != nil {
		panic(err)
	}

	return logClient.NewLogger()
}
