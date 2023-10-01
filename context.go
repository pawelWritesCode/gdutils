package gdutils

import (
	"crypto/tls"
	"net/http"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/debugger"
	"github.com/pawelWritesCode/gdutils/pkg/osutils"
	"github.com/pawelWritesCode/gdutils/pkg/pathfinder"
	"github.com/pawelWritesCode/gdutils/pkg/schema"
	"github.com/pawelWritesCode/gdutils/pkg/serializer"
	"github.com/pawelWritesCode/gdutils/pkg/template"
	"github.com/pawelWritesCode/gdutils/pkg/types"
	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

// APIContext holds utility services for working with HTTP(s) API.
type APIContext struct {
	// Debugger represents debugger.
	Debugger debuggable

	// Cache is storage for data.
	Cache cacheable

	// RequestDoer is service that has ability to send HTTP(s) requests.
	RequestDoer requestDoer

	// TemplateEngine is entity that has ability to work with template values.
	TemplateEngine templateEngine

	// SchemaValidators holds validators available to validate data against schemas.
	SchemaValidators SchemaValidators

	// PathFinders are entities that has ability to obtain data from different data formats.
	PathFinders PathFinders

	// Serializers are entities that has ability to serialize and deserialize data in particular format.
	Serializers Serializers

	// TypeMappers are entities that has ability to map underlying data type into different format data type.
	TypeMappers TypeMappers

	// fileRecognizer is entity that has ability to recognize file reference.
	fileRecognizer fileRecognizer
}

// Serializers is container for entities that know how to serialize and deserialize data.
type Serializers struct {
	// JSON is entity that has ability to serialize and deserialize JSON bytes.
	JSON serializable

	// YAML is entity that has ability to serialize and deserialize YAML bytes.
	YAML serializable

	// XML is entity that has ability to serialize and deserialize XML bytes.
	XML serializable
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
	JSON pathFinder

	// YAML is entity that has ability to obtain data from bytes in YAML format.
	YAML pathFinder

	// XML is entity that has ability to obtain data from bytes in XML format.
	XML pathFinder

	// HTML is entity that has ability to obtain data from bytes in HTML format.
	HTML pathFinder
}

// TypeMappers is container for different data format mappers
type TypeMappers struct {
	// JSON is entity that has ability to map underlying data type into JSON data type
	JSON typeMapper

	// YAML is entity that has ability to map underlying data type into YAML data type
	YAML typeMapper

	// GO is entity that has ability to map underlying data type into GO-like data type
	GO typeMapper
}

type CustomTransport struct {
	http.RoundTripper
}

func (ct *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "gdutils")
	return ct.RoundTripper.RoundTrip(req)
}

var DefaultTransport http.RoundTripper = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

// NewDefaultAPIContext returns *APIContext with default services.
// jsonSchemaDir may be empty string or valid full path to directory with JSON schemas.
func NewDefaultAPIContext(isDebug bool, jsonSchemaDir string) *APIContext {
	defaultCache := cache.NewConcurrentCache()
	tr := &CustomTransport{DefaultTransport}

	defaultHttpClient := &http.Client{Transport: tr}

	jsonSchemaValidators := SchemaValidators{
		StringValidator:    schema.NewJSONSchemaRawXGValidator(),
		ReferenceValidator: schema.NewDefaultJSONSchemaReferenceXGValidator(jsonSchemaDir),
	}

	pathFinders := PathFinders{
		JSON: pathfinder.NewDynamicJSONPathFinder(
			pathfinder.NewGJSONFinder(),
			pathfinder.NewOliveagleJSONFinder(),
			pathfinder.NewAntchfxJSONQueryFinder(),
		),
		YAML: pathfinder.NewGoccyGoYamlFinder(),
		XML:  pathfinder.NewAntchfxXMLFinder(),
		HTML: pathfinder.NewAntchfxHTMLFinder(),
	}

	serializers := Serializers{
		JSON: serializer.NewJSONFormatter(),
		YAML: serializer.NewYAMLFormatter(),
		XML:  serializer.NewXMLFormatter(),
	}

	typeMappers := TypeMappers{
		JSON: types.NewJSONTypeMapper(),
		YAML: types.NewYAMLTypeMapper(),
		GO:   types.NewGoTypeMapper(),
	}

	defaultDebugger := debugger.NewDefault(isDebug)

	return NewAPIContext(defaultHttpClient, defaultCache, jsonSchemaValidators, pathFinders, serializers, typeMappers, defaultDebugger)
}

