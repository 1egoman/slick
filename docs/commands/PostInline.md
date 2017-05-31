# PostInline

Type: Native (built into slick)

Arguments:
- `<post content>` - The text to add inside of the post.
- `[post name]` - An optional title for the file. If upspecified, uses slack's default of `-.txt`.

Command aliases:
- `postinline`
- `pin`

## Description

Given the contents of a post an optional title, create a post in the current slack channel.

## Example

`/postinline "post content" "post title"`

```lua
keymap("pp", function()
	err = PostInline("post content", "post title")
	if err then
		error(err)
	end
end)
```
