# OpenInSlack

Type: Native (built into slick)

Command aliases:
- `openinslack`
- `ops`

## Description

Sometimes, it's helpful to be able to switch to the first party slack app. This command opens the
current channel in officil slack app, deep linking to the correct connection and channel.

## Example

`/openinslack`

```lua
keymap("pp", function()
	err = OpenInSlack()
	if err then
		error(err)
	end
end)
```
