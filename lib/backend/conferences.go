// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package backend

import "fmt"

// A conferences represents a collection of Conferences.
type conferences struct {
	Conferences []Conference
}

// A Conference represents a conference as returned by the backend.
type Conference struct {
	Title  string            `json:"title"`
	Slug   string            `json:"slug"`
	Events []ConferenceEvent `json:"events"`

	// Url represents the backend URL of the conference. The last part of
	// it is the id that is needed for calls to GetConference.
	Url string `json:"url"`
}

// A Conference represents an event (talk, workshop, ...)  that took place
// during a Conference.
type ConferenceEvent struct {
	Title string `json:"title"`

	// Url represents the backend URL of the event. The last part of
	// it is the id that is needed for calls to GetEvent.
	Url string `json:"url"`
}

// GetConferences returns a []Conference containing all conferences known to the
// backend. The returned Conferences contain no ConferenceEvents. In order to
// get the events a call to GetConference is required.
func GetConferences() ([]Conference, error) {
	var data conferences

	err := getData("public/conferences", &data)
	if err != nil {
		return data.Conferences, err
	}

	return data.Conferences, nil
}

// GetConference returns a single Conference, containing the events that took
// place during the conference.
func GetConference(id string) (Conference, error) {
	var data Conference

	err := getData(fmt.Sprintf("%s/%s", "public/conferences", id), &data)
	if err != nil {
		return data, err
	}

	return data, nil
}
