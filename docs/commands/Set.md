# Set

Type: Native (built into slick)

Arguments:
- `<option name>`
- `[option value]`

Command aliases:
- `set`

## Description

Set the configuration option specified by `<option name>` with the value `[option value]`. If the
optino value is omitted, then unset the configuration option.

## Example

Set the configuration key `MyOption.NameHere` to `option value`:
`/set MyOption.NameHere "option value"`

Unset the configuration key `MyOption.NameHere`:
`/set MyOption.NameHere`

```lua
-- Set the config option `MyOption.NameHere` to `option value`:
Set("MyOption.NameHere", "option value")

-- Unset the config option `MyOption.NameHere`:
Set("MyOption.NameHere")
```
