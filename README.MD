# GDUTILS

## Simple library with godog steps


#### Example InitializeScenario func
```
func InitializeScenario(ctx *godog.ScenarioContext) {
    err := godotenv.Load()
    if err != nil {
		log.Fatal(err.Error())
	}

	af := &gdutils.ApiFeature{}

	ctx.BeforeScenario(func(*godog.Scenario) {
		af.ResetApiFeature()
		af.SetBaseUrl("http://" + os.Getenv("ADDR") + ":" + os.Getenv("PORT"))
	})

	ctx.Step(`^I generate a random string "([^"]*)"$`, af.IGenerateARandomString)
	ctx.Step(`^I generate a random int "([^"]*)"$`, af.IGenerateARandomInt)
	ctx.Step(`^I send "([^"]*)" request to "([^"]*)" with data:$`, af.ISendRequestToWithData)
	ctx.Step(`^I send "([^"]*)" request with token "([^"]*)" to "([^"]*)" with data:$`, af.ISendRequestWithTokenToWithData)
	ctx.Step(`^I send "([^"]*)" request with token "([^"]*)" to "([^"]*)"$`, af.ISendRequestWithTokenTo)
	ctx.Step(`^the response status code should be (\d+)$`, af.TheResponseStatusCodeShouldBe)
	ctx.Step(`^the response should be in JSON$`, af.TheResponseShouldBeInJSON)
	ctx.Step(`^I save from the last response JSON node "([^"]*)" as "([^"]*)"$`, af.ISaveFromTheLastResponseJSONNodeAs)
	ctx.Step(`^the JSON response should have key "([^"]*)"$`, af.TheJSONResponseShouldHaveKey)
	ctx.Step(`^I create data:$`, af.ICreateData)
	ctx.Step(`^I send "([^"]*)" request to "([^"]*)" with body and headers:$`, af.ISendRequestToWithBodyAndHeaders)
	ctx.Step(`^the JSON node "([^"]*)" should be "([^"]*)" of value "([^"]*)"$`, af.TheJSONNodeShouldBeOfValue)
	ctx.Step(`^I print last response$`, af.IPrintLastResponse)
}
```