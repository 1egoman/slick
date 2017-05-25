# StatusBar.GatewayFailedColor

- Type: `color`
- Default: `::B` [(format explanation)](../colors.md)

This option specifies how connections that have fallen into a failed state will be rendered. This
can happen if the underlying socket disconnects from slack's servers, or there is an error in the
socket code that is irrecoverable.

## Usage
`:set StatusBar.GatewayFailedColor red:green:`
![gifs/StatusBar.GatewayFailedColor.png](gifs/StatusBar.GatewayFailedColor.png)
