# MoveForwardMessage

Type: Native (built into slick)

Command aliases:
- `moveforwardmessage`

## Description
Moves selection up a message in the message history. This is exactly what `j` or the up arrow does
then you press them.

## Example

`/moveforwardmessage`

```lua
keymap("pp", function()
	err = MoveForwardMessage()
	if err then
		error(err)
	end
end)
```
