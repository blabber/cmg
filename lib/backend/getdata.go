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

// getData issues a HTTP GET to path on the backend and decodes the answer
// returned by the backend into data.
func getData(path string, data interface{}) error {
	resp, err := http.Get(fmt.Sprintf("%s/%s", backend, path))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}

	return nil
}
