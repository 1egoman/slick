package main

import (
	"log"
	"os"
	"path"
	"errors"
	"strings"
	"bytes"
	"io/ioutil"
	"encoding/gob"
	"github.com/1egoman/slime/gateway"
)

const CONFIG_FILE_NAME = ".slimerc"

type SerializedConnection struct {
	MessageHistory []gateway.Message
	Channels []gateway.Channel
	SelectedChannel gateway.Channel
}

func GetConfigFileContents() map[string]string {
	configFiles := make(map[string]string)

	pathElements := strings.Split(os.Getenv("PWD"), "/")
	for index := 0; index <= len(pathElements); index++ {
		filename := "/" + path.Join(path.Join(pathElements[:index]...), CONFIG_FILE_NAME)
		log.Println("Searching config path", filename)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			continue
		}

		configFiles[filename] = string(data)
	}

	return configFiles
}

func PathToSavedConnections() string {
	return os.Getenv("HOME") + "/.slimecache/"
}

func SaveConnection(conn gateway.Connection) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(SerializedConnection{
		MessageHistory: conn.MessageHistory(),
		Channels: conn.Channels(),
		SelectedChannel: *conn.SelectedChannel(),
	})

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(PathToSavedConnections() + conn.Name(), buf.Bytes(), 0777)
	if err != nil {
		return err
	}

	return nil
}

func ApplySaveToConnection(name string, conn *gateway.Connection) error {
	byt, err := ioutil.ReadFile(PathToSavedConnections() + name)
	if err != nil {
		return err
	}

	var serialized SerializedConnection

	buf := bytes.NewBuffer(byt)
	dec := gob.NewDecoder(buf)

	err = dec.Decode(&serialized)

	if err != nil {
		return err
	}

	if conn == nil {
		return errors.New("Passed connection was nil!")
	}

	(*conn).SetSelectedChannel(&serialized.SelectedChannel)
	(*conn).SetChannels(serialized.Channels)
	(*conn).SetMessageHistory(serialized.MessageHistory)

	return nil
}
