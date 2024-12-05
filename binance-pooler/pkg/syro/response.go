package syro

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Standartized response of the API
type ApiResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Params  any    `json:"params,omitempty"`
}

// ParseLogsQuery parses the query parameters from the URL and returns a LogFilter
func ParseLogsQuery(fullUrl string) (*LogFilter, error) {
	// Parse the full URL
	parsedURL, err := url.Parse(fullUrl)
	if err != nil {
		return nil, errors.New("failed to parse URL")
	}

	// Extract query parameters
	params := parsedURL.Query()
	filter := LogFilter{}

	// Parse "from" time
	if from := params.Get("from"); from != "" {
		parsedFrom, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return nil, fmt.Errorf("invalid 'from' time format: %v", err)
		}
		filter.From = parsedFrom
	}

	// Parse "to" time
	if to := params.Get("to"); to != "" {
		parsedTo, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return nil, fmt.Errorf("invalid 'to' time format: %v", err)
		}
		filter.To = parsedTo
	}

	// Parse "limit"
	if limit := params.Get("limit"); limit != "" {
		parsedLimit, err := strconv.ParseInt(limit, 10, 64)
		if err != nil || parsedLimit < 0 {
			return nil, errors.New("invalid 'limit' value")
		}
		filter.Limit = parsedLimit
	}

	// Parse "skip"
	if skip := params.Get("skip"); skip != "" {
		parsedSkip, err := strconv.ParseInt(skip, 10, 64)
		if err != nil || parsedSkip < 0 {
			return nil, errors.New("invalid 'skip' value")
		}
		filter.Skip = parsedSkip
	}

	filter.Source = params.Get("source")
	filter.Event = params.Get("event")
	filter.EventID = params.Get("event_id")

	if parsedLevel, err := strconv.Atoi(params.Get("level")); err == nil {
		logLevel := LogLevel(parsedLevel)
		filter.Level = &logLevel
	}

	return &filter, nil
}
