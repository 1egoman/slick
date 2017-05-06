package gatewaySlack

import (
	"log"
	"errors"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (c *SlackConnection) PostText(title string, content string) error {
	log.Printf("* Posting text to active channel: '%s' '%s'", title, content)
	title = url.QueryEscape(title)
	content = url.QueryEscape(content)

	// Assemble the query string
	queryString := "?token=" + c.token
	queryString += "&channels=" + c.selectedChannel.Id
	queryString += "&content=" + content

	if len(title) > 0 {
		queryString += "&title=" + title
	}

	resp, err := http.Get("https://slack.com/api/files.upload" + queryString)
	if err != nil {
		return err
	}

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)

	var fileBuffer struct {
		Ok bool `json:"ok"`
		Error string `json:"error"`
		File struct {
			Id string `json:"id"`
			Mimetype string `json:"mimetype"`
			Mode string `json:"snippet"`
			Public string `json:"is_public"`
			Preview string `json:"preview"`
		} `json:"file"`
	}

	json.Unmarshal(body, &fileBuffer)

	if len(fileBuffer.Error) != 0 {
		log.Println("Error posting to channel", fileBuffer.Error)
		return errors.New(fileBuffer.Error)
	}
	return nil
}
