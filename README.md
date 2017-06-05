# Slick

Slick is a modular and extendable [slack](//slack.com) client with a terminal based ui, while also
aiming to be simple and intuitive. It's been designed to have an approachable default configuration but to be easily
extended to add new functionality in [lua](//lua.org)

# Installing
[Here's the skinny](docs/Installing.md).

## Features

- **Modal** - Slick borrows a text based, [modal](docs/Modal.md) workflow from vi. Most
  functionality requires one keypress, or [can be easily mapped to a key](docs/Scripting.md).
- **Scriptable** - Add new commands (ie, `/foo`) or keyboard bindings (ie, press `a`) and bind
  them to slick commands. Or, write your own functionality in [Lua](//lua.org) - for example,
  [here's a plugin](examples/encrypt.lua) to encrypt a message to a user on keybase and send it to
  them via slack. [Learn More](docs/Scripting.md)
- **Batteries Included** - Distributed as a static binary with no dependencies.
  [Installation](docs/Installing.md) is simple. Slick is [updated automatically](docs/AutoUpdate.md)
  on start.
- **Not built on electron** - Slick is terminal based. Reduce the number of bloated [chrome
  vms](https://josephg.com/blog/electron-is-flash-for-the-desktop/) running on your system.

And a bunch of smaller things:
- Quick jump to another team / channel with `p`
- Multiple teams
- Tab completion for file paths
- A lua [standard library](https://github.com/1egoman/slick/blob/master/docs/Scripting.md#modules)
- Emoji support
- Extensive theming support - ie, [here](https://github.com/1egoman/slick/blob/master/docs/configuration/Message.Part.ChannelColor.md) [are](https://github.com/1egoman/slick/blob/master/docs/configuration/Message.Attachment.FieldValueColor.md) [a few](https://github.com/1egoman/slick/blob/master/docs/configuration/StatusBar.LogColor.md) (examples)[https://github.com/1egoman/slick/blob/master/docs/configuration/StatusBar.GatewayConnectingColor.md]
