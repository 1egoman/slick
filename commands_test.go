package main_test

import (
	. "github.com/1egoman/slime"
	"github.com/1egoman/slime/gateway"
	"github.com/1egoman/slime/gateway/slack"
	"github.com/jarcoal/httpmock"
	"testing"
	"errors"
)

func findCommand(name string) *Command {
	// Find post command
	for _, command := range COMMANDS {
		if command.Name == name {
			return &command
		}
	}
	return nil
}

func TestCommandPost(t *testing.T) {
	defer httpmock.DeactivateAndReset()

	for _, test := range []struct{
		Name string
		Command string
		Args []string
		Url string
		Response string
		Error error
	}{
		{
			Name: `/postinline "post body" works`,
			Command: "PostInline",
			Args: []string{"postinline", "post body"},
			Url: "https://slack.com/api/files.upload?token=token&channels=channel-id&content=post+body",
			Response: `{"ok": true}`,
		},
		{
			Name: `/postinline "post body" "post title" works`,
			Command: "PostInline",
			Args: []string{"postinline", "post body", "post title"},
			Url: "https://slack.com/api/files.upload?token=token&channels=channel-id&content=post+body&title=post+title",
			Response: `{"ok": true}`,
		},
		{
			Name: `/postinline alone doesn't work`,
			Command: "PostInline",
			Args: []string{"postinline"},
			Error: errors.New("Please use more arguments. /postinline \"post content\" [\"post title\"]"),
		},
		{
			Name: `If /postinline throws underlying errors we know about it`,
			Command: "PostInline",
			Args: []string{"postinline", "post body", "post title"},
			Url: "https://slack.com/api/files.upload?token=token&channels=channel-id&content=post+body&title=post+title",
			Response: `{"ok": true, "error": "My Error"}`,
			Error: errors.New("My Error"),
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
		command := findCommand(test.Command)
		err := command.Handler(test.Args, initialState)

		httpmock.DeactivateAndReset()

		// Verify the output
		if !((test.Error == nil && err == nil) || (test.Error.Error() == err.Error())) {
			t.Errorf("Test %s failed: %s", test.Name, err)
		}
	}
}
