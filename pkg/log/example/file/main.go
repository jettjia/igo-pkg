package main

import (
	"github.com/jettjia/igo-pkg/pkg/log/enum"
	"github.com/jettjia/igo-pkg/pkg/log/sdk"
	logv1 "github.com/jettjia/igo-pkg/pkg/log/service/log/v1"
)

func main() {
	logClient, err := logv1.NewClient()
	if err != nil {
		panic(err)
	}

	logClient.NewLogger(sdk.WithLogLevel("info"), sdk.WithLogSendType(enum.FILE), sdk.WithLogPath("./tmp")).Infof("info,console...")
}
