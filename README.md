[![gdutils](https://github.com/pawelWritesCode/gdutils/workflows/gdutils/badge.svg)](https://github.com/pawelWritesCode/gdutils/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/pawelWritesCode/gdutils.svg)](https://pkg.go.dev/github.com/pawelWritesCode/gdutils) ![Coverage](https://img.shields.io/badge/Coverage-52.1%25-brightgreen)

# GDUTILS

## Simple library with methods useful for e2e testing of HTTP(s) API.

Library is suitable for steps in godog framework.

### Downloading

`go get github.com/pawelWritesCode/gdutils`

### Related project:

Skeleton that allows to write e2e HTTP API tests using *godog & gdutils* almost instantly with minimal configuration.
https://github.com/pawelWritesCode/godog-example-setup

### Available methods:

| NAME                             |      DESCRIPTION                                       |
|----------------------------------|:------------------------------------------------------:|
| | |
|  **Sending HTTP(s) requests:**                                                                                  |
| | |
| ISendRequestToWithBodyAndHeaders |  Sends HTTP(s) request with provided body and headers. |
| IPrepareNewRequestToAndSaveItAs  |  Prepare HTTP(s) request |
| ISetFollowingHeadersForPreparedRequest  |  Sets provided headers for previously prepared request |
| ISetFollowingFormForPreparedRequest  |  Sets provided form for previously prepared request |
| ISetFollowingCookiesForPreparedRequest  |  Sets provided cookies for previously prepared request |
| ISetFollowingBodyForPreparedRequest  |  Sets body for previously prepared request |
| ISendRequest  |  Sends previously prepared HTTP(s) request |
| | |
| **Random data generation:** |
| | |
| IGenerateARandomIntInTheRangeToAndSaveItAs | Generates random integer from provided range and save it under provided cache key |
| IGenerateARandomFloatInTheRangeToAndSaveItAs | Generates random float from provided range and save it under provided cache key |
| IGenerateARandomRunesInTheRangeToAndSaveItAs | Creates generator for random strings from provided charset in provided range |
| IGenerateARandomSentenceInTheRangeFromToWordsAndSaveItAs | Creates generator for random sentence from provided charset in provided range |
| IGetTimeAndTravelByAndSaveItAs | Accepts time object and move in time by given time interval |
| IGenerateCurrentTimeAndTravelByAndSaveItAs | Creates current time object and move in time by given time interval |
| | |
| **Preserving data:** |
| | |
| ISaveFromTheLastResponseNodeAs | Saves from last response body JSON node under given cacheKey key |
| ISaveAs | Saves into cache arbitrary passed value |
| | |
| **Debugging:** |
| | |
| IPrintLastResponseBody | Prints last response from request |
| IStartDebugMode | Starts debugging mode |
| IStopDebugMode | Stops debugging mode |
| | |
| **Flow control:** |
| | |
| IWait | Stops test execution for provided amount of time |
| | |
| **Assertions:** |
| | |
| TheResponseShouldHaveHeader | Checks whether last HTTP(s) response has given header |
| TheResponseShouldHaveHeaderOfValue | Checks whether last HTTP(s) response has given header with provided value |
| TheResponseStatusCodeShouldBe | Checks last HTTP(s) response status code |
| TheResponseBodyShouldHaveType | Checks whether last HTTP(s) response body has given data format |
| TheResponseShouldHaveNode | Checks whether last response body contains given key |
| TheNodeShouldBeOfValue | Compares json node value from expression to expected by user |
| TheNodeShouldBe | Checks whether node from last HTTP(s) response body is of provided type |
| TheNodeShouldNotBe | Checks whether node from last response body is not of provided type |
| TheResponseShouldHaveNodes | Checks whether last HTTP(s) response body JSON has given nodes |
| TheNodeShouldMatchRegExp | Checks whether last HTTP(s) response body JSON node matches regExp |
| TheNodeShouldBeSliceOfLength | checks whether given key is slice and has given length |
| IValidateLastResponseBodyWithSchemaReference | Validates last HTTP(s) response body against provided in reference JSON schema |
| IValidateLastResponseBodyWithSchemaString | Validates last HTTP(s) response body against provided JSON schema |
| IValidateJSONNodeWithSchemaString | Validates last HTTP(s) response body JSON node against provided JSON schema |
| IValidateJSONNodeWithSchemaReference | Validates last HTTP(s) response body JSON node against provided in reference JSON schema |
| TimeBetweenLastHTTPRequestResponseShouldBeLessThanOrEqualTo | Asserts that last HTTP(s) request-response time is <= than expected |
| TheResponseShouldHaveCookie | Checks whether last HTTP(s) response has given cookie |
| TheResponseShouldHaveCookieOfValue | Checks whether last HTTP(s) response has given cookie of given value |
