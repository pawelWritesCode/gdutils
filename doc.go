// Package gdutils provides APIContext struct with methods that may be used for behavioral testing of HTTP API.
//
// APIContext may be initialized by two ways:
//
// First, returns *APIContext with default services:
//	func NewDefaultAPIContext(isDebug bool, jsonSchemaDir string) *APIContext
//
// Second, more customisable returns *APIContext with provided services:
//	func NewAPIContext(cli *http.Client, c cache.Cache, jv SchemaValidators, p PathFinders, f Formatters, d debugger.Debugger) *APIContext
//
// No matter which way you choose, you can inject your custom services afterwards with one of available setters:
//	func (apiCtx *APIContext) SetDebugger(d debugger.Debugger)
//	func (apiCtx *APIContext) SetCache(c cache.Cache)
//	func (apiCtx *APIContext) SetRequestDoer(r httpctx.RequestDoer)
//	func (apiCtx *APIContext) SetTemplateEngine(t template.Engine)
//	func (apiCtx *APIContext) SetSchemaStringValidator(j validator.SchemaValidator)
//	func (apiCtx *APIContext) SetSchemaReferenceValidator(j validator.SchemaValidator)
//	func (apiCtx *APIContext) SetJSONPathFinder(r pathfinder.PathFinder)
//	func (apiCtx *APIContext) SetJSONFormatter(jf formatter.Formatter)
//	func (apiCtx *APIContext) SetXMLPathFinder(r pathfinder.PathFinder)
//	func (apiCtx *APIContext) SetYAMLPathFinder(r pathfinder.PathFinder)
//	func (apiCtx *APIContext) SetYAMLFormatter(yd formatter.Formatter)
//	func (apiCtx *APIContext) SetXMLFormatter(xf formatter.Formatter)
//
// Those services will be used in utility methods.
// For example, if you want to use your own debugger, create your own struct, implement debugger.Debugger interface on it,
// and then inject it with "func (apiCtx *APIContext) SetDebugger(d debugger.Debugger)" method.
//
// Testing HTTP API usually consist the following aspects:
//
// * Data generation:
//
//	func (apiCtx *APIContext) IGenerateARandomIntInTheRangeToAndSaveItAs(from, to int, cacheKey string) error
//	func (apiCtx *APIContext) IGenerateARandomFloatInTheRangeToAndSaveItAs(from, to int, cacheKey string) error
//  func (apiCtx *APIContext) GeneratorRandomRunes(charset string) func(from, to int, cacheKey string) error
//	func (apiCtx *APIContext) GeneratorRandomSentence(charset string, wordMinLength, wordMaxLength int) func(from, to int, cacheKey string) error
//	func (apiCtx *APIContext) IGetTimeAndTravelByAndSaveItAs(t time.Time, timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error
//	func (apiCtx *APIContext) IGenerateCurrentTimeAndTravelByAndSaveItAs(timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error
//
// * Sending HTTP(s) requests:
//
//	func (apiCtx *APIContext) ISendRequestToWithBodyAndHeaders(method, urlTemplate string, bodyTemplate string) error
//
// or
//
//	func (apiCtx *APIContext) IPrepareNewRequestToAndSaveItAs(method, urlTemplate, cacheKey string) error
//	func (apiCtx *APIContext) ISetFollowingHeadersForPreparedRequest(cacheKey string, headersTemplate string) error
//	func (apiCtx *APIContext) ISetFollowingFormForPreparedRequest(cacheKey, formTemplate string) error
//	func (apiCtx *APIContext) ISetFollowingCookiesForPreparedRequest(cacheKey, cookiesTemplate string) error
//	func (apiCtx *APIContext) ISetFollowingBodyForPreparedRequest(cacheKey string, bodyTemplate string) error
//	func (apiCtx *APIContext) ISendRequest(cacheKey string) error
//
// * Assertions:
//
//	func (apiCtx *APIContext) TheResponseStatusCodeShouldBe(code int) error
//	func (apiCtx *APIContext) TheResponseBodyShouldHaveFormat(dataFormat format.DataFormat) error
//	func (apiCtx *APIContext) TheResponseShouldHaveCookie(name string) error
//	func (apiCtx *APIContext) TheResponseShouldHaveCookieOfValue(name, valueTemplate string) error
//	func (apiCtx *APIContext) TheResponseShouldHaveNode(dataFormat format.DataFormat, exprTemplate string) error
//	func (apiCtx *APIContext) TheNodeShouldNotBe(df format.DataFormat, exprTemplate string, goType string) error
//	func (apiCtx *APIContext) TheNodeShouldBe(df format.DataFormat, exprTemplate string, goType string) error
//	func (apiCtx *APIContext) TheNodeShouldMatchRegExp(dataFormat format.DataFormat, exprTemplate, regExpTemplate string) error
//	func (apiCtx *APIContext) TheResponseShouldHaveNodes(dataFormat format.DataFormat, expressionsTemplates string) error
//	func (apiCtx *APIContext) TheNodeShouldBeSliceOfLength(dataFormat format.DataFormat, exprTemplate string, length int) error
//	func (apiCtx *APIContext) TheNodeShouldBeOfValue(dataFormat format.DataFormat, exprTemplate, dataType, dataValue string) error
//	func (apiCtx *APIContext) TheResponseShouldHaveHeader(name string) error
//	func (apiCtx *APIContext) TheResponseShouldHaveHeaderOfValue(name, value string) error
//  func (apiCtx *APIContext) IValidateLastResponseBodyWithSchemaReference(referenceTemplate string) error
//	func (apiCtx *APIContext) IValidateLastResponseBodyWithSchemaString(schemaTemplate string) error
//	func (apiCtx *APIContext) IValidateNodeWithSchemaString(dataFormat format.DataFormat, exprTemplate, schemaTemplate string) error
//	func (apiCtx *APIContext) IValidateNodeWithSchemaReference(dataFormat format.DataFormat, exprTemplate, referenceTemplate string) error
//	func (apiCtx *APIContext) TimeBetweenLastHTTPRequestResponseShouldBeLessThanOrEqualTo(timeInterval time.Duration) error
//
// * Preserving JSON nodes:
//
//	func (apiCtx *APIContext) ISaveFromTheLastResponseNodeAs(dataFormat format.DataFormat, exprTemplate, cacheKey string) error
//  func (apiCtx *APIContext) ISaveAs(valueTemplate, cacheKey string) error
//
// * Flow control:
//
//	func (apiCtx *APIContext) IWait(timeInterval time.Duration) error
//
// * Debugging:
//
//	func (apiCtx *APIContext) IPrintLastResponseBody() error
//	func (apiCtx *APIContext) IStartDebugMode() error
//	func (apiCtx *APIContext) IStopDebugMode() error
package gdutils
