package nethttp

import (
	"crypto/tls"
	"net/http"
	"time"
)

// DefaultClient клиент по умолчанию
var DefaultClient = NewClient(DefaultRequestTimeout) // nolint

// NewClient - конструктор HTTP клиента
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
