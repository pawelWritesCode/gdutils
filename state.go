package gdutils

import (
	"crypto/tls"
	"net/http"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/dataformat"
	"github.com/pawelWritesCode/gdutils/pkg/debugger"
	"github.com/pawelWritesCode/gdutils/pkg/formatter"
	"github.com/pawelWritesCode/gdutils/pkg/httpctx"
	"github.com/pawelWritesCode/gdutils/pkg/jsonpath"
	"github.com/pawelWritesCode/gdutils/pkg/template"
	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

// State struct represents data shared across one scenario.
type State struct {
	// Debugger represents scenario debugger.
	Debugger debugger.Debugger

	// Cache is storage for scenario data.
	Cache cache.Cache

	// RequestDoer is service that has ability to send HTTP(s) requests
	RequestDoer httpctx.RequestDoer

	// TemplateEngine is entity that has ability to work with template values.
	TemplateEngine template.Engine

	// JSONSchemaValidators holds validators available to validate JSON Schemas
	JSONSchemaValidators JSONSchemaValidators

	// JSONPathResolver is entity that has ability to obtain data from JSON
	JSONPathResolver jsonpath.Resolver

	// Deserializer is entity that has ability to deserialize data in expected format
	Deserializer formatter.Deserializer
}

// JSONSchemaValidators is container for JSON schema validators
type JSONSchemaValidators struct {
	// StringValidator represents entity that has ability to validate document against string of JSON schema
	StringValidator validator.SchemaValidator
	// ReferenceValidator represents entity that has ability to validate document against JSON schema
	// provided by reference like URL or relative/full OS path
	ReferenceValidator validator.SchemaValidator
}

// NewDefaultState returns *State with default services.
// jsonSchemaDir may be empty string or valid full path to directory with JSON schemas
func NewDefaultState(isDebug bool, jsonSchemaDir string) *State {
	defaultCache := cache.NewConcurrentCache()
	defaultHttpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	jsonSchemaValidators := JSONSchemaValidators{
		StringValidator:    dataformat.NewJSONSchemaRawValidator(),
		ReferenceValidator: dataformat.NewDefaultJSONSchemaReferenceValidator(jsonSchemaDir),
	}

	resolver := jsonpath.NewDynamicJSONPathResolver(jsonpath.NewQJSONResolver(), jsonpath.NewOliveagleJSONpath())
	deserializer := formatter.NewAwareFormatter(formatter.NewJSONFormatter(), formatter.NewYAMLFormatter())

	return NewState(defaultHttpClient, defaultCache, jsonSchemaValidators, resolver, deserializer, isDebug)
}

// NewState returns *State
func NewState(cli *http.Client, c cache.Cache, jv JSONSchemaValidators, r jsonpath.Resolver, d formatter.Deserializer, isDebug bool) *State {
	defaultDebugger := debugger.New(isDebug)
	return &State{
		Debugger:             defaultDebugger,
		Cache:                c,
		RequestDoer:          cli,
		TemplateEngine:       template.New(),
		JSONSchemaValidators: jv,
		JSONPathResolver:     r,
		Deserializer:         d,
	}
}

// ResetState resets State struct instance to default values.
func (s *State) ResetState(isDebug bool) {
	s.Cache.Reset()
	s.Debugger.Reset(isDebug)
}

// SetDebugger sets new debugger for State
func (s *State) SetDebugger(d debugger.Debugger) {
	s.Debugger = d
}

// SetCache sets new Cache for State
func (s *State) SetCache(c cache.Cache) {
	s.Cache = c
}

// SetRequestDoer sets new RequestDoer for State
func (s *State) SetRequestDoer(r httpctx.RequestDoer) {
	s.RequestDoer = r
}

// SetTemplateEngine sets new template Engine for State
func (s *State) SetTemplateEngine(t template.Engine) {
	s.TemplateEngine = t
}

// SetJSONSchemaValidators sets new JSONSchemaValidators for State
func (s *State) SetJSONSchemaValidators(j JSONSchemaValidators) {
	s.JSONSchemaValidators = j
}

// SetJSONPathResolver sets new JSON path resolver for State
func (s *State) SetJSONPathResolver(j jsonpath.Resolver) {
	s.JSONPathResolver = j
}

// SetDeserializer sets new deserializer for State
func (s *State) SetDeserializer(d formatter.Deserializer) {
	s.Deserializer = d
}
