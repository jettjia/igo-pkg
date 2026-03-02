package main

import (
	"github.com/jettjia/go-pkg/pkg/log/enum"
	"github.com/jettjia/go-pkg/pkg/log/sdk"
	logv1 "github.com/jettjia/go-pkg/pkg/log/service/log/v1"
)

func main() {
	config := &sdk.Config{
		LogFilePath: "./tmp",
		LogName:     "otel",
		MaxSize:     1024,
		MaxBackup:   64,
		MaxAge:      30,
		LogLevel:    "info",
		LogSendType: enum.OTEL,
	}

	otel := &sdk.OTEL{
		Endpoint: "http://127.0.0.1:5080",
		User:     "xxx",
		Pwd:      "xxx",
	}

	logClient, err := logv1.NewClientOTEL(config, otel)
	if err != nil {
		panic(err)
	}

	logClient.NewLogger(sdk.WithLogLevel("info"), sdk.WithLogSendType(enum.OTEL)).Errorf("err, otel...")

}
