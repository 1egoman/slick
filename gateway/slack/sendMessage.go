package gatewaySlack

import (
	"log"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"strings"

	"github.com/1egoman/slime/gateway"
)


func sendSlashCommand(c *SlackConnection, message gateway.Message, channel *gateway.Channel) (*gateway.Message, error) {
  log.Printf("Sending slash command to team %s on channel %s", c.Team().Name)

  // If the message starts with a slash, it's a slash command.
  command := strings.Split(message.Text, " ")
  text := url.QueryEscape(strings.Join(command[1:], " "))
  resp, err := http.Get("https://slack.com/api/chat.command?token=" + c.token + "&channel=" + channel.Id + "&command=" + url.QueryEscape(command[0]) + "&text=" + text)
  if err != nil {
    return nil, err
  }

  body, _ := ioutil.ReadAll(resp.Body)
  var commandResponse struct {
    Response string `json:"response"`
  }
  err = json.Unmarshal(body, &commandResponse)
  if err != nil {
    return nil, err
  }

  // Return a response message if the response 
  if len(commandResponse.Response) > 0 {
    return &gateway.Message{
      Text: commandResponse.Response,
      Sender: &gateway.User{Name: "slackbot"},
    }, nil
  } else {
    return nil, nil
  }
}

// Send a given message to a given channel. Also, is able to process slash commands.
// Returns an optional pointer to a response message and an error.
func (c *SlackConnection) SendMessage(message gateway.Message, channel *gateway.Channel) (*gateway.Message, error) {
	if strings.HasPrefix(message.Text, "/") {
    // Slash commands require some preprocessing.
    return sendSlashCommand(c, message, channel)
	} else {
		log.Printf("Sending message to team %s on channel %s", c.Team().Name, channel.Name)

		// Otherwise just a plain message
		_, err := http.Get("https://slack.com/api/chat.postMessage?token=" + c.token + "&channel=" + channel.Id + "&text=" + url.QueryEscape(message.Text) + "&link_names=true&parse=full&unfurl_links=true&as_user=true")
		return nil, err
	}
}
