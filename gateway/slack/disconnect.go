package gatewaySlack

import (
	"github.com/1egoman/slick/gateway"
)

func (c *SlackConnection) Disconnect() error {
	if c.Status() == gateway.CONNECTED {
		c.conn.Close()
	}
	c.connectionStatus = gateway.DISCONNECTED
	return nil
}
