# Contributing

Source files are relatively well documented, but here's a quick rundown:

- `slick.go` is the main file, which contains initialization logic.

- `gateway` contains all the code that relates to talking to a message gateway:
	- All the structs (`Channel`, `Message`, `User`, etc...)
	- An implementation of the `gateway.Connection` interface for slack, called `SlackConnection`
- `frontend` contains all the code to draw the app to the screen
	- Each `draw_*.go` handles a ui element.
- `version` handles keeping track of the version and auto-updates.
- `color` takes care of conversion between the `red:green:B` color format and tcell's color format.
- `status` contains the core state management code for the status bar's status line.

- `state` stores the entire application's state in a data-structure.
- `gateway_events` contains code to handle incoming events from the message gateway. Updates the state.
	- `notifications` displays any notifications that the message gateway requests.
- `keyboard_events` contains keyboard event handling logic. Updates the state.
	- `commands` is sent any commands from `keyboard_events` and runs them.
- `render` kicks off a screen rerender using a modified state. 

## Commit messages
This project strives to use [semantic commit messages](https://seesparkbox.com/foundry/semantic_commit_messages). In particular, we make use of the prefixes:
- `code`
- `chore`
- `docs` - indicates a non-code update to documentation, such as a README change.
- `test` - indicates updates to test code.
- `slug` - indicates a useless commit, such as causing a ci rebuild. These should be rebased together prior
  to merging if at all possible.
- `design` - indicates stylesheet changes or other visual updates

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
