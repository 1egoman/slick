# Post

Type: Native (built into slick)

Arguments:
- `<post file>` - Path to a file to use as the post content.
- `[post name]` - Optional post name.

Command aliases:
- `post`
- `p`

## Description

Create a post in the current slack channel, using the passed file as the post content. If a post
name is passed, use it to title the post accordingly.

## Example

`/post /path/to/file.txt "My Cool Post"`

```lua
keymap("pp", function()
	err = Post("/path/to/file.txt", "My Cool Post")
	if err then
		error(err)
	end
end)
```
