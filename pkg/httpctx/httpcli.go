package httpctx

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/httpcache"
)

//ErrHTTPReqRes occurs when there is any problem with HTTP(s) req/res.
var ErrHTTPReqRes = errors.New("something wrong with HTTP(s) request/response")

//HttpContext represents set of operations related with HTTP.
type HttpContext interface {
	GetHTTPClient() *http.Client
	GetLastResponse() (*http.Response, error)
	GetLastResponseBody() ([]byte, error)
}

//HttpService is entity that implements HttpContext interface.
type HttpService struct {
	//cache is place where last HTTP(s) resp is stored
	cache cache.Cache

	//cli is entity that has ability to send HTTP(s) requests
	cli *http.Client
}

func New(c cache.Cache, cli *http.Client) HttpService {
	return HttpService{cache: c, cli: cli}
}

//GetHTTPClient returns HTTP client.
func (h HttpService) GetHTTPClient() *http.Client {
	return h.cli
}

//GetLastResponse returns last HTTP(s) response.
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

//GetLastResponseBody returns last HTTP(s) response body.
//internally method creates new NoPCloser on last response so this method is safe to reuse many times
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
