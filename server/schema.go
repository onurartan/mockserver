package server

import "net/url"
import "regexp"

type MockHandler struct {
	routeName string
	filePath string
	status   int
	headers  map[string]string
	delayMs  int
	data     []byte
}


type FetchHandler struct {
	routeName string
	targetURL   *url.URL
	method      string
	headers     map[string]string
	// [DEPRACTED] queryParams map[string]string
	queryParams map[string]struct{}
	fetchQueryParams map[string]string
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
