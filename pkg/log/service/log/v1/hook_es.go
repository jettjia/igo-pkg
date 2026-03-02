package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

type esHook struct {
	cmd        string
	client     *elasticsearch.TypedClient
	serverName string
}

// newEsHook
func newEsHook(serverName string, addr []string, user string, password string) (*esHook, error) {
	cfg := elasticsearch.Config{
		Addresses: addr,
		Username:  user,
		Password:  password,
	}

	es, err := elasticsearch.NewTypedClient(cfg)

	if err != nil {
		return nil, err
	}
	return &esHook{client: es, cmd: strings.Join(os.Args, " "), serverName: serverName}, nil
}

// Fire logrus hook interface
func (hook *esHook) Fire(entry *logrus.Entry) error {
	entry.Context = context.WithValue(context.TODO(), "service_name", hook.serverName)

	doc := newEsLog(entry)
	doc["cmd"] = hook.cmd
	gopool.Go(func() {
		hook.sendEs(doc)
	})
	return nil
}

// Levels logrus hook interface
func (hook *esHook) Levels() []logrus.Level {
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

// sendEs asynchronously send logs to es
func (hook *esHook) sendEs(doc appLogDocModel) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("send entry to es failed: ", r)
		}
	}()

	bodyJson, err := json.Marshal(doc)
	if err != nil {
		return
	}

	req := esapi.IndexRequest{
		Index:   doc.indexName(),
		Body:    strings.NewReader(string(bodyJson)),
		Refresh: "true",
	}

	_, err = req.Do(context.Background(), hook.client)

	if err != nil {
		fmt.Println(err)
	}

}

// appLogDocModel es model
type appLogDocModel map[string]interface{}

func newEsLog(e *logrus.Entry) appLogDocModel {
	ins := map[string]interface{}{}
	for kk, vv := range e.Data {
		ins[kk] = vv
	}
	ins["time"] = time.Now().Local()
	ins["lvl"] = e.Level
	ins["message"] = e.Message
	//ins["caller"] = fmt.Sprintf("%s:%d  %#v", e.Caller.File, e.Caller.Line, e.Caller.Func)
	return ins
}

// indexName es index name time splitting
func (m *appLogDocModel) indexName() string {
	return "doclib-" + time.Now().Local().Format("2006-01-02")
}
