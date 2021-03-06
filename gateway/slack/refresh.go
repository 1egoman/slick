package gatewaySlack

import (
	"github.com/1egoman/slick/gateway" // The thing to interface with slack
	"log"
)

// Called when the connection becomes active
func (c *SlackConnection) Refresh(force bool) error {
	var err error

	// Fetch details about all channels
	if force || len(c.channels) == 0 {
		c.channels, err = c.FetchChannels()
		if err != nil {
			return err
		}
	}

	// Fetch details about the currently logged in user
	var user *gateway.User
	user, err = c.UserById(c.Self().Id)
	if err != nil {
		return err
	} else {
		c.self = *user
	}

	// Fetch message history, if the message history is empty.
	if (force || len(c.messageHistory) == 0) && c.Team() != nil && c.SelectedChannel() != nil {
		log.Printf(
			"Fetching message history for team %s and channel %s",
			c.Team().Name,
			c.SelectedChannel().Name,
		)
		c.messageHistory, err = c.FetchChannelMessages(*c.selectedChannel, nil)
		if err != nil {
			return err
		}
	}

	// If no channel is selected, select a default.
	if len(c.channels) > 0 && c.selectedChannel == nil {
		// Try to find the general channel
		for _, channel := range c.channels {
			if channel.Name == "general" {
				c.selectedChannel = &channel
				break
			}
		}
		// or, if that can't be found, select the first one.
		if c.selectedChannel == nil {
			c.selectedChannel = &c.channels[0]
		}
	}

	return nil
}
