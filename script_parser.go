package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"errors"
	"reflect"

	"github.com/1egoman/slick/frontend"
	"github.com/1egoman/slick/gateway"

	"github.com/atotto/clipboard"
	"github.com/cjoudrey/gluahttp" // gopher-lua http library
	"github.com/yuin/gopher-lua"
)

type SlickEvent int

const (
	EVENT_KEYMAP SlickEvent = iota
	EVENT_CONNECTION_CHANGE
	EVENT_COMMAND_RUN
	EVENT_MESSAGE_SENT
	EVENT_MESSAGE_RECEIVED
	EVENT_MODE_CHANGE
)

type EventAction struct {
	Type    SlickEvent
	Key     []rune
	Handler func(*State, *map[string]string) error
}

// Send an event to all the stored listeners.
func EmitEvent(state *State, event SlickEvent, metadata map[string]string) error {
	log.Printf("Event emitted: %+v %+v", event, metadata)
	for _, i := range state.EventActions {
		if i.Type == event {
			if err := i.Handler(state, &metadata); err != nil {
				return err
			}
		}
	}

	return nil
}

func AddSlickStandardLib(L *lua.LState, state *State, term *frontend.TerminalDisplay) {
	// Add some logging utilities
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		state.Status.Printf(L.ToString(1))
		render(state, term)
		return 0
	}))
	L.SetGlobal("error", L.NewFunction(func(L *lua.LState) int {
		state.Status.Errorf(L.ToString(1))
		render(state, term)
		return 0
	}))
	L.SetGlobal("clear", L.NewFunction(func(L *lua.LState) int {
		state.Status.Clear()
		render(state, term)
		return 0
	}))

	// Allow lua to run things when a user presses a key.
	L.SetGlobal("keymap", L.NewFunction(func(L *lua.LState) int {
		key := L.ToString(1)
		function := L.ToFunction(2)

		state.EventActions = append(state.EventActions, EventAction{
			Type: EVENT_KEYMAP,
			Key:  []rune(key),
			Handler: func(state *State, metadata *map[string]string) error {
				return L.CallByParam(lua.P{Fn: function, NRet: 0})
			},
		})
		return 0
	}))

	L.SetGlobal("command", L.NewFunction(func(L *lua.LState) int {
		name := L.ToString(1)
		callback := L.ToFunction(4)
		COMMANDS = append(COMMANDS, Command{
			Name:         name,
			Type:         NATIVE,
			Description:  L.ToString(2),
			Arguments:    L.ToString(3),
			Permutations: []string{name},
			Handler: func(args []string, state *State) error {
				log.Println("Running lua command", name, args)
				// Convert arguments slice into table
				luaArgs := L.NewTable()
				for _, arg := range args {
					luaArgs.Append(lua.LString(arg))
				}

				return L.CallByParam(lua.P{Fn: callback, NRet: 0}, luaArgs)
			},
		})
		return 0
	}))

	L.SetGlobal("getenv", L.NewFunction(func(L *lua.LState) int {
		envName := L.ToString(1)
		L.Push(lua.LString(os.Getenv(envName)))
		return 1
	}))

	L.SetGlobal("shell", L.NewFunction(func(L *lua.LState) int {
		commandName := L.ToString(1)
		if len(commandName) == 0 {
			L.Push(lua.LString("First argument (command name) is required."))
			return 1
		}

		var args []string
		argc := 2
		for ; ; argc++ {
			arg := L.ToString(argc)
			if len(arg) > 0 {
				args = append(args, arg)
			} else {
				break
			}
		}
		log.Println("Running command", commandName, "with args", args)

		command := exec.Command(commandName, args...)
		output, err := command.Output()
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}

		log.Println("Command output", output)
		L.Push(lua.LNil)
		L.Push(lua.LString(string(output)))
		return 1
	}))

	L.SetGlobal("sendmessage", L.NewFunction(func(L *lua.LState) int {
		messageText := L.ToString(1)
		if len(messageText) == 0 {
			L.Push(lua.LString("First argument (message text) is required."))
			return 1
		}

		// Just send a normal message!
		message := gateway.Message{
			Sender: state.ActiveConnection().Self(),
			Text:   messageText,
		}

		// Sometimes, a message could have a response. This is for example true in the
		// case of slash commands, sometimes.
		_, err := state.ActiveConnection().SendMessage(
			message,
			state.ActiveConnection().SelectedChannel(),
		)

		if err != nil {
			L.Push(lua.LString("Error sending message: " + err.Error()))
		} else {
			L.Push(lua.LNil)
		}

		return 1
	}))

	L.SetGlobal("getclip", L.NewFunction(func(L *lua.LState) int {
		text, err := clipboard.ReadAll()
		if err != nil {
			L.Push(lua.LString(text))
			L.Push(lua.LString(err.Error()))
			return 2
		} else {
			L.Push(lua.LString(text))
			L.Push(lua.LNil)
			return 2
		}
	}))

	L.SetGlobal("setclip", L.NewFunction(func(L *lua.LState) int {
		clipboard.WriteAll(L.ToString(1))
		return 0
	}))

	// Allow lua to run things when a user presses a key.
	L.SetGlobal("onevent", L.NewFunction(func(L *lua.LState) int {
		eventString := L.ToString(1)
		function := L.ToFunction(2)

		// Convert the string typed by the user into an kj
		var event SlickEvent
		switch eventString {
		/* no EVENT_KEYMAP */
		case "connectionchange":
			event = EVENT_CONNECTION_CHANGE
		case "commandrun":
			event = EVENT_COMMAND_RUN
		case "messagesent":
			event = EVENT_MESSAGE_SENT
		case "messagereceived":
			event = EVENT_MESSAGE_RECEIVED
		case "modechange":
			event = EVENT_MODE_CHANGE
		}

		state.EventActions = append(state.EventActions, EventAction{
			Type: event,
			Handler: func(state *State, metadata *map[string]string) error {
				if metadata == nil {
					return errors.New("Metadata passed to event handler for " + eventString + " was nil.")
				}

				// Convert metadata into a table
				metadataTable := L.NewTable()
				for key, value := range *metadata {
					metadataTable.RawSet(lua.LString(key), lua.LString(value))
				}

				// Call into lua with with metadata
				return L.CallByParam(lua.P{Fn: function, NRet: 0}, metadataTable)
			},
		})
		return 0
	}))

	// Load Gluahttp so the config can make http requests: https://github.com/cjoudrey/gluahttp
	L.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{}).Loader)

	// Export all commands in the lua context
	for _, command := range COMMANDS {
		func(command Command) { // Close over command so it
			L.SetGlobal(command.Name, L.NewFunction(func(L *lua.LState) int {
				// Collect all arguments into an array
				args := []string{"__COMMAND"}
				argc := 1
				for ; ; argc += 1 {
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
					L.Push(lua.LString("No handler defined for the command " + command.Name + "."))
					return 1
				}

				err := command.Handler(args, state)

				if err == nil {
					L.Push(lua.LNil)
				} else {
					L.Push(lua.LString(err.Error()))
				}

				render(state, term)
				return 1
			}))
		}(command)
	}
}

