// Package httpctx holds utilities for working with HTTP protocol.
package httpctx

import (
	"errors"
	"net/http"
)

// ErrHTTPReqRes occurs when there is any problem with HTTP(s) req/res.
var ErrHTTPReqRes = errors.New("something wrong with HTTP(s) request/response")

// RequestDoer describes ability to make HTTP(s) requests.
type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}
