package main

import (
	"io"
	"os"
	"net/http"
	"github.com/0xAX/notificator"
)

var imagePath string = "/tmp/slime_notification.png"

var notify *notificator.Notificator = notificator.New(notificator.Options{
	DefaultIcon: "/tmp/slime_notification.png",
	AppName:     "Slime",
})

// Send a notification when a message is received.
func Notification(title string, text string, imageUrl string) {
	resp, err := http.Get(imageUrl)

	// Create an io.Reader to download the notification icon.
	if err != nil {
		notify.Push(title, text, "", notificator.UR_CRITICAL)
		return
	}
	defer resp.Body.Close()

	// Create an io.Writer to write the file to disk
	var file *os.File
	file, err = os.Create(imagePath)

	// io.Reader => io.Writer
	io.Copy(file, resp.Body)
	notify.Push(title, text, imagePath, notificator.UR_CRITICAL)
}
