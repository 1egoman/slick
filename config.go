package main

import (
	"log"
	"os"
	"path"
	"strings"
	"io/ioutil"
)

const CONFIG_FILE_NAME = ".slimerc"

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
