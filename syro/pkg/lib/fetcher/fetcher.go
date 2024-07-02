package fetcher

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// JsonHeader is a map of headers for json requests
var JsonHeader = map[string]string{"Content-Type": "application/json"}

// Baseline for the fetch request. the 4th parameter is optional.
func Fetch(method, url string, headers map[string]string, requestBody ...[]byte) (*Response, error) {
	var reqBody []byte
	if len(requestBody) == 1 {
		reqBody = requestBody[0]
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Set request headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching %v : %v", url, err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body when fetching %v, status %v: %v, error: ", url, res.Status, err)
	}

	if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 202 {
		return nil, fmt.Errorf("response did not return status in 200 group while requesting %v, status: %v, error: %v", url, res.Status, err)
	}

	if body == nil {
		return nil, fmt.Errorf("response returned empty body while requesting %v", url)
	}

	if res.Header == nil {

	}

	return &Response{body, res.Header, res.StatusCode}, err
}

// FetchZipFiles fetches an url which returns a zip file and returns the files.
func FetchZipFiles(method, url string) ([]*zip.File, error) {
	res, err := Fetch(method, url, nil, nil)
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(res.Body), int64(len(res.Body)))
	if err != nil {
		return nil, err
	}

	zipFiles := zipReader.File
	if len(zipFiles) == 0 || zipFiles == nil {
		return nil, fmt.Errorf("no files found in zip while fetching %v", url)
	}

	return zipFiles, nil
}

type Response struct {
	Body       []byte
	Headers    http.Header
	StatusCode int
}
