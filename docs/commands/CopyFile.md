# CopyFile

Type: Native (built into slick)

Command aliases:
- `copyfile`
- `cpf`

## Description
If a message has a file attached, then copy a link to the file to the clipboard.
Aliased to the `c` key when a message with a file attached is selected.

NOTE: no index is required because messages can only have 1 file attached - this is a limitation of
slack.

## Example

`/copyfile`

```lua
keymap("pp", function()
	err = CopyFile()
	if err then
		error(err)
	end
end)
```
