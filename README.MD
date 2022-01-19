[![gdutils](https://github.com/pawelWritesCode/gdutils/workflows/gdutils/badge.svg)](https://github.com/pawelWritesCode/gdutils/actions)

# GDUTILS

## Simple library with methods useful for e2e testing of HTTP(s) API.

Library is suitable for steps in godog framework.

### Downloading

`go get github.com/pawelWritesCode/gdutils`

### Docs:

https://pkg.go.dev/github.com/pawelWritesCode/gdutils

### Details:

| NAME                             |      DESCRIPTION                                       |
|----------------------------------|:------------------------------------------------------:|
| | |
|  **Sending HTTP(s) requests:**                                                                                  |
| | |
| ISendRequestToWithBodyAndHeaders |  Sends HTTP(s) request with provided body and headers. |
| IPrepareNewRequestToAndSaveItAs  |  Prepare HTTP(s) request |
| ISetFollowingHeadersForPreparedRequest  |  Sets provided headers for previously prepared request |
| ISetFollowingBodyForPreparedRequest  |  Sets body for previously prepared request |
| ISendRequest  |  Sends previously prepared HTTP(s) request |
| | |
| **Random data generation:** |
| | |
| IGenerateARandomIntInTheRangeToAndSaveItAs | Generates random integer from provided range and save it under provided cache key |
| IGenerateARandomFloatInTheRangeToAndSaveItAs | Generates random float from provided range and save it under provided cache key |
| IGenerateARandomStringInTheRangeToAndSaveItAs | Creates generator for random strings from provided charset in provided range |
| | |
| **Preserving data:** |
| | |
| ISaveFromTheLastResponseJSONNodeAs | Saves from last response body JSON node under given cacheKey key |
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
| TheJSONResponseShouldHaveNode | Checks whether last response body contains given key |
| TheJSONNodeShouldBeOfValue | Compares json node value from expression to expected by user |
| TheJSONNodeShouldBe | Checks whether JSON node from last HTTP(s) response body is of provided type |
| TheJSONNodeShouldNotBe | Checks whether JSON node from last response body is not of provided type |
| TheJSONResponseShouldHaveNodes | Checks whether last HTTP(s) response body JSON has given nodes |
| TheJSONNodeShouldBeSliceOfLength | checks whether given key is slice and has given length |
| IValidateLastResponseBodyWithSchema | Validates last HTTP(s) response body against provided JSON schema |
