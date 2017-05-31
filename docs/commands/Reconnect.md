# Reconnect

Type: Native (built into slick)

Arguments:
- `[connection index]`

Command aliases:
- `reconnect`
- `recon`

## Description

Reconnect the specified connection index. If no connection index is specified, then use the
currently active connection. A reconnect disconnects, connects, then force-refreshes the connection.
A force refresh re-fetches a list of channels and details about the currently logged in user.

This command is a helpful in troubleshooting - essentially, you're able to proove that you have the
latest message history from the server.

## Example

`/reconnect`

```lua
keymap("pp", function()
	err = Reconnect()
	if err then
		error(err)
	end
end)
```
