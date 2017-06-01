# MoveBackMessage

Type: Native (built into slick)

Command aliases:
- `movebackmessage`

## Description
Moves selection down a message in the message history. This is exactly what `k` or the down arrow does
then you press them.

## Example

`/movebackmessage`

```lua
keymap("pp", function()
	err = MoveBackMessage()
	if err then
		error(err)
	end
end)
```
