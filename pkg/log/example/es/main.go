package main

import (
	"github.com/jettjia/igo-pkg/pkg/log/enum"
	"github.com/jettjia/igo-pkg/pkg/log/sdk"
	logv1 "github.com/jettjia/igo-pkg/pkg/log/service/log/v1"
)

func main() {
	config := &sdk.Config{
		LogFilePath: "./tmp",
		LogName:     "es",
		MaxSize:     1024,
		MaxBackup:   64,
		MaxAge:      30,
		LogLevel:    "info",
		LogSendType: enum.ES, // send to es
	}

	es := &sdk.ES{
		EsAddrs:    []string{"http://127.0.0.1:9200/"},
		EsUser:     "",
		EsPassword: "",
	}

	logClient, err := logv1.NewClientES(config, es)
	if err != nil {
		panic(err)
	}

	logClient.NewLogger(sdk.WithLogLevel("info"), sdk.WithLogSendType(enum.ES)).Errorf("err,es...")

}
