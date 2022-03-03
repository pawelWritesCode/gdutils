// Package gdutils provides State struct with methods that may be used for behavioral testing of HTTP API.
//
// State may be initialized by two ways:
//
// First, returns *State with default services:
//	func NewDefaultState(isDebug bool, jsonSchemaDir string) *State
//
// Second, more customisable returns *State with provided services:
//	func NewState(cli *http.Client, c cache.Cache, jv SchemaValidators, p PathFinders, d Formatters, isDebug bool) *State
//
// No matter which way you choose, you can inject your custom services afterwards with one of available setters:
//	func (s *State) SetDebugger(d debugger.Debugger)
//	func (s *State) SetCache(c cache.Cache)
//	func (s *State) SetRequestDoer(r httpctx.RequestDoer)
//	func (s *State) SetTemplateEngine(t template.Engine)
//	func (s *State) SetSchemaStringValidator(j validator.SchemaValidator)
//	func (s *State) SetSchemaReferenceValidator(j validator.SchemaValidator)
//	func (s *State) SetJSONPathFinder(r pathfinder.PathFinder)
//	func (s *State) SetJSONFormatter(jf formatter.Formatter)
//	func (s *State) SetYAMLPathFinder(r pathfinder.PathFinder)
//	func (s *State) SetYAMLFormatter(yd formatter.Formatter)
//
// Those services will be used in utility methods.
// For example, if you want to use your own debugger, create your own struct, implement debugger.Debugger interface on it,
// and then inject it with "func (s *State) SetDebugger(d debugger.Debugger)" method.
//
// Testing HTTP API usually consist the following aspects:
//
// * Data generation:
//
//	func (s *State) IGenerateARandomIntInTheRangeToAndSaveItAs(from, to int, cacheKey string) error
//	func (s *State) IGenerateARandomFloatInTheRangeToAndSaveItAs(from, to int, cacheKey string) error
//  func (s *State) IGenerateARandomRunesInTheRangeToAndSaveItAs(charset string) func(from, to int, cacheKey string) error
//	func (s *State) IGenerateARandomSentenceInTheRangeFromToWordsAndSaveItAs(charset string, wordMinLength, wordMaxLength int) func(from, to int, cacheKey string) error
//	func (s *State) IGetTimeAndTravelByAndSaveItAs(t time.Time, timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error
//	func (s *State) IGenerateCurrentTimeAndTravelByAndSaveItAs(timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error
//
// * Sending HTTP(s) requests:
//
//	func (s *State) ISendRequestToWithFormatBodyAndHeaders(method, urlTemplate string, dataFormat format.DataFormat, bodyTemplate string) error
//
// or
//
//	func (s *State) IPrepareNewRequestToAndSaveItAs(method, urlTemplate, cacheKey string) error
//	func (s *State) ISetFollowingHeadersForPreparedRequest(cacheKey string, headersTemplate string) error
//	func (s *State) ISetFollowingCookiesForPreparedRequest(cacheKey, cookiesTemplate string) error
//	func (s *State) ISetFollowingBodyForPreparedRequest(cacheKey string, bodyTemplate string) error
//	func (s *State) ISendRequest(cacheKey string) error
//
// * Assertions:
//
//	func (s *State) TheResponseStatusCodeShouldBe(code int) error
//	func (s *State) TheResponseBodyShouldHaveFormat(dataFormat format.DataFormat) error
//	func (s *State) TheResponseShouldHaveCookie(name string) error
//	func (s *State) TheResponseShouldHaveCookieOfValue(name, valueTemplate string) error
//	func (s *State) TheResponseShouldHaveNode(dataFormat format.DataFormat, expr string) error
//	func (s *State) TheNodeShouldNotBe(df format.DataFormat, expr string, goType string) error
//	func (s *State) TheNodeShouldBe(df format.DataFormat, expr string, goType string) error
//	func (s *State) TheNodeShouldMatchRegExp(dataFormat format.DataFormat, expr, regExpTemplate string) error
//	func (s *State) TheResponseShouldHaveNodes(dataFormat format.DataFormat, expressions string) error
//	func (s *State) TheNodeShouldBeSliceOfLength(dataFormat format.DataFormat, expr string, length int) error
//	func (s *State) TheNodeShouldBeOfValue(dataFormat format.DataFormat, expr, dataType, dataValue string) error
//	func (s *State) TheResponseShouldHaveHeader(name string) error
//	func (s *State) TheResponseShouldHaveHeaderOfValue(name, value string) error
//  func (s *State) IValidateLastResponseBodyWithSchemaReference(reference string) error
//	func (s *State) IValidateLastResponseBodyWithSchemaString(schema string) error
//	func (s *State) IValidateNodeWithSchemaString(dataFormat format.DataFormat, expr, schema string) error
//	func (s *State) IValidateNodeWithSchemaReference(dataFormat format.DataFormat, expr, reference string) error
//	func (s *State) TimeBetweenLastHTTPRequestResponseShouldBeLessThanOrEqualTo(timeInterval time.Duration) error
//
// * Preserving JSON nodes:
//
//	func (s *State) ISaveFromTheLastResponseNodeAs(dataFormat format.DataFormat, expr, cacheKey string) error
//  func (s *State) ISaveAs(value, cacheKey string) error
//
// * Temporary stopping scenario execution:
//
//	func (s *State) IWait(timeInterval time.Duration) error
//
// * Debugging:
//
//	func (s *State) IPrintLastResponseBody() error
//	func (s *State) IStartDebugMode() error
//	func (s *State) IStopDebugMode() error
package gdutils
