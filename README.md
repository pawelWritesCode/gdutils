[![gdutils](https://github.com/pawelWritesCode/gdutils/workflows/gdutils/badge.svg)](https://github.com/pawelWritesCode/gdutils/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/pawelWritesCode/gdutils.svg)](https://pkg.go.dev/github.com/pawelWritesCode/gdutils) ![Coverage](https://img.shields.io/badge/Coverage-63.7%25-brightgreen)

# GDUTILS

## Simple library with methods useful for e2e testing of HTTP(s) API using JSON/YAML/XML.

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
| RequestSendWithBodyAndHeaders |  Sends HTTP(s) request with provided body and headers. |
| RequestPrepare  |  Prepare HTTP(s) request |
| RequestSetHeaders  |  Sets provided headers for previously prepared request |
| RequestSetForm  |  Sets provided form for previously prepared request |
| RequestSetCookies  |  Sets provided cookies for previously prepared request |
| RequestSetBody  |  Sets body for previously prepared request |
| RequestSend  |  Sends previously prepared HTTP(s) request |
| | |
| **Random data generation:** |
| | |
| GenerateRandomInt | Generates random integer from provided range and save it under provided cache key |
| IGenerateARandomFloatInTheRangeToAndSaveItAs | Generates random float from provided range and save it under provided cache key |
| GeneratorRandomRunes | Creates generator for random strings from provided charset in provided range |
| GeneratorRandomSentence | Creates generator for random sentence from provided charset in provided range |
| GetTimeAndTravel | Accepts time object and move in time by given time interval |
| GenerateTimeAndTravel | Creates current time object and move in time by given time interval |
| | |
| **Preserving data:** |
| | |
| SaveNode | Saves from last response body JSON node under given cacheKey key |
| SaveHeader | Saves into cache given header value |
| Save | Saves into cache arbitrary passed value |
| | |
| **Debugging:** |
| | |
| DebugPrintResponseBody | Prints last response from request |
| DebugStart | Starts debugging mode |
| DebugStop | Stops debugging mode |
| | |
| **Flow control:** |
| | |
| Wait | Stops test execution for provided amount of time |
| | |
| **Assertions:** |
| | |
| AssertResponseHeaderExists | Checks whether last HTTP(s) response has given header |
| AssertResponseHeaderNotExists | Checks whether last HTTP(s) response doesn't have given header |
| AssertResponseHeaderValueIs | Checks whether last HTTP(s) response has given header with provided value |
| AssertStatusCodeIs | Checks last HTTP(s) response status code |
| AssertStatusCodeIsNot | Checks if last HTTP(s) response status code is not of provided value |
| AssertResponseFormatIs | Checks whether last HTTP(s) response body has given data format |
| AssertResponseFormatIsNot | Checks whether last HTTP(s) response body doesn't have given data format |
| AssertNodeExists | Checks whether last response body contains given key |
| AssertNodeNotExists | Checks whether last response body doesn't contain given key |
| AssertNodesExist | Checks whether last HTTP(s) response body JSON has given nodes |
| AssertNodeIsTypeAndValue | Compares json node value from expression to expected by user |
| AssertNodeIsType | Checks whether node from last HTTP(s) response body is of provided type |
| AssertNodeIsNotType | Checks whether node from last HTTP(s) response body is not of provided type |
| AssertNodeIsNotType | Checks whether node from last response body is not of provided type |
| AssertNodeMatchesRegExp | Checks whether last HTTP(s) response body JSON node matches regExp |
| AssertNodeNotMatchesRegExp | Checks whether last HTTP(s) response body JSON node doesn't match regExp |
| AssertNodeSliceLengthIs | checks whether given key is slice and has given length |
| AssertNodeSliceLengthIsNot | checks whether given key is slice and doesn't have given length |
| AssertResponseMatchesSchemaByReference | Validates last HTTP(s) response body against provided in reference JSON schema |
| AssertResponseMatchesSchemaByString | Validates last HTTP(s) response body against provided JSON schema |
| AssertNodeMatchesSchemaByString | Validates last HTTP(s) response body JSON node against provided JSON schema |
| AssertNodeMatchesSchemaByReference | Validates last HTTP(s) response body JSON node against provided in reference JSON schema |
| AssertTimeBetweenRequestAndResponseIs | Asserts that last HTTP(s) request-response time is <= than expected |
| AssertResponseCookieExists | Checks whether last HTTP(s) response has given cookie |
| AssertResponseCookieNotExists | Checks whether last HTTP(s) response doesn't have given cookie |
| AssertResponseCookieValueIs | Checks whether last HTTP(s) response has given cookie of given value |
