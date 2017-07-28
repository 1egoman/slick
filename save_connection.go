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
	Self            gateway.User
	Team            gateway.Team
}

type SerializedGlobalState struct {
	ActiveConnectionIndex int
	SelectedMessageIndex int
	BottomDisplayedItem int
}

func PathToSavedConnections() string {
	return PathToCache() + "connections/"
}
func PathToCache() string {
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
		Self:            *conn.Self(),
		Team:            *conn.Team(),
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
	(*conn).SetSelf(serialized.Self)
	(*conn).SetTeam(serialized.Team)

	return nil
}

func SaveGlobalState(state *State) error {
	// Get the index of the active connection
	var activeConnectionIndex int = 0
	activeConnection := state.ActiveConnection()
	for ct, i := range state.Connections {
		if i == activeConnection {
			activeConnectionIndex = ct
			break
		}
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(SerializedGlobalState{
		ActiveConnectionIndex: activeConnectionIndex,
		SelectedMessageIndex: state.SelectedMessageIndex,
		BottomDisplayedItem: state.BottomDisplayedItem,
	})

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(PathToCache()+"globalstate", buf.Bytes(), 0777)
	if err != nil {
		return err
	}

	return nil
}

func ApplyGlobalStateToState(state *State) error {
	byt, err := ioutil.ReadFile(PathToCache()+"globalstate")
	if err != nil {
		return err
	}

	var serialized SerializedGlobalState

	buf := bytes.NewBuffer(byt)
	dec := gob.NewDecoder(buf)

	err = dec.Decode(&serialized)

	if err != nil {
		return err
	}

	if state == nil {
		return errors.New("Passed state object was nil!")
	}

	// If the user added or removed connections and this connection index wouldn't work, then don't
	// use it.
	if serialized.ActiveConnectionIndex < len(state.Connections) {
		state.SetActiveConnection(serialized.ActiveConnectionIndex)
	}
	state.SelectedMessageIndex = serialized.SelectedMessageIndex
	state.BottomDisplayedItem = serialized.BottomDisplayedItem
	return nil
}
