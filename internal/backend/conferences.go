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

type conferences struct {
	Conferences []Conference
}

type Conference struct {
	Title  string            `json:"title"`
	Slug   string            `json:"slug"`
	Url    string            `json:"url"`
	Events []ConferenceEvent `json:"events"`
}

type ConferenceEvent struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func GetConferences() ([]Conference, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", backend, "public/conferences"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data conferences
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data.Conferences, nil
}

func GetConferenceById(id string) (*Conference, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s/%s", backend, "public/conferences", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := new(Conference)
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