// NewAPIContext returns *APIContext
func NewAPIContext(cli *http.Client, c cacheable, jv SchemaValidators, p PathFinders, s Serializers, t TypeMappers, d debuggable) *APIContext {
	return &APIContext{
		Debugger:         d,
		Cache:            c,
		RequestDoer:      cli,
		TemplateEngine:   template.New(),
		SchemaValidators: jv,
		PathFinders:      p,
		Serializers:      s,
		TypeMappers:      t,
		fileRecognizer:   osutils.NewOSFileRecognizer("file://", osutils.NewFileValidator()),
	}
}

// ResetState resets state of APIContext to initial.
func (apiCtx *APIContext) ResetState(isDebug bool) {
	apiCtx.Cache.Reset()
	apiCtx.Debugger.Reset(isDebug)
}

// SetDebugger sets new debugger for APIContext.
func (apiCtx *APIContext) SetDebugger(d debuggable) {
	apiCtx.Debugger = d
}

// SetCache sets new cacheable for APIContext.
func (apiCtx *APIContext) SetCache(c cacheable) {
	apiCtx.Cache = c
}

// SetRequestDoer sets new RequestDoer for APIContext.
func (apiCtx *APIContext) SetRequestDoer(r requestDoer) {
	apiCtx.RequestDoer = r
}

// SetTemplateEngine sets new template Engine for APIContext.
func (apiCtx *APIContext) SetTemplateEngine(t templateEngine) {
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
func (apiCtx *APIContext) SetJSONPathFinder(r pathFinder) {
	apiCtx.PathFinders.JSON = r
}

// SetYAMLPathFinder sets new YAML pathfinder for APIContext.
func (apiCtx *APIContext) SetYAMLPathFinder(r pathFinder) {
	apiCtx.PathFinders.YAML = r
}

// SetXMLPathFinder sets new XML pathfinder for APIContext.
func (apiCtx *APIContext) SetXMLPathFinder(r pathFinder) {
	apiCtx.PathFinders.XML = r
}

// SetHTMLPathFinder sets new HTML pathfinder for APIContext.
func (apiCtx *APIContext) SetHTMLPathFinder(r pathFinder) {
	apiCtx.PathFinders.HTML = r
}

// SetJSONSerializer sets new JSON serializer for APIContext.
func (apiCtx *APIContext) SetJSONSerializer(jf serializable) {
	apiCtx.Serializers.JSON = jf
}

// SetYAMLSerializer sets new YAML serializer for APIContext.
func (apiCtx *APIContext) SetYAMLSerializer(yd serializable) {
	apiCtx.Serializers.YAML = yd
}

// SetXMLSerializer sets new XML serializer for APIContext.
func (apiCtx *APIContext) SetXMLSerializer(xf serializable) {
	apiCtx.Serializers.XML = xf
}

// SetJSONTypeMapper sets new type mapper for JSON.
func (apiCtx *APIContext) SetJSONTypeMapper(c typeMapper) {
	apiCtx.TypeMappers.JSON = c
}

// SetYAMLTypeMapper sets new type mapper for YAML.
func (apiCtx *APIContext) SetYAMLTypeMapper(c typeMapper) {
	apiCtx.TypeMappers.YAML = c
}

// SetGoTypeMapper sets new type mapper for Go.
func (apiCtx *APIContext) SetGoTypeMapper(c typeMapper) {
	apiCtx.TypeMappers.GO = c
}
