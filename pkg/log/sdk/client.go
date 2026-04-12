package sdk

type Client struct {
	Config     *Config // basic log configuration information
	ConfigZS   *ZS     // zincsearch
	ConfigMQ   *MQ     // MQ
	ConfigOTEL *OTEL   // OTEL
}

func (c *Client) WithConfig(config *Config) *Client {
	c.Config = config
	return c
}

func (c *Client) WithZS(config *ZS) *Client {
	c.ConfigZS = config
	return c
}

func (c *Client) WithMQ(config *MQ) *Client {
	c.ConfigMQ = config
	return c
}

func (c *Client) WithOTEL(config *OTEL) *Client {
	c.ConfigOTEL = config
	return c
}
