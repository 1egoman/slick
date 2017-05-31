# Disconnect

Type: Native (built into slick)

Arguments:
- `[connection index]`

Command aliases:
- `disconnect`
- `dis`

## Description

Disconnect the specified connection index. If no connection index is specified, then use the
currently active connection.

## Example

`/disconnect`

`/disconnect 3`

```lua
keymap("pp", function()
	err = Disconnect("1") -- or, `Disconnect()`
	if err then
		error(err)
	end
end)
```
