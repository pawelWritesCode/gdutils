package gdutils

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/debugger"
	"github.com/pawelWritesCode/gdutils/pkg/httpctx"
	"github.com/pawelWritesCode/gdutils/pkg/template"
)

//State struct represents data shared across one scenario.
type State struct {
	//IsDebug determine whether scenario is in debug mode
	Debugger debugger.Debugger
	//Cache is storage for scenario data
	Cache cache.Cache
	//HttpClient is entity that has ability to send HTTP(s) requests
	HttpContext httpctx.HttpContext
	//TemplateEngine is entity that has ability to retrieve templated values
	TemplateEngine template.TemplateEngine
}

//NewDefaultState returns *State with default http.Client, DefaultCache and default Debugger
func NewDefaultState(isDebug bool) *State {
	defaultCache := cache.New()
	defaultHttpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	return NewState(defaultHttpClient, defaultCache, isDebug)
}

//NewState returns *State with provided HttpClient, cache and default Debugger
func NewState(httpClient *http.Client, cache cache.Cache, isDebug bool) *State {
	defaultDebugger := debugger.New(isDebug)
	return &State{
		Debugger:       defaultDebugger,
		Cache:          cache,
		HttpContext:    httpctx.New(cache, defaultDebugger, httpClient),
		TemplateEngine: template.New(),
	}
}

//ResetState resets State struct instance to default values.
func (s *State) ResetState(isDebug bool) {
	s.Cache.Reset()
	s.Debugger.Reset(isDebug)
}

//getPreparedRequest returns prepared request from cache or error if failed
func (s *State) getPreparedRequest(cacheKey string) (*http.Request, error) {
	reqInterface, err := s.Cache.GetSaved(cacheKey)
	if err != nil {
		return &http.Request{}, err
	}

	req, ok := reqInterface.(*http.Request)
	if !ok {
		return &http.Request{}, fmt.Errorf("%w: value under key %s in cache doesn't contain *http.Request", ErrPreservedData, cacheKey)
	}

	return req, nil
}
