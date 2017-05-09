package main

import (
	"fmt"
	"io/ioutil"
	"errors"
	"strings"
	"log"

	"github.com/1egoman/slime/gateway"

	"github.com/skratchdot/open-golang/open"
	"github.com/yuin/gopher-lua"
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
	// OPEN IN SLACK
	//
	{
		Name: "OpenInSlack",
		Type: NATIVE,
		Description: "Open the current channel in the slack app.",
		Arguments: "",
		Permutations: []string{"slack", "ops"},
		Handler: func(args []string, state *State) error {
			team := state.ActiveConnection().Team()
			channel := state.ActiveConnection().SelectedChannel()
			if team == nil || channel == nil {
				return errors.New("Selected team or channel was nil.")
			}
			open.Run("slack://channel?team="+team.Id+"&id="+channel.Id)
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

func GetCommand(name string) *Command {
	for _, command := range COMMANDS {
		if command.Name == name {
			return &command
		}
	}
	return nil
}

// Given a command, a list of arguments to pass to the command, and the state, run the command.
func RunCommand(command Command, args []string, state *State) error {
	if command.Type == NATIVE && command.Handler != nil {
		return command.Handler(args, state)
	} else if command.Type == NATIVE && command.Handler == nil {
		return errors.New(fmt.Sprintf("The command %s doesn't have an associated handler function.", command.Name))
	} else {
		message := gateway.Message{
			Sender: state.ActiveConnection().Self(),
			Text:   fmt.Sprintf("/%s %s", command.Permutations[0], strings.Join(args, " ")),
		}

		// Sometimes, a message could have a response. This is for example true in the
		// case of slash commands, sometimes.
		responseMessage, err := state.ActiveConnection().SendMessage(
			message,
			state.ActiveConnection().SelectedChannel(),
		)

		if err != nil {
			return err
		} else if responseMessage != nil {
			// Got a response command? Append it to the message history.
			state.ActiveConnection().AppendMessageHistory(*responseMessage)
		}
		return nil
	}
}

type KeyAction struct {
	Key []rune
	Handler func(*State) error
}

func ParseScript(script string, state *State) error {
	L := lua.NewState()
	defer L.Close()

	// Add some logging utilities
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		state.Status.Printf(L.ToString(1))
		return 0
	}))
	L.SetGlobal("error", L.NewFunction(func(L *lua.LState) int {
		state.Status.Errorf(L.ToString(1))
		return 0
	}))
	L.SetGlobal("clear", L.NewFunction(func(L *lua.LState) int {
		state.Status.Clear()
		return 0
	}))

	// Allow lua to run things when a user presses a key.
	L.SetGlobal("keymap", L.NewFunction(func(L *lua.LState) int {
		key := L.ToString(1)
		function := L.ToFunction(2)

		state.KeyActions = append(state.KeyActions, KeyAction{
			Key: []rune(key),
			Handler: func(state *State) error {
				return L.CallByParam(lua.P{Fn: function, NRet: 0})
			},
		})
		return 0
	}))

	// Export all commands in the lua context
	for _, command := range COMMANDS {
		func(command Command) { // Close over command so it 
			L.SetGlobal(command.Name, L.NewFunction(func(L *lua.LState) int {
				// Collect all arguments into an array
				args := []string{}
				argc := 1
				for ;; argc += 1 {
					arg := L.ToString(argc)
					if len(arg) > 0 {
						args = append(args, arg)
					} else {
						break
					}
				}

				log.Printf("* Running command %s with args %v", command.Name, args)
				command := GetCommand(command.Name)
				if command.Handler == nil {
					L.Push(lua.LString("No handler defined for the command "+command.Name+"."))
					return 1
				}

				err := command.Handler(args, state)

				if err == nil {
					L.Push(lua.LNil)
				} else {
					L.Push(lua.LString(err.Error()))
				}
				return 1
			}))
		}(command)
	}

	return L.DoString(script)
}
