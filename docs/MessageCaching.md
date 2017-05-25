# Message Caching
To help speed up loading of new slack channels, slime caches a number of attributes about each
connection:

```golang
type SerializedConnection struct {
	MessageHistory []gateway.Message // A list of messages for the current channel
	Channels []gateway.Channel // A list of channels
	SelectedChannel gateway.Channel // The currently selected channel and its attributes.
}
```

This struct is packed with [gob](https://golang.org/pkg/encoding/gob/) and saved into
`~/.slimecache/<connection name>` before slime exits. Then, when slime starts up, this file is read
into memory, unpacked, and used to quickly show conenction details right away to make the experience
feel much more snappy.

## The cache seems to be keeping me from pulling down the latest updates from slack!
- First, try running `/reconnect`. This will cause slime to reconnect and pull down the latest
  channel list and user details from slack's servers.
- If that doesn't seem to be working, close all instances of slime, clear your cache: `rm -rf ~/.slimecache`, and
  start slime again. Note that slime writes to the cache on exit, so make sure that all copies of
  slime are closed before clearing the cache.

## This cache thing is more trouble than it's worth, I don't want it.
- :frowning: - [Leave an issue](https://github.com/1egoman/slime/issues/new)?
- To disabling the cache, clear the `Connection.Cache` configuration option: add `Set("Connection.Cache")` to your `.slimerc`.
