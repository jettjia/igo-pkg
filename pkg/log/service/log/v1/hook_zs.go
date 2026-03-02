package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/sirupsen/logrus"

	"github.com/jettjia/igo-pkg/pkg/log/util"
)

// zsHook
type zsHook struct {
	cmd        string
	client     *zsClient
	serverName string
}

// newZsHook
func newZsHook(serverName string, addr string, user string, password string) (*zsHook, error) {
	zs := newZsClient(serverName, addr, user, password)
	return &zsHook{client: zs, cmd: strings.Join(os.Args, " "), serverName: serverName}, nil
}

// Fire logrus hook interface
func (hook *zsHook) Fire(entry *logrus.Entry) error {
	entry.Context = context.WithValue(context.TODO(), "service_name", hook.serverName)

	doc := getLogInfo(entry)
	doc["cmd"] = hook.cmd
	gopool.Go(func() {
		hook.sendZs(doc)
	})
	return nil
}

// Levels logrus hook interface
func (hook *zsHook) Levels() []logrus.Level {
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

// sendZs asynchronously send logs to zs
func (hook *zsHook) sendZs(doc logInfoDocModel) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("send entry to zs failed: ", r)
		}
	}()
	err := hook.client.createDc(doc)
	if err != nil {
		log.Println(err)
	}
}

type zsClient struct {
	ServerName string
	Addr       string
	User       string
	Password   string
}

func newZsClient(serverName string, addr string, user string, password string) *zsClient {
	return &zsClient{ServerName: serverName, Addr: addr, User: user, Password: password}
}

func (c *zsClient) GetIndexName() string {
	return c.ServerName + "-" + time.Now().Local().Format("2006-01-02")
}

//// createIndex 创建索引
//func (c *zsClient) createIndex() {
//	reqParam := map[string]string{}
//	reqParam["name"] = c.GetIndexName()
//	req, _ := json.Marshal(reqParam)
//	err := util.ZsPost(c.Addr+"/api/"+c.GetIndexName()+"/_doc", c.User, c.Password, string(req))
//	if err != nil {
//		fmt.Println("create zincsearch index err: ", err)
//	}
//}

// createDc create a log document
func (c *zsClient) createDc(reqParam logInfoDocModel) error {
	req, _ := json.Marshal(reqParam)
	_, err := util.ZsHttpClient("POST", c.Addr+"/api/"+c.GetIndexName()+"/_doc", string(req), c.User, c.Password)
	if err != nil {
		fmt.Println("create zincsearch index err: ", err)
		return err
	}

	return nil
}
