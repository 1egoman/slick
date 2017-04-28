package gatewaySlack

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/1egoman/slime/gateway"
	"golang.org/x/net/websocket"
)

// Connect to the slack persistent socket.
func (c *SlackConnection) Connect() error {
	log.Println("Requesting slack team connection url...")
	// Create buffered channels to listen and send messages on
	c.incoming = make(chan gateway.Event, 1)
	c.outgoing = make(chan gateway.Event, 1)

	// Request a connection url with the token in the struct
	c.requestConnectionUrl()
	log.Printf("Got slack connection url for team %s: %s", c.Team().Name, c.url)

	// FIXME: what does this mean?
	origin := "http://localhost/"

	// Create a connection to the websocket
	var err error
	c.conn, err = websocket.Dial(c.url, "", origin)
	if err != nil {
		return err
	}
	log.Printf("Slack connection %s made!", c.Team().Name)

	// When messages are received, add them to the incoming buffer.
	go func(incoming chan gateway.Event) {
		var msgRaw = make([]byte, 512)
		var msg map[string]interface{}
		var n int

		for {
			// Listen for messages, and when some are received, write them to a channel.
			if n, err = c.conn.Read(msgRaw); err != nil {
				log.Fatal(err)
			}

			// Decode into a struct so that we can check message type later
			json.Unmarshal(msgRaw[:n], &msg)
			log.Printf("INCOMING %s: %s", c.Team().Name, msgRaw[:n])
			incoming <- gateway.Event{
				Direction: "incoming",
				Type:      msg["type"].(string),
				Data:      msg,
			}
		}
	}(c.incoming)

	// When messages are in the outgoing buffer waiting to be sent, send them.
	go func(outgoing chan gateway.Event) {
		// Add a sequential message id to each message sent, so replies can later be tracked.
		messageId := 0

		var event gateway.Event
		for {
			// Assemble the message to send.
			event = <-outgoing
			event.Data["type"] = event.Type
			messageId++
			event.Data["id"] = messageId

			// Marshal to json
			data, err := json.Marshal(event.Data)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("OUTGOING %s: %s", c.Team().Name, data)

			// Send it.
			if _, err = c.conn.Write(data); err != nil {
				log.Fatal(err)
			}
		}
	}(c.outgoing)

	return nil
}

func (c *SlackConnection) requestConnectionUrl() {
	// Make request to slack's api to get websocket credentials
	// https://api.slack.com/methods/rtm.connect
	resp, err := http.Get("https://slack.com/api/rtm.connect?token=" + c.token)

	if err != nil {
		log.Fatal(err)
	}

	// Decode json body.
	body, _ := ioutil.ReadAll(resp.Body)
	var connectionBuffer struct {
		Ok   bool         `json:"ok"`
		Url  string       `json:"url"`
		Team gateway.Team `json:"team"`
		Self gateway.User `json:"self"`
	}
	err = json.Unmarshal(body, &connectionBuffer)
	if err != nil {
		log.Fatal("Slack response: " + string(body))
	}

	// Add response data to struct
	c.url = connectionBuffer.Url
	c.self = connectionBuffer.Self
	c.team = connectionBuffer.Team
}

