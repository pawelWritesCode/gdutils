package gdutils

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/datatype"
	"github.com/pawelWritesCode/gdutils/pkg/debugger"
	"github.com/pawelWritesCode/gdutils/pkg/httpctx"
	"github.com/pawelWritesCode/gdutils/pkg/template"
	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

// State struct represents data shared across one scenario.
type State struct {
	// Debugger represents scenario debugger.
	Debugger debugger.Debugger

	// Cache is storage for scenario data.
	Cache cache.Cache

	// HttpContext is service that works with HTTP(s) req/res.
	HttpContext httpctx.HttpContext

	// TemplateEngine is entity that has ability to work with template values.
	TemplateEngine template.TemplateEngine

	// JSONSchemaValidator is entity that has ability to validate JSON schemas.
	JSONSchemaValidator validator.SchemaValidator
}

// NewDefaultState returns *State with default http.Client, DefaultCache and default Debugger.
// jsonSchemaDir may be empty string or valid full path to directory with JSON schemas
func NewDefaultState(isDebug bool, jsonSchemaDir string) *State {
	defaultCache := cache.NewConcurrentCache()
	defaultHttpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	jsonSchemaValidator := datatype.NewDefaultJSONSchemaValidator(jsonSchemaDir)

	return NewState(defaultHttpClient, defaultCache, jsonSchemaValidator, isDebug)
}

// NewState returns *State with provided HttpClient, cache and default Debugger
func NewState(httpClient *http.Client, cache cache.Cache, jsonSchemaValidator validator.SchemaValidator, isDebug bool) *State {
	defaultDebugger := debugger.New(isDebug)
	return &State{
		Debugger:            defaultDebugger,
		Cache:               cache,
		HttpContext:         httpctx.NewHttpService(cache, httpClient),
		TemplateEngine:      template.New(),
		JSONSchemaValidator: jsonSchemaValidator,
	}
}

// ResetState resets State struct instance to default values.
func (s *State) ResetState(isDebug bool) {
	s.Cache.Reset()
	s.Debugger.Reset(isDebug)
}

// getPreparedRequest returns prepared request from cache or error if failed
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
