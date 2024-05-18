package client

import (
	"encoding/json"
	"io"
	"net/http"
)

func Fetch(requestUrl string, variable interface{}) error {
	res, err := http.Get(requestUrl)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err := json.Unmarshal(body, variable); err != nil {
		return err
	}

	return nil
}
