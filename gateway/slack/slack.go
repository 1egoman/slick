package gatewaySlack

import (
	"log"
	"strconv"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/1egoman/slime/gateway"
	"golang.org/x/net/websocket"
)

func New(token string) *SlackConnection {
	return &SlackConnection{
		token: token,
	}
}

// SlackConnection meets the connection interface.
type SlackConnection struct {
	url   string
	token string
	conn  *websocket.Conn

	// Create two message channels, one for incoming messages and one for outgoing messages.
	incoming chan gateway.Event
	outgoing chan gateway.Event

	self gateway.User
	team gateway.Team

	// Internal state to store all channels and a pointer to the active one.
	channels        []gateway.Channel
	selectedChannel *gateway.Channel

	// Internal state to store message history of the active channel
	messageHistory []gateway.Message
}

// Return the name of the team.
func (c *SlackConnection) Name() string {
	if c.Team() != nil && len(c.Team().Name) != 0 {
		return c.Team().Name
	} else {
		return "(slack loading...)"
	}
}

// Fetch all channels for the given team
func (c *SlackConnection) FetchChannels() ([]gateway.Channel, error) {
	log.Printf("Fetching list of channels for team %s", c.Team().Name)
	resp, err := http.Get("https://slack.com/api/channels.list?token=" + c.token)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var slackChannelBuffer struct {
		Channels []struct {
			Id         string `json:"id"`
			Name       string `json:"name"`
			CreatorId  string `json:"creator"`
			Created    int    `json:"created"`
			IsMember   bool   `json:"is_member"`
			IsArchived bool   `json:"is_archived"`
		} `json:"channels"`
	}
	json.Unmarshal(body, &slackChannelBuffer)

	// Convert to more generic message format
	var channelBuffer []gateway.Channel
	var creator *gateway.User
	for _, channel := range slackChannelBuffer.Channels {
		creator, err = c.UserById(channel.CreatorId)
		if err != nil {
			return nil, err
		}
		channelBuffer = append(channelBuffer, gateway.Channel{
			Id:         channel.Id,
			Name:       channel.Name,
			Creator:    creator,
			Created:    channel.Created,
			IsMember:   channel.IsMember,
			IsArchived: channel.IsArchived,
		})
	}

	// Set the internal state of the component.
	// This is used by the `connect` step to prelaod a list of channels for the fuzzy picker
	c.channels = channelBuffer

	return channelBuffer, nil
}

// Given a channel, return all messages within that channel.
func (c *SlackConnection) FetchChannelMessages(channel gateway.Channel) ([]gateway.Message, error) {
	log.Printf("Fetching channel messages for team %s", c.Team().Name)
	resp, err := http.Get("https://slack.com/api/channels.history?token=" + c.token + "&channel=" + channel.Id + "&count=100")
	if err != nil {
		return nil, err
	}

	// Parse slack messages
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var slackMessageBuffer struct {
		Messages []struct {
			Ts        string `json:"ts"`
			UserId    string `json:"user"`
			Text      string `json:"text"`
			Reactions []struct {
				Name  string   `json:"name"`
				Users []string `json:"users"`
			} `json:"reactions"`
			File struct {
				Name       string `json:"name"`
				Filetype   string `json:"pretty_type"`
				User       string `json:"user"`
				PrivateUrl string `json:"url_private"`
				Permalink  string `json:"permalink"`
				Reactions  []struct {
					Name  string   `json:"name"`
					Users []string `json:"users"`
				} `json:"reactions"`
			} `json:"file,omitempty"`
		} `json:"messages"`
		hasMore bool `json:"has_more"`
	}
	if err = json.Unmarshal(body, &slackMessageBuffer); err != nil {
		return nil, err
	}

	// Convert to more generic message format
	var messageBuffer []gateway.Message
	var sender *gateway.User
	cachedUsers := make(map[string]*gateway.User)
	for i := len(slackMessageBuffer.Messages) - 1; i >= 0; i-- { // loop backwards to reverse the final slice
		msg := slackMessageBuffer.Messages[i]

		// Get the sender of the message
		// Since we're likely to have a lot of the same users, cache them.
		if cachedUsers[msg.UserId] != nil {
			sender = cachedUsers[msg.UserId]
		} else {
			sender, err = c.UserById(msg.UserId)
			cachedUsers[msg.UserId] = sender
			if err != nil {
				return nil, err
			}
		}

		// Convert the reactions fetched into reaction objects
		reactions := []gateway.Reaction{}
		reactionsLocation := msg.Reactions
		if len(msg.File.Reactions) > 0 {
			reactionsLocation = msg.File.Reactions
		}
		for _, reaction := range reactionsLocation {
			reactionUsers := []*gateway.User{}
			// reaction.Users is an array of string user ids. Convert each into user objects.
			for _, reactionUserId := range reaction.Users {
				if cachedUsers[reactionUserId] != nil {
					reactionUsers = append(reactionUsers, cachedUsers[reactionUserId])
				} else {
					var reactionUser *gateway.User
					reactionUser, err = c.UserById(reactionUserId)
					if err != nil {
						return nil, err
					}
					reactionUsers = append(reactionUsers, reactionUser)
				}
			}

			// Add the final reaction to the collection
			reactions = append(reactions, gateway.Reaction{Name: reaction.Name, Users: reactionUsers})
		}

		var file *gateway.File
		if len(msg.File.Name) > 0 {
			// Given a user id, get a reference to the user.
			var fileUser *gateway.User
			if cachedUsers[msg.File.User] != nil {
				fileUser = cachedUsers[msg.File.User]
			} else {
				fileUser, err = c.UserById(msg.File.User)
				if err != nil {
					return nil, err
				}
			}

			// Create the file struct representation.
			file = &gateway.File{
				Name:       msg.File.Name,
				Filetype:   msg.File.Filetype,
				User:       fileUser,
				PrivateUrl: msg.File.PrivateUrl,
				Permalink:  msg.File.Permalink,
			}
		} else {
			file = nil
		}

		// Convert timestamp to float64
		// I would unmarshal directly into float64, but that doesn't work since slack encodes their
		// timestamps as strings :/
		var timestamp float64
		timestamp, err = strconv.ParseFloat(msg.Ts, 64)
		if err != nil {
			return nil, err
		}

		messageBuffer = append(messageBuffer, gateway.Message{
			Sender:    sender,
			Text:      msg.Text,
			Reactions: reactions,
			Timestamp: int(timestamp), // this value is in seconds!
			Hash:      msg.Ts,
			File:      file,
		})
	}

	return messageBuffer, nil
}

