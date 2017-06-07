package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/1egoman/slick/gateway"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const CONFIG_FILE_NAME = ".slickrc"

func GetConfigFileContents() map[string]string {
	configFiles := make(map[string]string)

	homeFilename := path.Join(os.Getenv("HOME"), CONFIG_FILE_NAME)
	crawledHome := false

	pathElements := strings.Split(os.Getenv("PWD"), "/")
	for index := 0; index <= len(pathElements); index++ {
		filename := "/" + path.Join(path.Join(pathElements[:index]...), CONFIG_FILE_NAME)
		log.Println("Searching config path", filename)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			continue
		}
		log.Println("Config exists!", filename)

		// When traversing through the tree, did we come across the `~/.slickrc`?
		if homeFilename == filename {
			crawledHome = true
		}

		configFiles[filename] = string(data)
	}

	// Finally, look for ~/.slickrc last, if applicable.
	if !crawledHome {
		data, err := ioutil.ReadFile(homeFilename)
		if err == nil {
			configFiles[homeFilename] = string(data)
		}
	}

	return configFiles
}

//
// STORAGE OF SAVED CONNECTIONS
//

type SerializedConnection struct {
	MessageHistory  []gateway.Message
	Channels        []gateway.Channel
	SelectedChannel gateway.Channel
}

func PathToSavedConnections() string {
	return os.Getenv("HOME") + "/.slickcache/"
}

func SaveConnection(conn gateway.Connection) error {
	if conn.SelectedChannel() == nil {
		return errors.New("Can't save connection with no selected channel!")
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(SerializedConnection{
		MessageHistory:  conn.MessageHistory(),
		Channels:        conn.Channels(),
		SelectedChannel: *conn.SelectedChannel(),
	})

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(PathToSavedConnections()+conn.Name(), buf.Bytes(), 0777)
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
