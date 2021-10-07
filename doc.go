// Package gdutils provides State object with utility methods that may be used for behavioral testing of HTTP API.
//
// Main package struct is State that represents one testing scenario state. It can be initialized by 2 ways.
//
// First, returns *State with default http.Client, DefaultCache and provided debug mode:
//	func NewDefaultState(isDebug bool) *State
//
// Second, returns *State with provided http.Client, Cache and debug mode:
//	func NewState(HttpClient HttpClient, cache Cache, isDebug bool) *State
//
// Struct State contains of
//
//* cache with arbitrary values - State.Cache.
//
//* info whether scenario is in debug mode - State.IsDebug.
//
//	Useful State methods are
//
//	func (s *State) GetLastResponse() (*http.Response, error)
//	func (s *State) GetLastResponseBody() []byte
//	func (s *State) GetLastResponseHeaders() http.Header
//
// Testing HTTP API usually consist the following aspects:
//
// * Data generation:
//
//	func (s *State) IGenerateARandomIntInTheRangeToAndSaveItAs(from, to int, cacheKey string) error
//	func (s *State) IGenerateARandomFloatInTheRangeToAndSaveItAs(from, to int, cacheKey string) error
//	func (s *State) IGenerateARandomStringOfLengthWithoutUnicodeCharactersAndSaveItAs(strLength int, cacheKey string) error
//	func (s *State) IGenerateARandomStringOfLengthWithUnicodeCharactersAndSaveItAs(strLength int, cacheKey string) error
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
package gdutils