func (c *SlackConnection) UserById(id string) (*gateway.User, error) {
	resp, err := http.Get("https://slack.com/api/users.info?token=" + c.token + "&user=" + id)
	if err != nil {
		return nil, err
	}

	// Parse slack user buffer
	body, _ := ioutil.ReadAll(resp.Body)
	var slackUserBuffer struct {
		User struct {
			Id      string `json:"id"`
			Name    string `json:"name"`
			Color   string `json:"color"`
			Profile struct {
				Status   string `json:"color"`
				RealName string `json:"real_name"`
				Email    string `json:"email"`
				Phone    string `json:"phone"`
				Skype    string `json:"skype"`
				Image    string `json:"image_24"`
			} `json:"profile"`
		} `json:"user"`
	}
	if err = json.Unmarshal(body, &slackUserBuffer); err != nil {
		return nil, err
	}

	// Convert to a generic User
	return &gateway.User{
		Id:       slackUserBuffer.User.Id,
		Name:     slackUserBuffer.User.Name,
		Color:    slackUserBuffer.User.Color,
		Avatar:   slackUserBuffer.User.Profile.Image,
		Status:   slackUserBuffer.User.Profile.Status,
		RealName: slackUserBuffer.User.Profile.RealName,
		Email:    slackUserBuffer.User.Profile.Email,
		Skype:    slackUserBuffer.User.Profile.Skype,
		Phone:    slackUserBuffer.User.Profile.Phone,
	}, nil
}

func (c *SlackConnection) MessageHistory() []gateway.Message {
	return c.messageHistory
}
func (c *SlackConnection) AppendMessageHistory(message gateway.Message) {
	c.messageHistory = append(c.messageHistory, message)
}
func (c *SlackConnection) ClearMessageHistory() {
	c.messageHistory = []gateway.Message{}
}

func (c *SlackConnection) SelectedChannel() *gateway.Channel {
	return c.selectedChannel
}
func (c *SlackConnection) SetSelectedChannel(channel *gateway.Channel) {
	c.selectedChannel = channel
	// When setting a new channel, clear out the message history so that messages will be refetched.
	c.messageHistory = []gateway.Message{}
}

func (c *SlackConnection) Incoming() chan gateway.Event {
	return c.incoming
}
func (c *SlackConnection) Outgoing() chan gateway.Event {
	return c.outgoing
}
func (c *SlackConnection) Team() *gateway.Team {
	return &c.team
}
func (c *SlackConnection) Channels() []gateway.Channel {
	return c.channels
}
func (c *SlackConnection) Self() *gateway.User {
	return &c.self
}
