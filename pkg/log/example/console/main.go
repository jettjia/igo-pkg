package main

import (
	"github.com/jettjia/igo-pkg/pkg/log/sdk"
	logv1 "github.com/jettjia/igo-pkg/pkg/log/service/log/v1"
)

// print a normal log
func main() {
	logClient, err := logv1.NewClient()
	if err != nil {
		panic(err)
	}

	// logClient.NewLogger().Errorf("err,console...")

	logClient.NewLogger(sdk.WithLogLevel("info")).Infof("info,console...")
}
