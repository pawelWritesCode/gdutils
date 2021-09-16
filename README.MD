# GDUTILS

## Simple library with godog steps


#### Example InitializeScenario func
```
//InitializeScenario holds definitions of steps and holds logic for new scenario initialization
func InitializeScenario(ctx *godog.ScenarioContext) {
	var debug bool
	err := godotenv.Load()
	checkErr(err)

	debugString := os.Getenv("DEBUG")
	if strings.ToLower(debugString) == "true" {
		debug = true
	}

	s := &Scenario{}

	//BeforeScenario is method that is run before each scenario
	//its main purpose is to reset state of previously initialized Scenario struct
	ctx.BeforeScenario(func(*godog.Scenario) {
		s.ResetScenario(debug)
	})

	//Generation of random data
	//generated data is available via template value {{.SAVED_VALUE}} in some next steps
	ctx.Step(`^i generate a random string of length "([^"]*)" without unicode characters and save it as "([^"]*)"$`, s.IGenerateARandomStringOfLengthWithoutUnicodeCharactersAndSaveItAs)
	ctx.Step(`^i generate a random string of length "([^"]*)" with unicode characters and save it as "([^"]*)"$`, s.IGenerateARandomStringOfLengthWithUnicodeCharactersAndSaveItAs)
	ctx.Step(`^i generate a random float in the range "([^"]*)" to "([^"]*)" and save it as "([^"]*)"$`, s.IGenerateARandomFloatInTheRangeToAndSaveItAs)
	ctx.Step(`^i generate a random int in the range "([^"]*)" to "([^"]*)" and save it as "([^"]*)"$`, s.IGenerateARandomIntInTheRangeToAndSaveItAs)

	//Sending HTTP requests
	ctx.Step(`^i send "(GET|POST|PUT|PATCH|DELETE)" request to "([^"]*)" with body and headers:$`, s.ISendRequestToWithBodyAndHeaders)

	//Last response body assertions
	ctx.Step(`^the response should have header "([^"]*)"$`, s.TheResponseShouldHaveHeader)
	ctx.Step(`^the response should have header "([^"]*)" of value "([^"]*)"$`, s.TheResponseShouldHaveHeaderOfValue)
	ctx.Step(`^the response status code should be (\d+)$`, s.TheResponseStatusCodeShouldBe)
	ctx.Step(`^the JSON response should have key "([^"]*)"$`, s.TheJSONResponseShouldHaveKey)
	ctx.Step(`^the JSON node "([^"]*)" should be "([^"]*)" of value "([^"]*)"$`, s.TheJSONNodeShouldBeOfValue)
	ctx.Step(`^the JSON node "([^"]*)" should be slice of length "([^"]*)"$`, s.TheJSONNodeShouldBeSliceOfLength)
	ctx.Step(`^the JSON node "([^"]*)" should be "(nil|string|int|float|bool|map|slice)"$`, s.TheJSONNodeShouldBe)
	ctx.Step(`^the JSON node "([^"]*)" should not be "(nil|string|int|float|bool|map|slice)"$`, s.TheJSONNodeShouldNotBe)
	ctx.Step(`^the JSON response should have keys "([^"]*)"$`, s.TheJSONResponseShouldHaveKeys)

	//Response body type assertions
	ctx.Step(`^the response should be in "(JSON|XML)"$`, s.TheResponseShouldBeIn)

	//Saving JSON node to user defined variable, available as template value in next steps
	ctx.Step(`^i save from the last response JSON node "([^"]*)" as "([^"]*)"$`, s.ISaveFromTheLastResponseJSONNodeAs)

	//Printing last response body to console
	ctx.Step(`^i print last response body$`, s.IPrintLastResponse)

	//Blocking scenario execution for some time. Available method values should compatible with time.ParseDuration method
	ctx.Step(`^i wait "([^"]*)"`, s.IWait)
}
```