package gatewaySlack

import (
	"github.com/1egoman/slime/gateway"
)

func (c *SlackConnection) Disconnect() error {
	c.conn.Close()
	c.connectionStatus = gateway.DISCONNECTED
	return nil
}
