# Contributing

Source files are relatively well documented, but here's a quick rundown:

# `slick.go` is the main file, which contains initialization logic.

# `gateway` contains all the code that relates to talking to a message gateway:
	- All the structs (`Channel`, `Message`, `User`, etc...)
	- An implementation of the `gateway.Connection` interface for slack, called `SlackConnection`
# `frontend` contains all the code to draw the app to the screen
	- Each `draw_*.go` handles a ui element.

# `version` handles keeping track of the version and auto-updates.

The version module should handle tracking the current symantic version of the program and auto
updates.

Within the module you'll find a function called `Version` that returns the major, minor, and patch
versions concatenated together as a string in the format `vMAJOR.MINOR.PATCH`.

Also, there's another function called `AutoUpdate` that handles the automatic update process for the
application. Since all releases are made on github, checking for an update consists of pulling down
the latest release from github and getting its version number. If the version number is more recent
than the current version installed, the correct binary is fetched locally. Then, the location of the
current binary is fetched from `os.Executable()` and the new binary is written over the current
binary. Finally, slick tells the user to restart.

- `color` takes care of conversion between the `red:green:B` color format and tcell's color format.
- `status` contains the core state management code for the status bar's status line.

- `state` stores the entire application's state in a data-structure.

# `gateway_events` contains code to handle incoming events from the message gateway. Updates the state.

Gateway Events handles events that are received from the message gateway (in this case, slack). Each
event is handled in a slightly different way, but each somehow changes something in the main state,
which when rendered causes a ui change.

- `message`: When a message is received, a number of important properties are contained within,
  including message text, sender, and timestamp. First, after verifying that the message received
  wasn't duplicated, the message is emitted as an event (`EVENT_MESSAGE_RECEIVED`) so that lua can
  tap into it. Then, if the message wasn't sent by the user, and applicable, then notify the user.
  Finally, add the message to the end of the active channel, if required.

- `message_deleted`: If a message is deleted, search through the message history of the current
  channel and delete any messages with a matching timestamp. We don't have to worry about any other
  channels because they'll get refetched when they are switched to.

- `user_typing`: When a user starts typing, tell the active connection to keep track of the given
  user and the current timestamp, so that after a number of seconds the user's typing activity will
  expire.

- `reaction_added` / `reaction_removed`: Both events contain a message that they apply to. When
  the status of a reaction changes, then try find the message in the current message history and
  either add the reaction (tagged with the user id) or remove the reaction from the message. In the
  case of removing a reaction, if a user was the only use reacting, delete the whole reaction
  (instead of just removing the user tag on that reaction)

## `notifications` displays any notifications that the message gateway requests.

# `keyboard_events` contains keyboard event handling logic. Updates the state.

## `commands` is sent any commands from `keyboard_events` and runs them.

# `render` kicks off a screen rerender using a modified state. 

## Commit messages
This project strives to use [semantic commit messages](https://seesparkbox.com/foundry/semantic_commit_messages). In particular, we make use of the prefixes:
- `code`
- `chore`
- `docs` - indicates a non-code update to documentation, such as a README change.
- `test` - indicates updates to test code.
- `slug` - indicates a useless commit, such as causing a ci rebuild. These should be rebased together prior
  to merging if at all possible.
- `design` - indicates style changes or other visual updates

In addition, an optional second tier of description can be added if helpful with parenthesis. This
is usually used to provide context on a code update, a chore, or a slug.

### Examples
```
code: quit goroutine when all connections are disconnected.
code(test): tests for newline in messages
docs: add ci readme badge
slug: ci update
code(design): Fill all lines along the left with a background color, even if there's no line number in a cell.
docs(chore): add image
```
