# Slack Gateway

All code that communicates between slick and slack lives here.

## Constructing
```go
slack := gatewaySlack.New("my slack token here")
```

## Sending / Receiving Messages

Most of the code in `connect.go` has to do with sending / receiving messages. There's `.Incoming()`
and `.Outgoing()` methods on the struct that return a reference to the raw message sending channel.
For example, you can send a message like:

```go
slack.Outgoing() <- gateway.Event{
	Direction: "outgoing",
	Type: "message_type",
	Data: map[string]interface{}{
		"foo": "bar",
	},
}
```

and receive a message like:

```go
message := <-slack.Incoming()
message.Type // "message_type"
message.Direction // "incoming"
message.Data // map[string]interface{}{ "foo": "bar" }
```
