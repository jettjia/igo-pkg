package main

import (
	"github.com/jettjia/igo-pkg/pkg/log/enum"
	"github.com/jettjia/igo-pkg/pkg/log/sdk"
	logv1 "github.com/jettjia/igo-pkg/pkg/log/service/log/v1"
)

func main() {
	config := &sdk.Config{
		LogFilePath: "./tmp",
		LogName:     "mq",
		MaxSize:     1024,
		MaxBackup:   64,
		MaxAge:      30,
		LogLevel:    "error",
		LogSendType: enum.MQ,
	}

	mq := &sdk.MQ{
		MqType:       "kafka",
		ProducerHost: "127.0.0.1",
		ProducerPort: 9092,
	}

	logClient, err := logv1.NewClientMQ(config, mq)
	if err != nil {
		panic(err)
	}

	logClient.NewLogger(sdk.WithLogLevel("error"), sdk.WithLogSendType(enum.MQ)).Errorf("err, mq...")

}
