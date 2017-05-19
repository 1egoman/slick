package gatewaySlack

import (
	"fmt"
	"log"
	"strconv"
	"errors"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/1egoman/slime/gateway"
	"golang.org/x/net/websocket"
)

func New(token string) *SlackConnection {
	return &SlackConnection{
		token: token,
		nickname: nil,
	}
}
func NewWithName(nickname string, token string) *SlackConnection {
	return &SlackConnection{
		token: token,
		nickname: &nickname,
	}
}

// SlackConnection meets the connection interface.
type SlackConnection struct {
	url   string
	token string
	nickname *string
	conn  *websocket.Conn
	connectionStatus gateway.ConnectionStatus

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
	if c.nickname != nil && len(*c.nickname) > 0 {
		return *c.nickname
	} else if c.Team() != nil && len(c.Team().Name) != 0 {
		return c.Team().Name
	} else {
		return "(slack loading...)"
	}
}

// Fetch all channels for the given team
func (c *SlackConnection) FetchChannels() ([]gateway.Channel, error) {
	var channelBuffer []gateway.Channel

	// FETCH CHANNELs
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
	var creator *gateway.User
	for _, channel := range slackChannelBuffer.Channels {
		creator, err = c.UserById(channel.CreatorId)
		if err != nil {
			return nil, err
		}
		channelBuffer = append(channelBuffer, gateway.Channel{
			Id:         channel.Id,
			SubType:    gateway.TYPE_CHANNEL,
			Name:       channel.Name,
			Creator:    creator,
			Created:    channel.Created,
			IsMember:   channel.IsMember,
			IsArchived: channel.IsArchived,
		})
	}



	// FETCH IMs
	log.Printf("Fetching list of ims for team %s", c.Team().Name)
	resp, err = http.Get("https://slack.com/api/im.list?token=" + c.token)
	if err != nil {
		return nil, err
	}

	body, _ = ioutil.ReadAll(resp.Body)
	var slackImBuffer struct {
		Ims []struct {
			Id         string `json:"id"`
			User       string `json:"user"`
			Created    int    `json:"created"`
		} `json:"ims"`
	}
	json.Unmarshal(body, &slackImBuffer)

	var otherUser *gateway.User
	for _, im := range slackImBuffer.Ims {
		otherUser, err = c.UserById(im.User)
		if err != nil {
			return nil, err
		}
		channelBuffer = append(channelBuffer, gateway.Channel{
			Id:         im.Id,
			SubType:    gateway.TYPE_DIRECT_MESSAGE,
			Name:       fmt.Sprintf("im-%s-%s", c.Self().Name, otherUser.Name),
			Creator:    c.Self(),
			Created:    im.Created,
			IsMember:   true,
			IsArchived: false,
		})
	}




	// FETCH MPIMs
	log.Printf("Fetching list of mp ims for team %s", c.Team().Name)
	resp, err = http.Get("https://slack.com/api/mpim.list?token=" + c.token)
	if err != nil {
		return nil, err
	}

	body, _ = ioutil.ReadAll(resp.Body)
	var slackMpimBuffer struct {
		Groups []struct {
			Id         string `json:"id"`
			Name       string `json:"name"`
			CreatorId  string `json:"creator"`
			Created    int    `json:"created"`
			IsMember   bool   `json:"is_member"`
			IsArchived bool   `json:"is_archived"`
		} `json:"groups"`
	}
	json.Unmarshal(body, &slackMpimBuffer)

	for _, mpim := range slackMpimBuffer.Groups {
		creator, err = c.UserById(mpim.CreatorId)
		if err != nil {
			return nil, err
		}
		channelBuffer = append(channelBuffer, gateway.Channel{
			Id:         mpim.Id,
			SubType:    gateway.TYPE_GROUP_DIRECT_MESSAGE,
			Name:       mpim.Name,
			Creator:    creator,
			Created:    mpim.Created,
			IsMember:   mpim.IsMember,
			IsArchived: mpim.IsArchived,
		})
	}


	// Set the internal state of the component.
	// This is used by the `connect` step to prelaod a list of channels for the fuzzy picker
	c.channels = channelBuffer

	return channelBuffer, nil
}

