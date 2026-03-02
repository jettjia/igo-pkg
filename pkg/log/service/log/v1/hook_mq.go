package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/sirupsen/logrus"

	"github.com/jettjia/go-pkg/pkg/event"
)

var (
	LOG_TOPIC = "xtext.log.create"
)

// mqHook
type mqHook struct {
	client     *mqClient
	serverName string
}

// newMqHook
func newMqHook(serverName string, mqType string, mqProducerHost string, mqProducerPort int) (*mqHook, error) {
	mq := newMqClient(mqType, mqProducerHost, mqProducerPort)
	return &mqHook{client: mq, serverName: serverName}, nil
}

// Fire logrus hook interface
func (hook *mqHook) Fire(entry *logrus.Entry) error {
	entry.Context = context.WithValue(context.TODO(), "service_name", hook.serverName)

	doc := getLogInfo(entry)
	gopool.Go(func() {
		hook.sendmq(doc)
	})

	return nil
}

// Levels logrus hook interface
func (hook *mqHook) Levels() []logrus.Level {
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

// sendmq asynchronously send logs to mq
func (hook *mqHook) sendmq(doc logInfoDocModel) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("send entry to mq failed: ", r)
		}
	}()
	err := hook.client.sendToMq(doc)
	if err != nil {
		log.Println(err)
	}
}

type mqClient struct {
	MqType         string `yaml:"mq_type"`
	MqProducerHost string `yaml:"mq_producer_host"`
	MqProducerPort int    `yaml:"mq_producer_port"`
}

func newMqClient(mqType string, mqProducerHost string, mqProducerPort int) *mqClient {
	return &mqClient{MqType: mqType, MqProducerHost: mqProducerHost, MqProducerPort: mqProducerPort}
}

// sendToMq send to mq
func (c *mqClient) sendToMq(reqParam logInfoDocModel) (err error) {
	var req []byte
	if req, err = json.Marshal(reqParam); err != nil {
		return
	}

	client := event.NewMQClient(c.MqProducerHost, c.MqProducerPort, c.MqProducerHost, c.MqProducerPort, c.MqType)
	err = client.Pub(LOG_TOPIC, req)

	return
}
