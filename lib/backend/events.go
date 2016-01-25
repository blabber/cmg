// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package backend

import "fmt"

// An Event represents an event (talk, workshop, ...) as returned by the
// backend.
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

// An EventRecording represents a recording of an Event.
type EventRecording struct {
	MimeType string `json:"mime_type"`
	Filename string `json:"filename"`

	// RecordingUrl is the URL of the recording in the media.ccc.de CDN.
	RecordingUrl string `json:"recording_url"`
}

// GetEvent returns an Event representing the event (talk, workshop, ...)
// identified by id.
func GetEvent(id string) (Event, error) {
	var data Event

	err := getData(fmt.Sprintf("%s/%s", "public/events", id), &data)
	if err != nil {
		return data, err
	}

	return data, nil
}
