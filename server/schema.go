package server

import "net/url"
import "regexp"

import (
	msconfig "mockserver/config"
	server_utils "mockserver/server/utils"
)

type MockHandler struct {
	routeName    string
	filePath     string
	status       int
	headers      map[string]string
	delayMs      int
	mockFileData []byte
	mockBodyData interface{}
	stateStore   *server_utils.StateStore
	routecfg     msconfig.RouteConfig
}

type FetchHandler struct {
	routeName string
	targetURL *url.URL
	method    string
	headers   map[string]string
	queryParams      map[string]struct{}
	fetchQueryParams map[string]string
	passStatus       bool
	delayMs          int
	timeoutMs        int
	urlRegex         *regexp.Regexp
	basePath         string
}

// ApiError represents a structured API error response.
type ApiError struct {
	Success   bool   `json:"success"`
	Status    int    `json:"status"`
	Err       string `json:"error"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}
