# OpenAttachmentLink

Type: Native (built into slick)

Arguments:
- `[attachment index]`

Command aliases:
- `attachmentlink`
- `atlink`
- `atlk`

## Description
If a message has an attachment, then open the title link in that attachment in the default web
browser. Specify an index to target an attachment other than the first attachment. Aliased to the
`l` key when a message with at least one attachment is selected.

## Example

`/openattachmentlink` (open first attachment' link by default)

`/openattachmentlink 2`

```lua
keymap("pp", function()
	err = OpenAttachmentLink()
	if err then
		error(err)
	end
end)
```
