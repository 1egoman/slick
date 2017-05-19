package main

import (
	"github.com/0xAX/notificator"
)

var notify *notificator.Notificator = notificator.New(notificator.Options{
	DefaultIcon: "icon/default.png",
	AppName:     "Slime",
})

// Send a notification when a message is received.
func Notification(title string, text string) {
  notify.Push(title, text, "/home/user/icon.png", notificator.UR_CRITICAL)
}
