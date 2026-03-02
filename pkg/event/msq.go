package event

import (
	"errors"
	"fmt"
	"strings"
)

type MessageHandler func(msg []byte) error

type EventMQClient interface {
	Pub(topic string, msg []byte) error
	Sub(topic string, channel string, handler MessageHandler, pollIntervalMilliseconds int64, maxInFlight int) error
	Close()
}

func parseHost(host string) string {
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]", host)
	}
	return host
}

type NewClientFn func(pubServer string, pubPort int, subServer string, subPort int) EventMQClient

var ncfFactory map[string]NewClientFn

func init() {
	ncfFactory = make(map[string]NewClientFn, 3)
	ncfFactory["redis"] = NewRedisClient
	ncfFactory["redis-stream"] = NewRedisStreamClient
	ncfFactory["memory"] = NewMemoryClient
	ncfFactory["pgsql"] = NewPgClient
}

func NewEventMQClient(pubServer string, pubPort int, subServer string, subPort int, msqType string) (EventMQClient, error) {
	if fn, ok := ncfFactory[msqType]; !ok {
		err := fmt.Errorf("NewEventMQClient unknown msq type %v", msqType)
		return nil, err
	} else {
		client := fn(pubServer, pubPort, subServer, subPort)
		var errs []error
		return client, errors.Join(errs...)
	}
}
