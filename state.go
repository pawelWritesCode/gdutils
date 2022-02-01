package gdutils

import (
	"crypto/tls"
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

	// JSONSchemaValidators holds validators available to validate JSON Schemas
	JSONSchemaValidators JSONSchemaValidators
}

// JSONSchemaValidators is container for JSON schema validators
type JSONSchemaValidators struct {
	// StringValidator represents entity that has ability to validate document against string of JSON schema
	StringValidator validator.SchemaValidator
	// ReferenceValidator represents entity that has ability to validate document against JSON schema
	// provided by reference like URL or relative/full OS path
	ReferenceValidator validator.SchemaValidator
}

// NewDefaultState returns *State with default http.Client, DefaultCache and default Debugger.
// jsonSchemaDir may be empty string or valid full path to directory with JSON schemas
func NewDefaultState(isDebug bool, jsonSchemaDir string) *State {
	defaultCache := cache.NewConcurrentCache()
	defaultHttpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	jsonSchemaValidators := JSONSchemaValidators{
		StringValidator:    datatype.NewJSONSchemaRawValidator(),
		ReferenceValidator: datatype.NewDefaultJSONSchemaReferenceValidator(jsonSchemaDir),
	}

	return NewState(defaultHttpClient, defaultCache, jsonSchemaValidators, isDebug)
}

// NewState returns *State with provided HttpClient, cache and default Debugger
func NewState(httpClient *http.Client, cache cache.Cache, jsonSchemaValidators JSONSchemaValidators, isDebug bool) *State {
	defaultDebugger := debugger.New(isDebug)
	return &State{
		Debugger:             defaultDebugger,
		Cache:                cache,
		HttpContext:          httpctx.NewHttpService(cache, httpClient),
		TemplateEngine:       template.New(),
		JSONSchemaValidators: jsonSchemaValidators,
	}
}

// ResetState resets State struct instance to default values.
func (s *State) ResetState(isDebug bool) {
	s.Cache.Reset()
	s.Debugger.Reset(isDebug)
}
