// Package gdutils provides State struct with methods that may be used for behavioral testing of HTTP API.
//
// Main package struct is State that represents one testing scenario state. It can be initialized by 2 ways:
//
// First, returns *State with default http.Client, DefaultCache, default JSONschemaValidator and provided debug mode:
//	func NewDefaultState(isDebug bool, jsonSchemaDir string) *State
//
// Second, returns *State with provided http.Client, Cache, JSONschemaValidator and debug mode:
//	func NewState(httpClient *http.Client, cache cache.Cache, jsonSchemaValidator validator.SchemaValidator, isDebug bool) *State
//
// Testing HTTP API usually consist the following aspects:
//
// * Data generation:
//
//	func (s *State) IGenerateARandomIntInTheRangeToAndSaveItAs(from, to int, cacheKey string) error
//	func (s *State) IGenerateARandomFloatInTheRangeToAndSaveItAs(from, to int, cacheKey string) error
//  func (s *State) IGenerateARandomStringInTheRangeToAndSaveItAs(charset string) func(from, to int, cacheKey string) error
//
// * Sending HTTP(s) requests:
//
//	func (s *State) ISendRequestToWithBodyAndHeaders(method, urlTemplate string, bodyTemplate *godog.DocString) error
//
// or
//
//	func (s *State) IPrepareNewRequestToAndSaveItAs(method, urlTemplate, cacheKey string) error
//	func (s *State) ISetFollowingHeadersForPreparedRequest(cacheKey string, headersTemplate *godog.DocString) error
//	func (s *State) ISetFollowingBodyForPreparedRequest(cacheKey string, bodyTemplate *godog.DocString) error
//	func (s *State) ISendRequest(cacheKey string) error
//
// * Assertions:
//
//	func (s *State) TheResponseStatusCodeShouldBe(code int) error
//	func (s *State) TheResponseBodyShouldHaveType(dataType string) error
//	func (s *State) TheJSONResponseShouldHaveNode(expr string) error
//	func (s *State) TheJSONNodeShouldNotBe(expr string, goType string) error
//	func (s *State) TheJSONNodeShouldBe(expr string, goType string) error
//	func (s *State) TheJSONResponseShouldHaveNodes(nodeExprs string) error
//	func (s *State) TheJSONNodeShouldBeSliceOfLength(expr string, length int) error
//	func (s *State) TheJSONNodeShouldBeOfValue(expr, dataType, dataValue string) error
//	func (s *State) TheResponseShouldHaveHeader(name string) error
//	func (s *State) TheResponseShouldHaveHeaderOfValue(name, value string) error
//  func (s *State) IValidateLastResponseBodyWithSchema(source string) error
//
// * Preserving JSON nodes:
//
//	func (s *State) ISaveFromTheLastResponseJSONNodeAs(expr, cacheKey string) error
//
// * Temporary stopping scenario execution:
//
//	func (s *State) IWait(timeInterval string) error
//
// * Debugging:
//
//	func (s *State) IPrintLastResponseBody() error
//	func (s *State) IStartDebugMode() error
//	func (s *State) IStopDebugMode() error
//
// State has also some useful utility methods like:
//
//	func (s *State) GetLastResponse() (*http.Response, error)
//	func (s *State) GetLastResponseBody() []byte
//	func (s *State) GetLastResponseHeaders() http.Header
package gdutils
