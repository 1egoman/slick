package main_test

import (
	"runtime"
	"errors"
	. "github.com/1egoman/slick"
	"github.com/1egoman/slick/frontend"
	"github.com/yuin/gopher-lua"
	"testing"
)

func TestEmitEventNoError(t *testing.T) {
	state := NewInitialStateMode("chat")
	state.EventActions = append(state.EventActions, EventAction{
		Type: EVENT_CONNECTION_CHANGE,
		Handler: func(state *State, metadata *map[string]string) error {
			return nil
		},
	})
	err := EmitEvent(state, EVENT_CONNECTION_CHANGE, map[string]string{})
	if err != nil {
		t.Errorf("Error was returned when emitting event: %s", err)
	}
}

func TestEmitEventError(t *testing.T) {
	state := NewInitialStateMode("chat")
	state.EventActions = append(state.EventActions, EventAction{
		Type: EVENT_CONNECTION_CHANGE,
		Handler: func(state *State, metadata *map[string]string) error {
			return errors.New("my error")
		},
	})
	err := EmitEvent(state, EVENT_CONNECTION_CHANGE, map[string]string{})
	if err.Error() != "my error" {
		t.Errorf("Error 'my error' was not returned when emitting event: %s", err)
	}
}

var noopFunction lua.LFunction = lua.LFunction{}

var scriptGlobals map[string][]interface{} = map[string][]interface{}{
	"print":       []interface{}{"foo"},
	"error":       []interface{}{"bar"},
	"clear":       []interface{}{},
	"keymap":      []interface{}{"k", noopFunction},
	"command":     []interface{}{"CmdName", "desc", "[one] <two>", noopFunction},
	"getenv":      []interface{}{"FOO"},
	"shell":       []interface{}{"date"},
	"sendmessage": []interface{}{"foo"},
	"getclip":     []interface{}{},
	"setclip":     []interface{}{},
	"onevent":     []interface{}{"connectionchange", noopFunction},
}

func TestScriptEnvironmentConstruction(t *testing.T) {
	state := NewInitialStateMode("chat")
	term := frontend.NewTerminalDisplay(nil)

	L := lua.NewState()
	defer L.Close()

	AddSlickStandardLib(L, state, term)

	for global, _ := range scriptGlobals {
		if L.GetGlobal(global) == nil {
			t.Errorf("Global `%s` isn't defined", global)
		}
	}
}

// func TestScriptSendMessage(t *testing.T) {
// 	for global, value := range scriptGlobals {
// 		state := NewInitialStateMode("chat")
// 		state.Connections = []gateway.Connection{
// 			gatewaySlack.New("token"),
// 		}
// 		state.ActiveConnection().SetSelectedChannel(&gateway.Channel{Id: "channel-id"})
// 		term := frontend.NewTerminalDisplay(nil)
//
// 		L := lua.NewState()
// 		defer L.Close()
//
// 		AddSlickStandardLib(L, state, term)
//
// 		args := lua.NewState()
// 		defer args.Close()
// 		for _, item := range value {
// 			switch item.(type) {
// 			case string:
// 				args.Push(lua.LString(item.(string)))
// 			case lua.LFunction:
// 				args.Push(L.NewFunction(func(L *lua.LState) int { return 0 }))
// 			}
// 		}
//
// 		fn := *(L.GetGlobal(global).(*lua.LFunction))
// 		fn.GFunction(args)
// 		err := args.Get(-1)
//
// 		if err != lua.LNil {
// 			t.Errorf("Error sending message with `sendmessage`: %s", err)
// 		}
// 	}
// }

// A utility function used below in `TestStructToTable`
func assertLuaEqual(t *testing.T, a lua.LValue, b lua.LValue) {
	_, file, line, _ := runtime.Caller(1)
	if a.String() != b.String() {
		t.Errorf("Assertion failed %s:%d: %s != %s", file, line, a, b)
	}
}

func TestStructToTable(t *testing.T) {
	type MyStruct struct {
		Foo int
		Bar string
		Baz bool
		Quux int
		Hello []int
	}

	instance := MyStruct{
		Foo: 5,
		Bar: "hello world",
		Baz: false,
		/* no Quux key */
		Hello: []int{1, 2, 3}, // A complex key that can't be converted
	}

	L := lua.NewState()
	defer L.Close()

	table := StructToTable(L, instance)
	assertLuaEqual(t, table.RawGet(lua.LString("Foo")), lua.LNumber(5))
	assertLuaEqual(t, table.RawGet(lua.LString("Bar")), lua.LString("hello world"))
	assertLuaEqual(t, table.RawGet(lua.LString("Baz")), lua.LBool(false))

	// An unset key
	assertLuaEqual(t, table.RawGet(lua.LString("Quux")), lua.LNumber(0))
	// A complex value like a slice or map
	assertLuaEqual(t, table.RawGet(lua.LString("Hello")), lua.LNil)
}

func TestStructToTableRecursiveStruct(t *testing.T) {
	type Foo struct {
		Hello string
	}
	type MyStruct struct {
		Foo Foo
		Bar *Foo
	}

	instance := MyStruct{
		Foo: Foo{Hello: "world"},
		Bar: &Foo{Hello: "Bob"},
	}

	L := lua.NewState()
	defer L.Close()

	table := StructToTable(L, instance)
	assertLuaEqual(
		t,
		table.RawGet(lua.LString("Foo")).(*lua.LTable).RawGet(lua.LString("Hello")),
		lua.LString("world"),
	)
	assertLuaEqual(
		t,
		table.RawGet(lua.LString("Bar")).(*lua.LTable).RawGet(lua.LString("Hello")),
		lua.LString("Bob"),
	)
	assertLuaEqual(
		t,
		table.RawGet(lua.LString("Baz")),
		lua.LNil,
	)
}

// Ensure that a nil pointer is properly skipped over
func TestStructToTableNilPointer(t *testing.T) {
	type Foo struct {
		Hello string
	}
	type MyStruct struct {
		Foo *Foo
	}

	instance := MyStruct{ Foo: nil }

	L := lua.NewState()
	defer L.Close()

	table := StructToTable(L, instance)
	assertLuaEqual(t, table.RawGet(lua.LString("Foo")), lua.LNil)
}
