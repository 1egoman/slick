# General Design Pattern

```
            Frontend (`frontend` package)
               ||
               ||
           Application Logic (`main` package)
              //   \\
             //     \\
         Gateway   Gateway
```


- The `frontend` package is the interface to the terminal. It contains all the drawing code and the is
  (currently) tightly coupled to tcell, the terminal drawing library this project uses.
- The `main` package contains the application logic. This switching modes, typing commands, and
  handling keyboard input are all handled by this code.
- Gateways are connections to slack. One websocket conenction is made per team, and all "pushed"
  events flow over that connection. In addition, gateways also abstract away deatils such as how to
  get channels, teams details, and more.

By abstracting each of these three "layers" into their own packages, the hope is to minimize the
amount of coupling of layers to each other.

## Frontend
The frontend is available as `frontend.TerminalDisplay`. This struct handles drawing ui components
such at the status bar, the command bar, and the fuzzy picker.

## Gateway
Gateways have a main interface, `gateway.Connection`. Anything that wants to be a gateway needs to
implement this interface. There's a single gateway at the moment, `gateway.SlackConnection`. It has
a constructor called `gateway.Slack` that takes a single parameter - the auth token for slack.
