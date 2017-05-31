# Connect

Type: Native (built into slick)

Arguments:
- `[team name]` - A name to associate with the connection. If unspecified, fetch the team name from
  slack.
- `<token>` - The token for the slack team. [Here's how to get this](../../Connecting.md).

Command aliases:
- `connect`
- `con`

## Description

Connect will attach slick to a given slack team, creating a connection in the process.

## Example

`/connect "team name" "xoxp-SLACK-TOKEN-HERE"`

```lua
Connect("team name", "xoxp-SLACK-TOKEN-HERE")
```
