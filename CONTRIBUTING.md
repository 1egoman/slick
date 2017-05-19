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

# Config files
Slime configuration files are written in lua. Configuration files are searched from the current
directory up to the root, and each must have the name or `.slimerc`. For example, if the current
folder or any folder below the current folder contains a file named `.slimerc`, it will be loaded
automatically.

Config files contain a few unique functions:

- `print("hello world")` logs a given string to the bottom status bar.
- `error("oh noes")` logs an error to the bottom status bar.
- `clear()` clear's the bottom status bar.
- `getenv("ENVIRONMENT_VARIABLE")` fetches the contents of the specified environment variable.
- `keymap("key sequence", <callback>)` calls the callback when a given key sequence is pressed. For
  example:

```lua
keymap("ff", function()
	print("User pressed ff!")
end)
```
- `output, err = shell("ls", "-l")` calls the given shell command. Returns stdout and err.
- `output, err = command("ls", "-l")` calls the given shell command. Returns stdout and err.
- `getenv("HOME")` returns the contents of an environment variable
- `sendmessage("Hello World!")` sends a message to the currently active connection and channel.
- `command("name", "desc", "args", <callback>)` constructs a command. For example:

```lua
command("hello", "A nice greeting!", "[name]", function(args)
	if strings.len(args[2]) > 0 then
		sendmessage("Hello @"..args[2]) -- Greet a user
	else
		sendmessage("Hello world!") -- Greet the world
	end
end)
```

Now, if the user runs `/hello`, the message `Hello world!` will be sent. If the user runs `/hello
bob`, `Hello @bob` will be sent.

## Commands in config files
Commands map one-to-one with functions in a config file. For example, the command `/postinline`
defines a function called `PostInline` that can be used to run that command within the config file.
As arguments to these command functions, pass the same arguments you'd pass to the command. Ie,
`/postinline foo bar` == `PostInline("foo", "bar")`.

## Example config file
```lua
-- Connect to slack teams
Connect(getenv("SLACK_TOKEN_ONE"))
Connect(getenv("SLACK_TOKEN_TWO"))

-- Example key binding
keymap("ff", function()
	err = PostInline("content", "title")
	if err then
		error(err)
	else
		print("Successfully posted!")
	end
end)
```
