package main

import (
	"github.com/1egoman/slime/gateway"
)

// This struct contains the main application state. I have fluxy intentions.
type State struct {
	Mode string

	Command               []rune
	CommandCursorPosition int

	Gateway gateway.Connection

	MessageHistory []gateway.Message
}

