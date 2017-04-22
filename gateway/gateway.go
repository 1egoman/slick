package gateway

import (
	"encoding/json"
	// "fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/websocket"
)

type Event struct {
	Direction string
	Channel   Channel
	User      User

	Type string `json:"type"`
	Data map[string]interface{}
}

type Connection interface {
	Connect() error
	Channels() ([]Channel, error)
	GetChannelMessages(Channel) ([]Message, error)

	GetUserById(string) (*User, error)
}

// Conenction Primatives
type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Avatar   string `json:"avatar"`
	Status   string `json:"status"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Skype    string `json:"skype"`
	Phone    string `json:"phone"`
}

type Team struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type Channel struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Creator *User  `json:"creator"`
	Created int    `json:"created"`
}

type Reaction struct {
	Name  string `json:"name"`
	Users []User `json:"users"`
}

type Message struct {
	Sender    *User      `json:"sender"`
	Text      string     `json:"text"`
	Reactions []Reaction `json:"reactions"`
}

func Slack(token string) *SlackConnection {
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
	Incoming chan Event
	Outgoing chan Event

	Self User
	Team Team
}

func (c *SlackConnection) requestConnectionUrl() {
	// Make request to slack's api to get websocket credentials
	// https://api.slack.com/methods/rtm.connect
	resp, err := http.Get("https://slack.com/api/rtm.connect?token=" + c.token)

	if err != nil {
		panic(err)
	}

	// Decode json body.
	body, _ := ioutil.ReadAll(resp.Body)
	var connectionBuffer struct {
		Ok   bool   `json:"ok"`
		Url  string `json:"url"`
		Team Team   `json:"team"`
		Self User   `json:"self"`
	}
	json.Unmarshal(body, &connectionBuffer)

	// Add response data to struct
	c.url = connectionBuffer.Url
	c.Self = connectionBuffer.Self
	c.Team = connectionBuffer.Team
}

// Connect to the slack persistent socket.
func (c *SlackConnection) Connect() error {
	// Request a connection url with the token in the struct
	c.requestConnectionUrl()

	// FIXME: what does this mean?
	origin := "http://localhost/"

	// Create a connection to the websocket
	var err error
	c.conn, err = websocket.Dial(c.url, "", origin)
	if err != nil {
		return err
	}

	// Create buffered channels to listen and send messages on
	c.Incoming = make(chan Event, 5)
	c.Outgoing = make(chan Event, 5)

	// When messages are received, add them to the incoming buffer.
	go func(incoming chan Event) {
		var msgRaw = make([]byte, 512)
		var msg map[string]interface{}
		var n int

		for {
			// Listen for messages, and when some are received, write them to a channel.
			if n, err = c.conn.Read(msgRaw); err != nil {
				panic(err)
			}
			// fmt.Printf("Received in goroutine: %s.\n", msgRaw[:n])

			// Decode into a struct so that we can check message type later
			json.Unmarshal(msgRaw[:n], &msg)
			incoming <- Event{
				Direction: "incoming",
				Type:      msg["type"].(string),
				Data:      msg,
			}
		}
	}(c.Incoming)

	// When messages are in the outgoing buffer waiting to be sent, send them.
	go func(outgoing chan Event) {
		// Add a sequential message id to each message sent, so replies can later be tracked.
		messageId := 0

		var event Event
		for {
			// Assemble the message to send.
			event = <-outgoing
			event.Data["type"] = event.Type
			messageId++
			event.Data["id"] = messageId

			// Marshal to json
			data, err := json.Marshal(event.Data)
			if err != nil {
				panic(err)
			}
			// fmt.Printf("Writing to slack: %s\n", data)

			// Send it.
			if _, err = c.conn.Write(data); err != nil {
				panic(err)
			}
		}
	}(c.Outgoing)

	return nil
}

// Fetch all channels for the given team
func (c *SlackConnection) Channels() ([]Channel, error) {
	resp, err := http.Get("https://slack.com/api/channels.list?token=" + c.token)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var slackChannelBuffer struct {
		Channels []struct {
			Id        string `json:"id"`
			Name      string `json:"name"`
			CreatorId string `json:"creator"`
			Created   int    `json:"created"`
		} `json:"channels"`
	}
	json.Unmarshal(body, &slackChannelBuffer)

	// Convert to more generic message format
	var channelBuffer []Channel
	var creator *User
	for _, channel := range slackChannelBuffer.Channels {
		creator, err = c.GetUserById(channel.CreatorId)
		if err != nil {
			return nil, err
		}
		channelBuffer = append(channelBuffer, Channel{
			Id:      channel.Id,
			Name:    channel.Name,
			Creator: creator,
			Created: channel.Created,
		})
	}

	return channelBuffer, nil
}

// Given a channel, return all messages within that channel.
func (c *SlackConnection) GetChannelMessages(channel Channel) ([]Message, error) {
	resp, err := http.Get("https://slack.com/api/channels.history?token=" + c.token + "&channel=" + channel.Id + "&count=100")
	if err != nil {
		return nil, err
	}

	// Parse slack messages
	body, _ := ioutil.ReadAll(resp.Body)
	var slackMessageBuffer struct {
		Messages []struct {
			Timestamp string     `json:"ts"`
			UserId    string     `json:"user"`
			Text      string     `json:"text"`
			Reactions []Reaction `json:"reactions"`
		} `json:"messages"`
		hasMore bool `json:"has_more"`
	}
	if err = json.Unmarshal(body, &slackMessageBuffer); err != nil {
		return nil, err
	}

	// Convert to more generic message format
	var messageBuffer []Message
	var sender *User
	for _, msg := range slackMessageBuffer.Messages {
		sender, err = c.GetUserById(msg.UserId)
		if err != nil {
			return nil, err
		}

		messageBuffer = append(messageBuffer, Message{
			Sender:    sender,
			Text:      msg.Text,
			Reactions: msg.Reactions,
		})
	}

	return messageBuffer, nil
}

func (c *SlackConnection) GetUserById(id string) (*User, error) {
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
	return &User{
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
