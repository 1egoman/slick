package main_test

import (
	"errors"
	. "github.com/1egoman/slick"
	"github.com/1egoman/slick/gateway"
	"github.com/1egoman/slick/gateway/slack"
	"github.com/jarcoal/httpmock"
	"testing"
)

func TestRunCommand(t *testing.T) {
	var command Command
	var err error
	state := NewInitialStateMode("writ")

	// Test a native command
	wasCalled := false
	command = Command{
		Name:         "My Command",
		Description:  "Command Description",
		Type:         NATIVE,
		Permutations: []string{"foo", "bar"},
		Arguments:    "<required> [optional]",
		Handler: func(args []string, state *State) error {
			wasCalled = true
			return nil
		},
	}
	err = RunCommand(command, []string{"foo"}, state)
	if wasCalled && err != nil {
		t.Errorf("Error in running valid command: %s (wascalled = %v)", err, wasCalled)
	}

	// Test a native command without a handler
	command = Command{
		Name:         "My Command",
		Description:  "Command Description",
		Type:         NATIVE,
		Permutations: []string{"foo", "bar"},
		Arguments:    "<required> [optional]",
		Handler:      nil,
	}
	err = RunCommand(command, []string{"foo"}, state)
	if err.Error() != "The command My Command doesn't have an associated handler function." {
		t.Errorf("Wrong error in running nil handler command: %s", err)
	}
}

func TestGetCommand(t *testing.T) {
	var command *Command

	// Test a native command
	command = GetCommand("Connect")
	if command.Name != "Connect" {
		t.Errorf("Invalid command found with GetCommand: %+v", command)
	}

	// Test a native command
	command = GetCommand("CommandThatDoesntExist")
	if command != nil {
		t.Errorf("Unexpected command found by GetCommand: %+v", command)
	}
}

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

func TestCommandPick(t *testing.T) {
	// Create initial state
	state := NewInitialStateMode("writ")
	state.Connections = []gateway.Connection{
		gatewaySlack.NewWithName("team name", "token"),
	}

	channels := []gateway.Channel{
		gateway.Channel{Name: "channel name", Id: "channel-id"},
		gateway.Channel{Name: "other channel", Id: "other-channel-id"},
	}
	state.ActiveConnection().SetChannels(channels)
	state.ActiveConnection().SetSelectedChannel(&channels[1])

	// Execute the command
	command := *GetCommand("Pick")
	err := RunCommand(command, []string{"pick", "team name", "channel name"}, state)

	// Verify the output
	if err != nil {
		t.Errorf("Couldn't pick another team:", err)
	}

	if name := state.ActiveConnection().Name(); name != "team name" {
		t.Errorf("Didn't choose slack team 'team name', used %s instead.", name)
	}
	if channel := state.ActiveConnection().SelectedChannel().Name; channel != "channel name" {
		t.Errorf("Didn't choose slack channel 'channel name', used %s instead.", channel)
	}
}
func TestCommandPickBadArgs(t *testing.T) {
	// Create initial state
	state := NewInitialStateMode("writ")

	// Execute the command, with
	command := *GetCommand("Pick")
	err := RunCommand(command, []string{"pick" /* team name, channel name */}, state)

	// Verify the output
	if err.Error() != "Please specify more args. /pick <connection name> <channel name>" {
		t.Errorf("Bad pick command args didn't emit right error: %s", err)
	}
}

func TestCommandExpandAttachment(t *testing.T) {
	// Create initial state
	state := NewInitialStateMode("writ")
	state.Connections = []gateway.Connection{
		gatewaySlack.NewWithName("team name", "token"),
	}
	channels := []gateway.Channel{gateway.Channel{Name: "channel name", Id: "channel-id"}}

	state.ActiveConnection().SetChannels(channels)
	state.ActiveConnection().SetSelectedChannel(&channels[0])
	state.ActiveConnection().SetMessageHistory([]gateway.Message{
		gateway.Message{
			Text: "My Message",
			Attachments: &[]gateway.Attachment{
				gateway.Attachment{
					Title: "title",
					Body: "body",
				},
			},
		},
	})

	// Execute the command, with
	command := *GetCommand("ExpandAttachment")
	err := RunCommand(command, []string{"expandattachment", "1"}, state)

	// Verify the output
	if err != nil {
		t.Errorf("Failed to open modal for expanding attachment: %s", err)
	}

	if state.Mode != "modl" {
		t.Errorf("Mode not set to `modl`: %s", state.Mode)
	}
	if state.Modal.Title != "title" || state.Modal.Body != "body" {
		t.Errorf("Modal title and body not set to the right values: %+v", state.Modal)
	}
}
