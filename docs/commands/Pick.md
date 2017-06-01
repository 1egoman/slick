# Pick

Type: Native (built into slick)

Arguments:
- `<connection name>`
- `<channel name>`

Command aliases:
- `pick`
- `p`

## Description
Switch the active connection and active channel to the passed connection and channel.

This is very similar to the connection picker that is activated by pressin `p`, but in command form.

## Example

`/pick "my connection" "general"`

```lua
keymap("pp", function()
	err = Pick("my connection", "general")
	if err then
		error(err)
	end
end)
```