// Given a struct, convert it to a lua table.
func StructToTable(L *lua.LState, s interface{}) *lua.LTable {
	tbl := L.NewTable()
    v := reflect.ValueOf(s)

	// For each field, try to add to the lua table
	var luaValue lua.LValue
	for i := 0; i < v.NumField(); i++ {
		// Start value at nil for each iteration
		luaValue = lua.LNil

		key := v.Type().Field(i).Name
		value := v.Field(i).Interface()

		typ := v.Field(i).Type().Kind()

		// Dereference pointers into their native types.
		var deref reflect.Value = v.Field(i)
		for typ == reflect.Ptr {
			deref = reflect.Indirect(deref)
			typ = deref.Type().Kind()
			value = deref.Interface()
		}

		switch typ {
		case reflect.String:
			luaValue = lua.LString(value.(string))
		case reflect.Int:
			luaValue = lua.LNumber(value.(int))
		case reflect.Bool:
			luaValue = lua.LBool(value.(bool))
		case reflect.Struct:
			luaValue = StructToTable(L, value)
		}

		// Add the value to the lua table, if it could be successfully converted. Otherwise, use
		// nil.
		tbl.RawSetString(key, luaValue)
	}

	return tbl
}

func ParseScript(script string, state *State, term *frontend.TerminalDisplay) error {
	L := lua.NewState()
	defer L.Close()

	AddSlickStandardLib(L, state, term)
	return L.DoString(script)
}
