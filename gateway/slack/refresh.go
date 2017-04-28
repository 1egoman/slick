package gatewaySlack

import (
  "log"
	"github.com/1egoman/slime/gateway" // The thing to interface with slack
)

// Called when the connection becomes active
func (c *SlackConnection) Refresh() error {
	var err error

	// Fetch details about all channels
	c.channels, err = c.FetchChannels()
	if err != nil {
		return err
	}

	// Fetch details about the currently logged in user
	var user *gateway.User
	user, err = c.UserById(c.Self().Id)
	if err != nil {
		return err
	} else {
		c.self = *user
	}

	// Select the first channel, by default
	if len(c.channels) > 0 {
		c.selectedChannel = &c.channels[0]

		// Fetch Message history, if the emssage history is empty.
		if len(c.messageHistory) == 0 {
			log.Printf(
				"Fetching message history for team %s and channel %s",
				c.Team().Name,
				c.SelectedChannel().Name,
			)
			c.messageHistory, err = c.FetchChannelMessages(*c.selectedChannel)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

