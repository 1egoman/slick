# Upload

Type: Native (built into slick)

Arguments:
- `<file path>` - Path to the file to upload.
- `[file name]` - Optional file name.

Command aliases:
- `upload`
- `up`

## Description
Upload the given file into the active slack channel and connection. If an optional filename is
specified, name the file accordingly.

## Example

`/upload "/path/to/file.txt"`

`/upload "/path/to/file.txt" "foo.txt"`

```lua
keymap("pp", function()
	err = Upload("/path/to/file.txt")
	if err then
		error(err)
	end
end)
```
