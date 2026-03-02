package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/sirupsen/logrus"
)

// otelHook
type otelHook struct {
	client     *otelClient
	serverName string
}

// newOtelHook
func newOtelHook(serverName string, endpoint string, user string, pwd string) (*otelHook, error) {
	otel := newotelClient(endpoint, user, pwd)
	return &otelHook{client: otel, serverName: serverName}, nil
}

// Fire logrus hook interface
func (hook *otelHook) Fire(entry *logrus.Entry) (err error) {
	entry.Context = context.WithValue(context.TODO(), "service_name", hook.serverName)

	doc := getLogInfo(entry)
	gopool.Go(func() {
		hook.sendotel(doc)
	})
	return err
}

// Levels logrus hook interface
func (hook *otelHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
}

// sendotel asynchronously send logs to otel
func (hook *otelHook) sendotel(doc logInfoDocModel) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("send entry to otel failed: ", r)
		}
	}()
	err = hook.client.sendToOtel(doc)
	if err != nil {
		log.Println(err)
	}

	return err
}

type otelClient struct {
	Endpoint string
	User     string
	Pwd      string
}

func newotelClient(endpoint string, user string, pwd string) *otelClient {
	return &otelClient{Endpoint: endpoint, User: user, Pwd: pwd}
}

// sendTootel send to otel
func (c *otelClient) sendToOtel(reqParam logInfoDocModel) (err error) {
	var data []byte
	if data, err = json.Marshal(reqParam); err != nil {
		return
	}
	url := c.Endpoint + "/api/default/default/_json"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(data)))
	if err != nil {
		log.Println(err)
		return
	}

	req.SetBasicAuth(c.User, c.Pwd)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	return
}
