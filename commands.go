package main

import (
	"fmt"
	"io/ioutil"
	"errors"

	"github.com/skratchdot/open-golang/open"
)

type CommandType int

const (
	NATIVE CommandType = iota
	SLACK
)

type Command struct {
	Name         string
	Description  string
	Type         CommandType
	Permutations []string
	Arguments    string
	Handler      func(args []string, state *State) error
}

var COMMANDS = []Command{
	{
		Name:         "Quit",
		Type:         NATIVE,
		Description:  "Quits slime.",
		Permutations: []string{"quit", "q"},
		/* NO HANDLER, SPECIAL CASE */
	},

	//
	// POSTS
	//
	{
		Name:         "Post",
		Type:         NATIVE,
		Description:  "Make a post in the current channel with the contents of the specified file.",
		Arguments:    "<post file> [post name]",
		Permutations: []string{"post", "p"},
		Handler:      func(args []string, state *State) error {
			var postTitle string
			var postPath string
			if len(args) == 3 { // /post "post content" "post title"
				postPath = args[1]
				postTitle = args[2]
			} else if len(args) == 2 { // /post "post content"
				postPath = args[1]
			} else {
				return errors.New("Please use more arguments. /post path/to/post.txt [\"post title\"]")
			}

			// Read post from filesystem
			postContent, err := ioutil.ReadFile(postPath)
			if err != nil {
				return errors.New(fmt.Sprintf("Couldn't read file %s: %s", postPath, err.Error()))
			}

			// Make the post
			if err = state.ActiveConnection().PostText(postTitle, string(postContent)); err != nil {
				return err
			}

			return nil
		},
	},
	{
		Name:         "PostInline",
		Type:         NATIVE,
		Description:  "Make a post in the current channel using the given text.",
		Arguments:    "<post content> [post name]",
		Permutations: []string{"postinline", "pin"},
		Handler:      func(args []string, state *State) error {
			var postTitle string
			var postContent string
			if len(args) == 3 { // /postinline "post content" "post title"
				postContent = args[1]
				postTitle = args[2]
			} else if len(args) == 2 { // /postinline "post content"
				postContent = args[1]
			} else {
				return errors.New("Please use more arguments. /postinline \"post content\" [\"post title\"]")
			}

			// Make the post
			if err := state.ActiveConnection().PostText(postTitle, postContent); err != nil {
				return err
			}

			return nil
		},
	},


	//
	// MESSAGE ACTIONS
	//

	{
		Name: "Reaction",
		Type: NATIVE,
		Description: "Post a reaction to the selected message",
		Arguments: "<raection name>",
		Permutations: []string{"react"},
		Handler: func(args []string, state *State) error {
			selectedMessageIndex := len(state.ActiveConnection().MessageHistory()) - 1 - state.SelectedMessageIndex
			selectedMessage := state.ActiveConnection().MessageHistory()[selectedMessageIndex]

			if len(args) == 2 {
				reaction := args[1]
				return state.ActiveConnection().ToggleMessageReaction(selectedMessage, reaction)
			} else {
				return errors.New("Please use more arguments. /react <reaction name>")
			}
		},
	},
	{
		Name: "OpenFile",
		Type: NATIVE,
		Description: "If a file is attached to the current message, open it.",
		Arguments: "",
		Permutations: []string{"openfile", "opf"},
		Handler: func(args []string, state *State) error {
			selectedMessageIndex := len(state.ActiveConnection().MessageHistory()) - 1 - state.SelectedMessageIndex
			selectedMessage := state.ActiveConnection().MessageHistory()[selectedMessageIndex]

			// Open the private image url in the browser
			if selectedMessage.File != nil {
				open.Run(selectedMessage.File.Permalink)
			} else {
				return errors.New("Selected message has no file")
			}

			return nil
		},
	},
	{
		Name: "CopyFile",
		Type: NATIVE,
		Description: "If a file is attached to the current message, copy it into the system clipboard.",
		Arguments: "",
		Permutations: []string{"copyfile", "cpf"},
		Handler: func(args []string, state *State) error {
			selectedMessageIndex := len(state.ActiveConnection().MessageHistory()) - 1 - state.SelectedMessageIndex
			selectedMessage := state.ActiveConnection().MessageHistory()[selectedMessageIndex]

			// Open the private image url in the browser
			if selectedMessage.File != nil {
				state.Status.Printf("FIXME: will eventually copy to clipboard: %s", selectedMessage.File.Permalink)
			} else {
				return errors.New("Selected message has no file")
			}

			return nil
		},
	},



	//
	// BUILT INTO SLACK
	//
	{
		Type:         SLACK,
		Name:         "Apps",
		Permutations: []string{"apps"},
		Arguments:    "[search term]",
		Description:  "Search for Slack Apps in the App Directory",
	},
	{
		Type:         SLACK,
		Name:         "Away",
		Permutations: []string{"away"},
		Arguments:    "Toggle your away status",
	},
	{
		Type:         SLACK,
		Name:         "Call",
		Permutations: []string{"call"},
		Arguments:    "[help]",
		Description:  "Start a call",
	},
	{
		Type:         SLACK,
		Name:         "Dnd",
		Permutations: []string{"dnd"},
		Arguments:    "[some description of time]",
		Description:  "Starts or ends a Do Not Disturb session",
	},
	{
		Type:         SLACK,
		Name:         "Feed",
		Permutations: []string{"feed"},
		Arguments:    "help [or subscribe, list, remove...]",
		Description:  "Manage RSS subscriptions",
	},
	{
		Type:         SLACK,
		Name:         "Invite",
		Permutations: []string{"invite"},
		Arguments:    "@user [channel]",
		Description:  "Invite another member to a channel",
	},
	{
		Type:         SLACK,
		Name:         "Invite people",
		Permutations: []string{"invite_people"},
		Arguments:    "[name@example.com, ...]",
		Description:  "Invite people to your Slack team",
	},
	{
		Type:         SLACK,
		Name:         "Leave",
		Permutations: []string{"leave", "close", "part"},
		Description:  "Leave a channel",
	},
	{
		Type:         SLACK,
		Name:         "Me",
		Permutations: []string{"me"},
		Arguments:    "your message",
		Description:  "Displays action text",
	},
	{
		Type:         SLACK,
		Name:         "Msg",
		Permutations: []string{"msg", "dm"},
		Arguments:    "[your message]",
	},
	{
		Type:         SLACK,
		Name:         "Mute",
		Permutations: []string{"mute"},
		Arguments:    "[channel]",
		Description:  "Mutes [channel] or the current channel",
	},
	{
		Type:         SLACK,
		Name:         "Remind",
		Permutations: []string{"remind"},
		Arguments:    "[@someone or #channel] [what] [when]",
		Description:  "Set a reminder",
	},
	{
		Type:         SLACK,
		Name:         "Rename",
		Permutations: []string{"rename"},
		Arguments:    "[new name]",
		Description:  "Rename a channel",
	},
	{
		Type:         SLACK,
		Name:         "Shrug",
		Permutations: []string{"shrug"},
		Arguments:    "[your message]",
		Description:  "Appends ¯\\_(ツ)_/¯ to your message",
	},
	{
		Type:         SLACK,
		Name:         "Star",
		Permutations: []string{"star"},
		Arguments:    "Stars the current channel or conversation",
	},
	{
		Type:         SLACK,
		Name:         "Status",
		Permutations: []string{"status"},
		Arguments:    "[clear] or [:your_new_status_emoji:] [your new status message]",
		Description:  "Set or clear your custom status",
	},
	{
		Type:         SLACK,
		Name:         "Who",
		Permutations: []string{"who"},
		Description:  "List users in the current channel or group",
	},
}
