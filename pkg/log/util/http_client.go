package util

import (
	"bytes"
	"io"
	"net/http"
)

func ZsHttpClient(method string, url string, paramJsonString string, username string, password string) (string, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(paramJsonString)))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(username, password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	//fmt.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
