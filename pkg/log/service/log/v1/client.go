package v1

import (
	"github.com/jettjia/igo-pkg/pkg/log/sdk"
)

type Client struct {
	sdk.Client
	Config *sdk.Config
	ES     *sdk.ES
	ZS     *sdk.ZS
	MQ     *sdk.MQ
	OTEL   *sdk.OTEL
}

// NewClient default client
func NewClient() (client *Client, err error) {
	client = &Client{}
	return
}

// NewClientCfg file client
func NewClientCfg(config *sdk.Config) (client *Client, err error) {
	client = &Client{
		Config: config,
	}
	client.WithConfig(config)

	return
}

func NewClientES(config *sdk.Config, es *sdk.ES) (client *Client, err error) {
	client = &Client{
		Config: config,
		ES:     es,
	}
	client.WithConfig(config).WithES(es)

	return
}

func NewClientZs(config *sdk.Config, zs *sdk.ZS) (client *Client, err error) {
	client = &Client{
		Config: config,
		ZS:     zs,
	}
	client.WithConfig(config).WithZS(zs)

	return
}

func NewClientMQ(config *sdk.Config, mq *sdk.MQ) (client *Client, err error) {
	client = &Client{
		Config: config,
		MQ:     mq,
	}
	client.WithConfig(config).WithMQ(mq)

	return
}

func NewClientOTEL(config *sdk.Config, otel *sdk.OTEL) (client *Client, err error) {
	client = &Client{
		Config: config,
		OTEL:   otel,
	}
	client.WithConfig(config).WithOTEL(otel)

	return
}
