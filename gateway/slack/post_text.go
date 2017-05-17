package gatewaySlack

import (
	"log"
	"errors"
	"bytes"
	"mime/multipart"
	"io"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (c *SlackConnection) PostText(title string, content string) error {
	log.Printf("* Posting text to active channel: '%s'", title)
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

func (c *SlackConnection) PostBinary(title string, filename string, content []byte) error {
	log.Printf("* Posting binary to active channel: '%s'", filename)
	title = url.QueryEscape(title)

	// Assemble the query string
	url := "https://slack.com/api/files.upload"
	url += "?token=" + c.token
	url += "&file=file"
	url += "&channels=" + c.selectedChannel.Id
	if len(title) > 0 {
		url += "&title=" + title
	}

    // Prepare a form that you will submit to that URL.
    var b bytes.Buffer
    w := multipart.NewWriter(&b)
    // Add your image file
    fw, err := w.CreateFormFile("file", filename)
    if err != nil {
        return err
    }
    if _, err = io.Copy(fw, bytes.NewBuffer(content)); err != nil {
        return err
    }
    // Don't forget to close the multipart writer.
    // If you don't close it, your request will be missing the terminating boundary.
    w.Close()

    req, err := http.NewRequest("POST", url, &b)
    // Don't forget to set the content type, this will contain the boundary.
    req.Header.Set("Content-Type", w.FormDataContentType())

    client := &http.Client{}
    _, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
