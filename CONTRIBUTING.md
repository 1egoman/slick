# General Design Pattern

```
            Frontend (`frontend` package)
               ||
               ||
           Application Logic (`main` package)
              //   \\
             //     \\
         Gateway   Gateway
```


- The `frontend` package is the interface to the terminal. It contains all the drawing code and the is
  (currently) tightly coupled to tcell, the terminal drawing library this project uses.
- The `main` package contains the application logic. This switching modes, typing commands, and
  handling keyboard input are all handled by this code.
- Gateways are connections to slack. One websocket conenction is made per team, and all "pushed"
  events flow over that connection. In addition, gateways also abstract away deatils such as how to
  get channels, teams details, and more.

By abstracting each of these three "layers" into their own packages, the hope is to minimize the
amount of coupling of layers to each other.

## Frontend
The frontend is available as `frontend.TerminalDisplay`. This struct handles drawing ui components
such at the status bar, the command bar, messages, and the fuzzy picker.

## Gateway
Gateways have a main interface, `gateway.Connection`. Anything that wants to be a gateway needs to
implement this interface. There's a single gateway at the moment, `gatewaySlack.SlackConnection`. It has
a constructor called `gatewaySlack.New` that takes a single parameter - the auth token for slack.

# Project structure

```
.
├── Biomefile                // Holds environment variables needed to run the app
├── CONTRIBUTING.md          // (this doc!)
├── README.md
├── SHORTCUTS.md
├── commands.go              // Handles slash and colon command execution, processing, and definition.
├── commands_test.go
├── connect.go               // When the app starts, this kicks off each connection (calls `connection.Connect`)
├── fuzzy_picker.go          // Fuzzy picker struct and logic. See file for a usage example.
│
├── frontend
│   ├── draw_fuzzy_picker.go // Renders the fuzzy picker, the thing used to pick channels and provide autocomplete
│   ├── draw_messages.go     // Renders a list of messages
│   └── frontend.go          // Miscellaneous drawing and initialization logic.
│
├── gateway
│   ├── gateway.go           // Defines the gateway interface that slack implements.
│   └── slack
│       ├── connect.go       // Connect to slack message server via websocket. Called by ../connect.go.
│       ├── post_text.go     // Make a post / snipper to a given channel.
│       ├── refresh.go       // When a slack connection becomes active again, call this to "refresh" channel list / active user / etc
│       ├── send_message.go  // Send a message to a channel
│       └── slack.go         // Miscellaneous slack constants and initialization logic.
│
├── status
│   └── status.go            // Struct for managing the storage and usage of the status message.
│
├── gateway_events.go        // Handle incoming events from gateways.
├── keyboard_events.go       // Handle incoming events from the keyboard.
├── keyboard_events_test.go
├── main.go                  // Program starts here! :)
├── render.go                // Contains main render loop.
└── state.go                 // Contains app state struct definitions.
```



# Contributing

Source files are relatively well documented, but here's a quick rundown:
- `gateway` contains all the code that relates to talking to a message gateway:
	- All the structs (`Channel`, `Message`, `User`, etc...)
	- An implementation of the `gateway.Connection` interface for slack, called `SlackConnection`
- ``
