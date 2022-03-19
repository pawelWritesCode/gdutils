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

// APIContext holds utility services for working with HTTP(s) API.
type APIContext struct {
	// Debugger represents debugger.
	Debugger debugger.Debugger

	// Cache is storage for data.
	Cache cache.Cache

	// RequestDoer is service that has ability to send HTTP(s) requests.
	RequestDoer httpctx.RequestDoer

	// TemplateEngine is entity that has ability to work with template values.
	TemplateEngine template.Engine

	// SchemaValidators holds validators available to validate data against schemas.
	SchemaValidators SchemaValidators

	// PathFinders are entities that has ability to obtain data from different data formats.
	PathFinders PathFinders

	// Formatters are entities that has ability to format data in particular format.
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

	// XML is entity that has ability to serialize and deserialize XML bytes.
	XML formatter.Formatter
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

	// XML is entity that has ability to obtain data from bytes in XML format.
	XML pathfinder.PathFinder
}

// NewDefaultAPIContext returns *APIContext with default services.
// jsonSchemaDir may be empty string or valid full path to directory with JSON schemas.
func NewDefaultAPIContext(isDebug bool, jsonSchemaDir string) *APIContext {
	defaultCache := cache.NewConcurrentCache()
	defaultHttpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	jsonSchemaValidators := SchemaValidators{
		StringValidator:    schema.NewJSONSchemaRawValidator(),
		ReferenceValidator: schema.NewDefaultJSONSchemaReferenceValidator(jsonSchemaDir),
	}

	pathFinders := PathFinders{
		JSON: pathfinder.NewDynamicJSONPathFinder(pathfinder.NewQJSONFinder(), pathfinder.NewOliveagleJSONFinder()),
		YAML: pathfinder.NewGoccyGoYamlFinder(),
		XML:  pathfinder.NewAntchfxXMLFinder(),
	}

	formatters := Formatters{
		JSON: formatter.NewJSONFormatter(),
		YAML: formatter.NewYAMLFormatter(),
		XML:  formatter.NewXMLFormatter(),
	}

	defaultDebugger := debugger.New(isDebug)

	return NewAPIContext(defaultHttpClient, defaultCache, jsonSchemaValidators, pathFinders, formatters, defaultDebugger)
}

// NewAPIContext returns *APIContext
func NewAPIContext(cli *http.Client, c cache.Cache, jv SchemaValidators, p PathFinders, f Formatters, d debugger.Debugger) *APIContext {
	fileRecognizer := osutils.NewOSFileRecognizer("file://", osutils.NewFileValidator())

	return &APIContext{
		Debugger:         d,
		Cache:            c,
		RequestDoer:      cli,
		TemplateEngine:   template.New(),
		SchemaValidators: jv,
		PathFinders:      p,
		Formatters:       f,
		fileRecognizer:   fileRecognizer,
	}
}

// ResetState resets state of APIContext to initial.
func (apiCtx *APIContext) ResetState(isDebug bool) {
	apiCtx.Cache.Reset()
	apiCtx.Debugger.Reset(isDebug)
}

// SetDebugger sets new debugger for APIContext.
func (apiCtx *APIContext) SetDebugger(d debugger.Debugger) {
	apiCtx.Debugger = d
}

// SetCache sets new Cache for APIContext.
func (apiCtx *APIContext) SetCache(c cache.Cache) {
	apiCtx.Cache = c
}

// SetRequestDoer sets new RequestDoer for APIContext.
func (apiCtx *APIContext) SetRequestDoer(r httpctx.RequestDoer) {
	apiCtx.RequestDoer = r
}

// SetTemplateEngine sets new template Engine for APIContext.
func (apiCtx *APIContext) SetTemplateEngine(t template.Engine) {
	apiCtx.TemplateEngine = t
}

// SetSchemaStringValidator sets new schema StringValidator for APIContext.
func (apiCtx *APIContext) SetSchemaStringValidator(j validator.SchemaValidator) {
	apiCtx.SchemaValidators.StringValidator = j
}

// SetSchemaReferenceValidator sets new schema ReferenceValidator for APIContext.
func (apiCtx *APIContext) SetSchemaReferenceValidator(j validator.SchemaValidator) {
	apiCtx.SchemaValidators.ReferenceValidator = j
}

// SetJSONPathFinder sets new JSON pathfinder for APIContext.
func (apiCtx *APIContext) SetJSONPathFinder(r pathfinder.PathFinder) {
	apiCtx.PathFinders.JSON = r
}

// SetYAMLPathFinder sets new YAML pathfinder for APIContext.
func (apiCtx *APIContext) SetYAMLPathFinder(r pathfinder.PathFinder) {
	apiCtx.PathFinders.YAML = r
}

// SetXMLPathFinder sets new XML pathfinder for APIContext.
func (apiCtx *APIContext) SetXMLPathFinder(r pathfinder.PathFinder) {
	apiCtx.PathFinders.XML = r
}

// SetJSONFormatter sets new JSON formatter for APIContext.
func (apiCtx *APIContext) SetJSONFormatter(jf formatter.Formatter) {
	apiCtx.Formatters.JSON = jf
}

// SetYAMLFormatter sets new YAML formatter for APIContext.
func (apiCtx *APIContext) SetYAMLFormatter(yd formatter.Formatter) {
	apiCtx.Formatters.YAML = yd
}

// SetXMLFormatter sets new XML formatter for APIContext.
func (apiCtx *APIContext) SetXMLFormatter(xf formatter.Formatter) {
	apiCtx.Formatters.XML = xf
}
