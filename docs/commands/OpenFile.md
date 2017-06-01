# OpenFile

Type: Native (built into slick)

Command aliases:
- `openfile`
- `opf`

## Description
If a message has a file attached, then open the file in an external application to view.
Aliased to the `o` key when a message with a file attached is selected.

NOTE: no index is required because messages can only have 1 file attached - this is a limitation of
slack.

## Example

`/openfile`

```lua
keymap("pp", function()
	err = OpenFile()
	if err then
		error(err)
	end
end)
```