// Given a channel, return all messages within that channel.
func (c *SlackConnection) FetchChannelMessages(channel gateway.Channel, startTs *string) ([]gateway.Message, error) {
	log.Printf("Fetching channel messages for team %s starting at %s", c.Team().Name, startTs)

	// Contruct the request url
	var url string
	if channel.SubType == gateway.TYPE_CHANNEL {
		url = "https://slack.com/api/channels.history?token=" + c.token
	} else if channel.SubType == gateway.TYPE_DIRECT_MESSAGE {
		url = "https://slack.com/api/im.history?token=" + c.token
	} else if channel.SubType == gateway.TYPE_GROUP_DIRECT_MESSAGE {
		url = "https://slack.com/api/mpim.history?token=" + c.token
	}
	url += "&channel=" + channel.Id
	url += "&count=100"

	// If a starting timestamp was passed, get all messages after that timestamp.
	if startTs != nil {
		url += "&latest=" + *startTs
	}

	log.Println("Fetching history from slack", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Parse slack messages
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var slackMessageBuffer struct {
		Messages []map[string]interface{} `json:"messages"`
		hasMore bool `json:"has_more"`
	}
	if err = json.Unmarshal(body, &slackMessageBuffer); err != nil {
		return nil, err
	}

	// Convert to more generic message format
	var messageBuffer []gateway.Message
	cachedUsers := make(map[string]*gateway.User)
	for i := len(slackMessageBuffer.Messages) - 1; i >= 0; i-- { // loop backwards to reverse the final slice
		var message *gateway.Message
		message, err = c.ParseMessage(slackMessageBuffer.Messages[i], cachedUsers)
		if err == nil {
			messageBuffer = append(messageBuffer, *message)
		} else {
			return nil, err
		}
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
func (c *SlackConnection) SetMessageHistory(messages []gateway.Message) {
	c.messageHistory = messages
}
func (c *SlackConnection) AppendMessageHistory(message gateway.Message) {
	c.messageHistory = append(c.messageHistory, message)
}
func (c *SlackConnection) PrependMessageHistory(message gateway.Message) {
	c.messageHistory = append([]gateway.Message{message}, c.messageHistory...)
}
func (c *SlackConnection) DeleteMessageHistory(index int) {
	c.messageHistory = append(c.messageHistory[:index], c.messageHistory[index+1:]...)
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
func (c *SlackConnection) SetChannels(channels []gateway.Channel) {
	c.channels = channels
}
func (c *SlackConnection) Self() *gateway.User {
	return &c.self
}

type RawSlackMessage struct {
	Ts        string `json:"ts"`
	UserId    string `json:"user"`
	Text      string `json:"text"`
	Reactions []struct {
		Name  string   `json:"name"`
		Users []string `json:"users"`
	} `json:"reactions"`
	File struct {
		Id         string `json:"id"`
		Name       string `json:"name"`
		Filetype   string `json:"pretty_type"`
		User       string `json:"user"`
		PrivateUrl string `json:"url_private"`
		Permalink  string `json:"permalink"`
		Text       string `json:"plain_text"`
		Reactions  []struct {
			Name  string   `json:"name"`
			Users []string `json:"users"`
		} `json:"reactions"`
	} `json:"file,omitempty"`
	Attachments []struct {
		Title string `json:"title"`
		TitleLink string `json:"title_link"`
		Color string `json:"color"`
		Fields []struct {
			Title string `json:"title"`
			Value string `json:"value"`
			Short bool `json:"short"`
		} `json:"fields"`
	} `json:"attachments"`
}

func (c *SlackConnection) ParseMessage(
	preMessage map[string]interface{},
	cachedUsers map[string]*gateway.User,
) (*gateway.Message, error) {
	var slackMessageBuffer RawSlackMessage
	var intermediate []byte

	// First, convert the map to json.
	intermediate, err := json.Marshal(preMessage)
	if err != nil {
		return nil, err
	}

	// Then, marshal the json into the struct
	err = json.Unmarshal(intermediate, &slackMessageBuffer)
	if err != nil {
		return nil, err
	}

	// Get the sender of the message
	// Since we're likely to have a lot of the same users, cache them.
	var sender *gateway.User
	if cachedUsers[slackMessageBuffer.UserId] != nil {
		sender = cachedUsers[slackMessageBuffer.UserId]
	} else {
		sender, err = c.UserById(slackMessageBuffer.UserId)
		if err != nil {
			return nil, err
		}
		cachedUsers[slackMessageBuffer.UserId] = sender
	}

	// Convert the reactions fetched into reaction objects
	reactions := []gateway.Reaction{}
	reactionsLocation := slackMessageBuffer.Reactions
	if len(slackMessageBuffer.File.Reactions) > 0 {
		reactionsLocation = slackMessageBuffer.File.Reactions
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
	if len(slackMessageBuffer.File.Name) > 0 {
		// Given a user id, get a reference to the user.
		var fileUser *gateway.User
		if cachedUsers[slackMessageBuffer.File.User] != nil {
			fileUser = cachedUsers[slackMessageBuffer.File.User]
		} else {
			fileUser, err = c.UserById(slackMessageBuffer.File.User)
			if err != nil {
				return nil, err
			}
		}

		// Create the file struct representation.
		file = &gateway.File{
			Id:         slackMessageBuffer.File.Id,
			Name:       slackMessageBuffer.File.Name,
			Filetype:   slackMessageBuffer.File.Filetype,
			User:       fileUser,
			PrivateUrl: slackMessageBuffer.File.PrivateUrl,
			Permalink:  slackMessageBuffer.File.Permalink,
		}
	} else {
		file = nil
	}

	var attachments []gateway.Attachment
	if len(slackMessageBuffer.Attachments) > 0 {
		for _, attach := range slackMessageBuffer.Attachments {
			// Create attachment
			a := gateway.Attachment{
				Title: attach.Title,
				TitleLink: attach.TitleLink,
				Color: attach.Color,
			}

			// Add fields to attachment
			for _, field := range attach.Fields {
				a.Fields = append(a.Fields, gateway.AttachmentField{
					Title: field.Title,
					Value: field.Value,
					Short: field.Short,
				})
			}

			// Add attachment to list.
			attachments = append(attachments, a)
		}
	} else {
		attachments = nil
	}

	// Convert timestamp to float64
	// I would unmarshal directly into float64, but that doesn't work since slack encodes their
	// timestamps as strings :/
	var timestamp float64
	timestamp, err = strconv.ParseFloat(slackMessageBuffer.Ts, 64)
	if err != nil {
		return nil, err
	}

	return &gateway.Message{
		Sender:    sender,
		Text:      slackMessageBuffer.Text,
		Reactions: reactions,
		Timestamp: int(timestamp), // this value is in seconds!
		Hash:      slackMessageBuffer.Ts,
		File:      file,
		Attachments: &attachments,
	}, nil
}

func (c *SlackConnection) ToggleMessageReaction(message gateway.Message, reaction string) error {
	// Has the active user reacted to this message?
	messageReactedTo := false
	Outer:
	for _, r := range message.Reactions {
		if r.Name == reaction {
			for _, user := range r.Users {
				if c.Self().Name == user.Name {
					// This messag has already been reacted to by this user.
					messageReactedTo = true
					break Outer
				}
			}
		}
	}

	// Toggle the reaction on the message
	var reactionUrl string
	if messageReactedTo {
		log.Printf("Adding reaction to message %s: %s", message.Hash, reaction)
		reactionUrl = "https://slack.com/api/reactions.remove"
	} else {
		log.Printf("Removing reaction to message %s: %s", message.Hash, reaction)
		reactionUrl = "https://slack.com/api/reactions.add"
	}

	reactionUrl += "?token=" + c.token
	reactionUrl += "&name=" + reaction
	reactionUrl += "&channel=" + c.selectedChannel.Id
	reactionUrl += "&timestamp=" + message.Hash
	// If reacting to a message that has a file attached, pass to the file too.
	if message.File != nil {
		reactionUrl += "&file=" + message.File.Id
	}

	// Make the request
	resp, err := http.Get(reactionUrl)
	if err != nil {
		return err
	}

	// Fetch body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Verify the response
	var response struct {
		Ok bool `json:"ok"`
		Error string `json:"error"`
	}
	json.Unmarshal(body, &response)
	if response.Ok {
		return nil
	} else {
		return errors.New(fmt.Sprintf("Slack error: %s", response.Error))
	}
}

func (c *SlackConnection) Status() gateway.ConnectionStatus {
	return c.connectionStatus
}
