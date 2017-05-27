package main_test

import (
	"errors"
	. "github.com/1egoman/slick"
	"github.com/1egoman/slick/gateway"
	"github.com/1egoman/slick/gateway/slack"
	"github.com/jarcoal/httpmock"
	"testing"
)

func TestHttpBasedCommands(t *testing.T) {
	defer httpmock.DeactivateAndReset()

	for _, test := range []struct {
		Name     string
		Command  string
		Args     []string
		Url      string
		Response string
		Error    error
	}{
		{
			Name:     `/postinline "post body" works`,
			Command:  "PostInline",
			Args:     []string{"postinline", "post body"},
			Url:      "https://slack.com/api/files.upload?token=token&channels=channel-id&content=post+body",
			Response: `{"ok": true}`,
		},
		{
			Name:     `/postinline "post body" "post title" works`,
			Command:  "PostInline",
			Args:     []string{"postinline", "post body", "post title"},
			Url:      "https://slack.com/api/files.upload?token=token&channels=channel-id&content=post+body&title=post+title",
			Response: `{"ok": true}`,
		},
		{
			Name:    `/postinline alone doesn't work`,
			Command: "PostInline",
			Args:    []string{"postinline"},
			Error:   errors.New("Please use more arguments. /postinline \"post content\" [\"post title\"]"),
		},
		{
			Name:     `If /postinline throws underlying errors we know about it`,
			Command:  "PostInline",
			Args:     []string{"postinline", "post body", "post title"},
			Url:      "https://slack.com/api/files.upload?token=token&channels=channel-id&content=post+body&title=post+title",
			Response: `{"ok": true, "error": "My Error"}`,
			Error:    errors.New("My Error"),
		},
	} {
		// Listen for command response
		httpmock.Activate()
		httpmock.RegisterResponder("GET", test.Url, httpmock.NewStringResponder(200, test.Response))

		// Create initial state
		initialState := NewInitialStateMode("writ")
		initialState.Connections = []gateway.Connection{
			gatewaySlack.New("token"),
		}
		initialState.ActiveConnection().SetSelectedChannel(&gateway.Channel{Id: "channel-id"})

		// Execute the command
		command := GetCommand(test.Command)
		err := command.Handler(test.Args, initialState)

		httpmock.DeactivateAndReset()

		// Verify the output
		if !((test.Error == nil && err == nil) || (test.Error.Error() == err.Error())) {
			t.Errorf("Test %s failed: %s", test.Name, err)
		}
	}
}

// Test `/connect token-here` and `/disconnect`
func TestCommandConnectDisconnect(t *testing.T) {
	defer httpmock.DeactivateAndReset()

	// Listen for command response
	// FIXME: we should spin up a local websocket server here and not use one on the internet.
	httpmock.Activate()
	httpmock.RegisterResponder("GET", "https://slack.com/api/rtm.start?token=token",
		httpmock.NewStringResponder(200, `{"ok": true, "url": "wss://echo.websocket.org/?encoding=text"}`))

	// Create initial state
	state := NewInitialStateMode("writ")
	state.Connections = []gateway.Connection{
		gatewaySlack.New("token"),
	}
	state.ActiveConnection().SetSelectedChannel(&gateway.Channel{Id: "channel-id"})

	// Execute the command
	command := *GetCommand("Connect")
	err := RunCommand(command, []string{"connect", "token"}, state)

	// Verify the output
	if err != nil {
		t.Errorf("Couldn't connect to mock slack: %s", err)
	}

	// Attempt to disconnect from the socket.
	command = *GetCommand("Disconnect")
	err = RunCommand(command, []string{"disconnect"}, state)
	if err != nil {
		t.Errorf("Couldn't disconnect from mock slack: %s", err)
	}
}

// Test `/connect "team name" token-here`
func TestCommandConnectWithName(t *testing.T) {
	defer httpmock.DeactivateAndReset()

	// Listen for command response
	// FIXME: we should spin up a local websocket server here and not use one on the internet.
	httpmock.Activate()
	httpmock.RegisterResponder("GET", "https://slack.com/api/rtm.start?token=token",
		httpmock.NewStringResponder(200, `{"ok": true, "url": "wss://echo.websocket.org/?encoding=text"}`))

	// Create initial state
	state := NewInitialStateMode("writ")
	state.Connections = []gateway.Connection{
		gatewaySlack.New("token"),
	}
	state.ActiveConnection().SetSelectedChannel(&gateway.Channel{Id: "channel-id"})

	// Execute the command
	command := *GetCommand("Connect")
	err := RunCommand(command, []string{"connect", "team name", "token"}, state)

	// Verify the output
	if err != nil {
		t.Errorf("Couldn't connect to mock slack: %s", err)
	}

	if name := state.ActiveConnection().Name(); name != "team name" {
		t.Errorf("Invalid name for slack team: %s", name)
	}
}
