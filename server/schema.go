package server

import "net/url"
import "regexp"

type MockHandler struct {
	filePath string
	status   int
	headers  map[string]string
	delayMs  int
	data     []byte
}


type FetchHandler struct {
	targetURL   *url.URL
	method      string
	headers     map[string]string
	queryParams map[string]string
	passStatus  bool
	delayMs     int
	timeoutMs   int
	urlRegex *regexp.Regexp
	basePath string
}

// ApiError defines structured error responses
type ApiError struct {
	Success   bool   `json:"success"`
	Status    int    `json:"status"`
	Err       string `json:"error"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}
