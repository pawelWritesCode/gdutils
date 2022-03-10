package gdutils

import (
	"crypto/tls"
	"net/http"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/debugger"
	"github.com/pawelWritesCode/gdutils/pkg/formatter"
	"github.com/pawelWritesCode/gdutils/pkg/httpctx"
	"github.com/pawelWritesCode/gdutils/pkg/osutils"
	"github.com/pawelWritesCode/gdutils/pkg/pathfinder"
	"github.com/pawelWritesCode/gdutils/pkg/schema"
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

	// SchemaValidators holds validators available to validate JSON Schemas.
	SchemaValidators SchemaValidators

	// PathFinders are entities that has ability to obtain data from different data formats.
	PathFinders PathFinders

	// Formatters are entities that has ability to deserialize data from one of available formats.
	Formatters Formatters

	// fileRecognizer is entity that has ability to recognize file reference.
	fileRecognizer osutils.FileRecognizer
}

// Formatters is container for entities that know how to serialize and deserialize data.
type Formatters struct {
	// JSON is entity that has ability to serialize and deserialize JSON bytes.
	JSON formatter.Formatter

	// YAML is entity that has ability to serialize and deserialize YAML bytes.
	YAML formatter.Formatter
}

// SchemaValidators is container for JSON schema validators.
type SchemaValidators struct {
	// StringValidator represents entity that has ability to validate document against string of containing schema.
	StringValidator validator.SchemaValidator

	// ReferenceValidator represents entity that has ability to validate document against string with reference
	// to schema, which may be URL or relative/full OS path for example.
	ReferenceValidator validator.SchemaValidator
}

// PathFinders is container for different data types pathfinders.
type PathFinders struct {
	// JSON is entity that has ability to obtain data from bytes in JSON format.
	JSON pathfinder.PathFinder

	// YAML is entity that has ability to obtain data from bytes in YAML format.
	YAML pathfinder.PathFinder
}

// NewDefaultState returns *State with default services.
// jsonSchemaDir may be empty string or valid full path to directory with JSON schemas.
func NewDefaultState(isDebug bool, jsonSchemaDir string) *State {
	defaultCache := cache.NewConcurrentCache()
	defaultHttpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	jsonSchemaValidators := SchemaValidators{
		StringValidator:    schema.NewJSONSchemaRawValidator(),
		ReferenceValidator: schema.NewDefaultJSONSchemaReferenceValidator(jsonSchemaDir),
	}

	pathResolvers := PathFinders{
		JSON: pathfinder.NewDynamicJSONPathFinder(pathfinder.NewQJSONFinder(), pathfinder.NewOliveagleJSONFinder()),
		YAML: pathfinder.NewGoccyGoYamlFinder(),
	}

	deserializers := Formatters{
		JSON: formatter.NewJSONFormatter(),
		YAML: formatter.NewYAMLFormatter(),
	}

	return NewState(defaultHttpClient, defaultCache, jsonSchemaValidators, pathResolvers, deserializers, isDebug)
}

// NewState returns *State
func NewState(cli *http.Client, c cache.Cache, jv SchemaValidators, p PathFinders, d Formatters, isDebug bool) *State {
	fileRecognizer := osutils.NewOSFileRecognizer("file://", osutils.NewFileValidator())
	defaultDebugger := debugger.New(isDebug)
	return &State{
		Debugger:         defaultDebugger,
		Cache:            c,
		RequestDoer:      cli,
		TemplateEngine:   template.New(),
		SchemaValidators: jv,
		PathFinders:      p,
		Formatters:       d,
		fileRecognizer:   fileRecognizer,
	}
}

// ResetState resets State struct instance to default values.
func (s *State) ResetState(isDebug bool) {
	s.Cache.Reset()
	s.Debugger.Reset(isDebug)
}

// SetDebugger sets new debugger for State.
func (s *State) SetDebugger(d debugger.Debugger) {
	s.Debugger = d
}

// SetCache sets new Cache for State.
func (s *State) SetCache(c cache.Cache) {
	s.Cache = c
}

// SetRequestDoer sets new RequestDoer for State.
func (s *State) SetRequestDoer(r httpctx.RequestDoer) {
	s.RequestDoer = r
}

// SetTemplateEngine sets new template Engine for State.
func (s *State) SetTemplateEngine(t template.Engine) {
	s.TemplateEngine = t
}

// SetSchemaStringValidator sets new schema StringValidator for State.
func (s *State) SetSchemaStringValidator(j validator.SchemaValidator) {
	s.SchemaValidators.StringValidator = j
}

// SetSchemaReferenceValidator sets new schema ReferenceValidator for State.
func (s *State) SetSchemaReferenceValidator(j validator.SchemaValidator) {
	s.SchemaValidators.ReferenceValidator = j
}

// SetJSONPathFinder sets new JSON pathfinder for State.
func (s *State) SetJSONPathFinder(r pathfinder.PathFinder) {
	s.PathFinders.JSON = r
}

// SetYAMLPathFinder sets new YAML pathfinder for State.
func (s *State) SetYAMLPathFinder(r pathfinder.PathFinder) {
	s.PathFinders.YAML = r
}

// SetJSONFormatter sets new JSON formatter for State.
func (s *State) SetJSONFormatter(jf formatter.Formatter) {
	s.Formatters.JSON = jf
}

// SetYAMLFormatter sets new YAML formatter for State.
func (s *State) SetYAMLFormatter(yd formatter.Formatter) {
	s.Formatters.YAML = yd
}
