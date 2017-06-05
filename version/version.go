package version

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
)

const VERSION_MAJOR = 0
const VERSION_MINOR = 2
const VERSION_PATCH = 0

func Version() string {
	return fmt.Sprintf("v%d.%d.%d", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
}

//
// AUTO-UPDATE
//

type Release struct {
	Name    string `json:"name"`
	Author  string
	Version string `json:"tag_name"`
	Binary  string
}

func ReleaseBinary(data map[string]interface{}) string {
	binaryName := fmt.Sprintf("slick-%s-%s", runtime.GOOS, runtime.GOARCH)

	if assets, ok := data["assets"].([]interface{}); ok {
		for _, a := range assets {
			asset := a.(map[string]interface{})
			if asset["name"] == binaryName {
				return asset["browser_download_url"].(string)
			}
		}
	}

	return ""
}

func ReleaseDetails() (*Release, error) {
	resp, err := http.Get("https://api.github.com/repos/1egoman/slick/releases/latest")
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	json.Unmarshal(body, &data)
	return &Release{
		Name:    data["name"].(string),
		Author:  data["author"].(map[string]interface{})["login"].(string),
		Version: data["tag_name"].(string),
		Binary:  ReleaseBinary(data),
	}, nil
}

func DoUpdate() *string {
	release, err := ReleaseDetails()

	if err != nil {
		log.Printf("Error getting release details: %s", err)
		return nil
	}

	log.Println("Fetched slick release information from github")

	if release.Version != Version() {
		log.Printf("Downloading release %s...", release.Version)
		resp, err := http.Get(release.Binary)
		if err != nil {
			log.Printf("Error downloading binary: %s", err.Error())
			return nil
		}

		ex, err := os.Executable()
		if err != nil {
			log.Printf("Error getting executable: %s", err.Error())
			return nil
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error decoding body: %s", err.Error())
			return nil
		}

		err = ioutil.WriteFile(ex, body, 0644)
		if err != nil {
			log.Printf("Error creating temporary file: %s", err.Error())
			return nil
		}

		log.Printf("Downloaded slick %s into %s", release.Version, ex)
		return &release.Version
	} else {
		log.Printf("Slick already on latest version: %s", release.Version)
		return nil
	}
}
