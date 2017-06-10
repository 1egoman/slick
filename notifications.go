package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/0xAX/notificator"
	"github.com/1egoman/slick/gateway"
)

var imagePath string = "/tmp/slick_notification.png"

var notify *notificator.Notificator = notificator.New(notificator.Options{
	DefaultIcon: "/tmp/slick_notification.png",
	AppName:     "Slick",
})

// Send a notification when a message is received.
func Notification(title string, text string) {
	log.Println("SEND NOTIFICATION", title, text)
	notify.Push(title, text, "", notificator.UR_CRITICAL)
}

// When should we notify a user about a message?
// 1. When the channel that the message is received in is a channel we're a member of.
// 2. When the regex matches.
func ShouldMessageNotifyUser(messageBody string, selectedChannel *gateway.Channel, self *gateway.User) bool {
	if self != nil && selectedChannel != nil && selectedChannel.IsMember {
		notificationRegex := fmt.Sprintf("(%s|<@%s|<@%s|<!channel|<!here|<!everyone)", self.Name, self.Id, self.RealName)
		match, _ := regexp.MatchString(notificationRegex, messageBody)
		return match
	} else {
		return false
	}
}
