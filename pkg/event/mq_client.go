package event

import (
	"log"
	"sync"
)

var (
	mqOnce      sync.Once
	eventClient MQClient
)

type MQClient interface {
	Pub(string, []byte) (err error)
	Sub(topic, channel string, cmd func([]byte) error) (err error)
	Close()
}

// mqclient
type mqclient struct {
	prontonMQClient          EventMQClient
	pollIntervalMilliseconds int64
	maxInFlight              int
}

// NewMQClient create a message queuing client
func NewMQClient(producerHost string, producerPort int, consumerHost string, consumerPort int, connectorType string) MQClient {
	mqOnce.Do(func() {
		client, err := NewEventMQClient(producerHost, producerPort, consumerHost, consumerPort, connectorType)
		if err != nil {
			panic(err)
		}

		eventClient = &mqclient{
			prontonMQClient:          client,
			pollIntervalMilliseconds: int64(100),
			maxInFlight:              16,
		}
	})

	return eventClient
}

// Pub mq producer
func (m *mqclient) Pub(topic string, msg []byte) (err error) {
	err = m.prontonMQClient.Pub(topic, msg)
	return
}

// Sub mq consumers
func (m *mqclient) Sub(topic, channel string, cmd func([]byte) error) (err error) {
	go func() {
		err = m.prontonMQClient.Sub(topic, channel, cmd, m.pollIntervalMilliseconds, m.maxInFlight)
		if err != nil {
			log.Printf("mqclient Sub error: %v", err)
		}
	}()

	return
}

// Close
func (m *mqclient) Close() {
	m.prontonMQClient.Close()
}
