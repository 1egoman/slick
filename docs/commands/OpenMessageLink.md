# OpenMessageLink

Type: Native (built into slick)

Arguments:
- `[link index]`

Command aliases:
- `openmessagelink`
- `openmsglk`
- `olk`

## Description
Messages can have links within them:

![gifs/OpenMessageLink.png](gifs/OpenMessageLink.png)
(All the underlined bits are links)

This command opens a link's target in a web browser. Ie, in the example above, `/openmessagelink 1`
opens the first link in the selected message (`http://example.com`) in a web browser.

If the link index isn't specified, it defaults to `1`.

## Example

`/openmessagelink 1`

```lua
keymap("pp", function()
	err = OpenMessageLink()
	if err then
		error(err)
	end
end)
```
