# Commands

When in [`write` mode](Modes.md), the user can send messages as well as send commands. To start
executing a command, type `/`:

![Running Commands](gifs/RunningCommands.png)

As you continue to type, the command list will filter to the command you're typing and give a hint
about its syntax. Once done typing the command, press `enter`:

![Running Commands](gifs/RunningCommands.gif)

(Learn about [PostInline](commands/PostInline.md))

## Quoting
Each word in a command by default becomes an argument to the command. For example, in `/foo bar`
baz`, the first argument is `bar` and the second is `baz`. However, what if you want to pass the
string `bar baz` as the combind first argument to `/foo`? If you surround the argument in quotes, like
`/foo "bar baz"`, then `bar baz` will be considered a single command.

## Styling
- [CommandBar.PrefixColor](configuration/CommandBar.PrefixColor.md)
- [CommandBar.TextColor](configuration/CommandBar.TextColor.md)
