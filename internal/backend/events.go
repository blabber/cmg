// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Event struct {
	Title       string           `json:"title"`
	Subtitle    string           `json:"subtitle"`
	Description string           `json:"description"`
	Language    string           `json:"original_language"`
	Persons     []string         `json:"persons"`
	Tags        []string         `json:"tags"`
	Date        string           `json:"date"`
	Length      int              `json:"length"`
	Recordings  []EventRecording `json:"recordings"`
}

type EventRecording struct {
	MimeType     string `json:"mime_type"`
	Filename     string `json:"filename"`
	RecordingUrl string `json:"recording_url"`
}

func GetEvent(id string) (*Event, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s/%s", backend, "public/events", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := new(Event)
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
