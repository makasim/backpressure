package backpressure

import (
	"net/http"
	"net/url"
)

func IsResponseCongested(resp http.Response, err error) bool {
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return true
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		return true
	}

	// todo: check for other possible cases that fail into this case
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				return true
			}
		}
	}

	return false
}
