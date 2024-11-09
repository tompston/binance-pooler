package fetcher

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Response struct {
	Body       []byte
	Headers    http.Header
	StatusCode int
}

// JsonHeader is a map of headers for json requests
var JsonHeader = map[string]string{"Content-Type": "application/json"}
var XmlHeader = map[string]string{"Content-Type": "application/xml"}

type Params struct {
	RequestBody       []byte
	IgnoreStatusCodes bool
	WithSleepInSec    int
	TLSClientConfig   *tls.Config
}

// Baseline of fetch request.
func Fetch(method, url string, headers map[string]string, optParams ...Params) (*Response, error) {

	var params Params
	if len(optParams) == 1 {
		params = optParams[0]
	}

	var reqBody []byte
	if params.RequestBody != nil {
		reqBody = params.RequestBody
	}

	if params.WithSleepInSec > 0 {
		time.Sleep(time.Duration(params.WithSleepInSec) * time.Second)
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
	if params.TLSClientConfig != nil {
		client.Transport = &http.Transport{TLSClientConfig: params.TLSClientConfig}
	}

	// fmt.Printf("client.Transport: %v\n", client.Transport)

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching %v : %v", url, err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body when fetching %v, status %v: %v, error: ", url, res.Status, err)
	}

	if params.IgnoreStatusCodes {
		return &Response{body, res.Header, res.StatusCode}, nil
	}

	if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 202 {
		bodyStr := ""
		if body != nil {
			bodyUpTo := len(body)
			if 1000 < bodyUpTo {
				bodyUpTo = 1000 // limit to 1000 characters
			}

			// checked, if body contains wrong char codes, code will still work
			// if body is nil it will work too
			bodyStr = string(body[:bodyUpTo])
		}

		return nil, fmt.Errorf("response did not return status in 200 group while requesting %v, status: %v, body: %v", url, res.Status, bodyStr)
	}

	if body == nil {
		return nil, fmt.Errorf("response returned empty body while requesting %v", url)
	}

	return &Response{body, res.Header, res.StatusCode}, err
}

// NOTE: i'm not sure if this is the best way to check for the files
// (aka if there are edge cases where this won't work)
func IsZipFile(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// ZIP file magic number (first 4 bytes)
	return bytes.Equal(data[:4], []byte{0x50, 0x4b, 0x03, 0x04})
}

// EncodeString for urls
func EncodeString(str string) string { return url.QueryEscape(str) }

// DecodeString decodes url params and returns a normalized string
func DecodeString(str string) (string, error) { return url.QueryUnescape(str) }
