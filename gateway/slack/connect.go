package gatewaySlack

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"errors"

	"github.com/1egoman/slick/gateway"
	"golang.org/x/net/websocket"
)

const pingFrequency = 30 // In seconds

// Connect to the slack persistent socket.
func (c *SlackConnection) Connect() error {
	c.connectionStatus = gateway.CONNECTING

	// Create buffered channels to listen and send messages on
	c.incoming = make(chan gateway.Event, 10)
	c.outgoing = make(chan gateway.Event, 10)

	// Request a connection url with the token in the struct
	log.Println("Requesting slack team connection url...")
	var err error
	err = c.requestConnectionUrl()
	if err != nil {
		log.Println("Error getting conenction url", err)
		return err
	}
	log.Printf("Got slack connection url for team %s: %s", c.Team().Name, c.url)

	// FIXME: what does this mean?
	origin := "http://localhost/"

	// Create a connection to the websocket
	c.conn, err = websocket.Dial(c.url, "", origin)
	if err != nil {
		return err // Panic when cannot talk to messaging servers
	}
	log.Printf("Slack connection %s made!", c.Team().Name)

	// When messages are received, add them to the incoming buffer.
	go func(incoming chan gateway.Event) {
		var msgRaw = make([]byte, 512)
		var msg map[string]interface{}
		var n int
		var messageBuffer []byte

		for {
			// Listen for messages, and when some are received, write them to a channel.
			if n, err = c.conn.Read(msgRaw); err != nil {
				log.Println("Error reading from websocket:", err)
				if c.Status() != gateway.DISCONNECTED {
					log.Println("Error reading from slack socket", err.Error())
					c.connectionStatus = gateway.FAILED

					// Try to recover!
					// If we were disconencted from the slack socket, then attempt to reconnect.
					// This can happen beacause of rate limiting, timeouts, etc...
					err = c.requestConnectionUrl()
					if err != nil {
						log.Println("Error getting connection url", err)
						return
					}
					c.conn, err = websocket.Dial(c.url, "", origin)
					if err != nil {
						log.Println("Error reading from slack socket for second time", err.Error())
						return
					} else {
						log.Println("Was able to reconnect to slack socket!")
						if n, err = c.conn.Read(msgRaw); err != nil {
							log.Println("Error reading from slack socket for second time", err.Error())
							return
						} else {
							log.Println("Successful read, recovered from socket error!")
							c.connectionStatus = gateway.CONNECTED
						}
					}
				}
			}

			// Add the latest packet to the message buffer
			messageBuffer = append(messageBuffer, msgRaw[:n]...)

			// Decode message buffer into a struct so that we can check message type later
			err = json.Unmarshal(messageBuffer, &msg)
			if err == nil {
				// Clear the message buffer after unpacking
				messageBuffer = []byte{}

				log.Printf("INCOMING %s: %s", c.Team().Name, msgRaw[:n])
				incoming <- gateway.Event{
					Direction: "incoming",
					Type:      msg["type"].(string),
					Data:      msg,
				}
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
				if c.Status() != gateway.DISCONNECTED {
					log.Println("Error writing to slack socket", err.Error())
					c.connectionStatus = gateway.FAILED

					// Try to recover!
					// If we were disconencted from the slack socket, then attempt to reconnect.
					// This can happen beacause of rate limiting, timeouts, etc...
					err = c.requestConnectionUrl()
					if err != nil {
						log.Println("Error getting connection url", err)
						return
					}
					c.conn, err = websocket.Dial(c.url, "", origin)
					if err != nil {
						log.Println("Error connecting to slack socket for second time", err.Error())
						return
					} else {
						log.Println("Was able to reconnect to slack socket!")
						if _, err = c.conn.Write(data); err != nil {
							log.Println("Error writing to slack socket for second time", err.Error())
							return
						} else {
							log.Println("Successful write, recovered from socket error!")
							c.connectionStatus = gateway.CONNECTED
						}
					}
				}
			}
		}
	}(c.outgoing)

	// Periodically ping slack
	// This is to ensure slack doesn't think we stopped listening on the socket
	go func(outgoing chan gateway.Event) {
		pingCount := 0
		for {
			time.Sleep(pingFrequency * time.Second)
			if c.Status() == gateway.CONNECTED {
				// Send a ping
				outgoing <- gateway.Event{
					Type: "ping",
					Data: map[string]interface{}{"count": pingCount},
				}
				pingCount++
			} else if c.Status() == gateway.DISCONNECTED {
				// If the gateway disconnected, then return
				break
			}
		}
	}(c.outgoing)

	c.connectionStatus = gateway.CONNECTED
	return nil
}

func (c *SlackConnection) requestConnectionUrl() error {
	// Make request to slack's api to get websocket credentials
	// https://api.slack.com/methods/rtm.connect
	resp, err := http.Get("https://slack.com/api/rtm.connect?token=" + c.token)

	if err != nil {
		return err
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
		log.Println("Slack response: " + string(body))
		return errors.New("Slack connection error: "+string(body))
	}

	// Add response data to struct
	c.url = connectionBuffer.Url
	c.self = connectionBuffer.Self
	c.team = connectionBuffer.Team

	return nil
}
