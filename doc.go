// Package gdutils provides State object with utility methods that may be used for behavioral testing of HTTP API.
//
// Main package struct is State that represents one testing scenario state. It contains of
//
//* DefaultCache with arbitrary values - State.Cache,
//* info about last HTTP response,
//	func (s *State) GetLastResponseBody() []byte
//	func (s *State) GetLastResponseHeaders() http.Header
//* info whether scenario is in debug mode - State.IsDebug
//
// Testing of HTTP API usually has the following aspects:
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
// * Stopping scenario execution:
//
//	func (s *State) IWait(timeInterval string) error
//
// * Debugging:
//
//	func (s *State) IPrintLastResponseBody() error
package gdutils
