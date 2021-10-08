package httpctx

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/debugger"
	"github.com/pawelWritesCode/gdutils/pkg/httpcache"
)

var ErrHTTPReqRes = errors.New("something wrong with HTTP(s) request/response")

type HttpContext interface {
	GetHTTPClient() *http.Client
	GetLastResponse() (*http.Response, error)
	GetLastResponseBody() ([]byte, error)
}

type HttpService struct {
	cache    cache.Cache
	debugger debugger.Debugger
	cli      *http.Client
}

func New(c cache.Cache, d debugger.Debugger, cli *http.Client) HttpService {
	return HttpService{cache: c, debugger: d, cli: cli}
}

func (h HttpService) GetHTTPClient() *http.Client {
	return h.cli
}

func (h HttpService) GetLastResponse() (*http.Response, error) {
	respInterface, err := h.cache.GetSaved(httpcache.LastHTTPResponseCacheKey)
	if err != nil {
		return nil, fmt.Errorf("%w: missing last HTTP(s) response, err: %s", ErrHTTPReqRes, err.Error())
	}

	lastResp, ok := respInterface.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("%w: last HTTP(s) response data structure is not type *http.Response", ErrHTTPReqRes)
	}

	return lastResp, nil
}

func (h HttpService) GetLastResponseBody() ([]byte, error) {
	lastResponse, err := h.GetLastResponse()
	if err != nil {
		return []byte(""), err
	}

	var bodyBytes []byte

	if lastResponse != nil {
		bodyBytes, _ = ioutil.ReadAll(lastResponse.Body)
		defer lastResponse.Body.Close()
		lastResponse.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return bodyBytes, nil
}
