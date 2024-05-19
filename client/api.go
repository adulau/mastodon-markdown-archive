package client

import (
	"encoding/json"
	"io"
	"net/http"
)

func Fetch(requestUrl string, variable interface{}, headers map[string]string) error {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", requestUrl, nil)

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	res, err := client.Do(req)

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
