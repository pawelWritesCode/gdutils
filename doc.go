// Package gdutils provides APIContext struct with methods that may be used for behavioral testing of HTTP API.
//
// APIContext may be initialized by two ways:
//
// First, returns *APIContext with default services:
//
//	func NewDefaultAPIContext(isDebug bool, jsonSchemaDir string) *APIContext
//
// Second, more customisable returns *APIContext with provided services:
//
//	func NewAPIContext(cli *http.Client, c cache.Cache, jv SchemaValidators, p PathFinders, f Formatters, t TypeMappers, d debugger.Debugger) *APIContext
//
// No matter which way you choose, you can inject your custom services afterwards with one of available setters:
//
//	func (apiCtx *APIContext) SetDebugger(d debuggable)
//	func (apiCtx *APIContext) SetCache(c cacheable)
//	func (apiCtx *APIContext) SetRequestDoer(r requestDoer)
//	func (apiCtx *APIContext) SetTemplateEngine(t templateEngine)
//	func (apiCtx *APIContext) SetSchemaStringValidator(j validator.SchemaValidator)
//	func (apiCtx *APIContext) SetSchemaReferenceValidator(j validator.SchemaValidator)
//	func (apiCtx *APIContext) SetJSONPathFinder(r pathFinder)
//	func (apiCtx *APIContext) SetJSONFormatter(jf serializable)
//	func (apiCtx *APIContext) SetXMLPathFinder(r pathFinder)
//	func (apiCtx *APIContext) SetXMLFormatter(xf serializable)
//	func (apiCtx *APIContext) SetYAMLPathFinder(r pathFinder)
//	func (apiCtx *APIContext) SetYAMLFormatter(yd serializable)
//	func (apiCtx *APIContext) SetHTMLPathFinder(r pathFinder)
//	func (apiCtx *APIContext) SetJSONTypeMapper(c typeMapper)
//	func (apiCtx *APIContext) SetYAMLTypeMapper(c typeMapper)
//	func (apiCtx *APIContext) SetGoTypeMapper(c typeMapper)
//
// Those services will be used in utility methods and can be accessed directly if needed (to use in any custom methods).
// For example, if you want to use your own debugger - because default one is not suitable for you, create your own struct,
// implement debugger.Debugger interface on it, and then inject it with "func (apiCtx *APIContext) SetDebugger(d debugger.Debugger)" method.
//
// Testing HTTP API usually consist of the following aspects:
//
// * Data generation:
//
//	func (apiCtx *APIContext) GenerateRandomInt(from, to int, cacheKey string) error
//	func (apiCtx *APIContext) GenerateFloat64(from, to float64, cacheKey string) error
//	func (apiCtx *APIContext) GeneratorRandomRunes(charset string) func(from, to int, cacheKey string) error
//	func (apiCtx *APIContext) GeneratorRandomSentence(charset string, wordMinLength, wordMaxLength int) func(from, to int, cacheKey string) error
//	func (apiCtx *APIContext) GetTimeAndTravel(t time.Time, timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error
//	func (apiCtx *APIContext) GenerateTimeAndTravel(timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error
//
// * Sending HTTP(s) requests:
//
//	func (apiCtx *APIContext) RequestSendWithBodyAndHeaders(method, urlTemplate string, bodyTemplate string) error
//
// or
//
//	func (apiCtx *APIContext) RequestPrepare(method, urlTemplate, cacheKey string) error
//	func (apiCtx *APIContext) RequestSetHeaders(cacheKey string, headersTemplate string) error
//	func (apiCtx *APIContext) RequestSetForm(cacheKey, formTemplate string) error
//	func (apiCtx *APIContext) RequestSetCookies(cacheKey, cookiesTemplate string) error
//	func (apiCtx *APIContext) RequestSetBody(cacheKey string, bodyTemplate string) error
//	func (apiCtx *APIContext) RequestSend(cacheKey string) error
//
// * Assertions:
//
//	func (apiCtx *APIContext) AssertStatusCodeIs(code int) error
//	func (apiCtx *APIContext) AssertStatusCodeIsNot(code int) error
//	func (apiCtx *APIContext) AssertResponseFormatIs(dataFormat format.DataFormat) error
//	func (apiCtx *APIContext) AssertResponseFormatIsNot(dataFormat format.DataFormat) error
//	func (apiCtx *APIContext) AssertResponseCookieExists(name string) error
//	func (apiCtx *APIContext) AssertResponseCookieNotExists(name string) error
//	func (apiCtx *APIContext) AssertResponseCookieValueIs(name, valueTemplate string) error
//	func (apiCtx *APIContext) AssertResponseCookieValueNotMatchesRegExp(name, regExpTemplate string) error
//	func (apiCtx *APIContext) AssertNodesExist(dataFormat format.DataFormat, expressionsTemplates string) error
//	func (apiCtx *APIContext) AssertNodeExists(dataFormat format.DataFormat, exprTemplate string) error
//	func (apiCtx *APIContext) AssertNodeNotExists(dataFormat format.DataFormat, exprTemplate string) error
//	func (apiCtx *APIContext) AssertNodeIsType(dataFormat format.DataFormat, exprTemplate string, inType types.DataType) error
//	func (apiCtx *APIContext) AssertNodeIsNotType(dataFormat format.DataFormat, exprTemplate string, inType types.DataType) error
//	func (apiCtx *APIContext) AssertNodeMatchesRegExp(dataFormat format.DataFormat, exprTemplate, regExpTemplate string) error
//	func (apiCtx *APIContext) AssertNodeNotMatchesRegExp(dataFormat format.DataFormat, exprTemplate, regExpTemplate string) error
//	func (apiCtx *APIContext) AssertNodeIsTypeAndValue(dataFormat format.DataFormat, exprTemplate string, dataType types.DataType, dataValue string) error
//	func (apiCtx *APIContext) AssertNodeIsTypeAndHasOneOfValues(dataFormat format.DataFormat, exprTemplate string, dataType types.DataType, valuesTemplates string) error
//	func (apiCtx *APIContext) AsserNodeNotContainsSubString(dataFormat format.DataFormat, exprTemplate string, subTemplate string) error
//	func (apiCtx *APIContext) AsserNodeContainsSubString(dataFormat format.DataFormat, exprTemplate string, subTemplate string) error
//	func (apiCtx *APIContext) AssertNodeSliceLengthIs(dataFormat format.DataFormat, exprTemplate string, length int) error
//	func (apiCtx *APIContext) AssertNodeSliceLengthIsNot(dataFormat format.DataFormat, exprTemplate string, length int) error
//	func (apiCtx *APIContext) AssertResponseHeaderExists(name string) error
//	func (apiCtx *APIContext) AssertResponseHeaderNotExists(name string) error
//	func (apiCtx *APIContext) AssertResponseHeaderValueIs(name, value string) error
//	func (apiCtx *APIContext) AssertResponseMatchesSchemaByReference(referenceTemplate string) error
//	func (apiCtx *APIContext) AssertResponseMatchesSchemaByString(schemaTemplate string) error
//	func (apiCtx *APIContext) AssertNodeMatchesSchemaByString(dataFormat format.DataFormat, exprTemplate, schemaTemplate string) error
//	func (apiCtx *APIContext) AssertNodeMatchesSchemaByReference(dataFormat format.DataFormat, exprTemplate, referenceTemplate string) error
//	func (apiCtx *APIContext) AssertTimeBetweenRequestAndResponseIs(timeInterval time.Duration) error
//
// * Preserving nodes:
//
//	func (apiCtx *APIContext) SaveNode(dataFormat format.DataFormat, exprTemplate, cacheKey string) error
//	func (apiCtx *APIContext) SaveHeader(name, cacheKey string) error
//	func (apiCtx *APIContext) Save(valueTemplate, cacheKey string) error
//
// * Flow control:
//
//	func (apiCtx *APIContext) Wait(timeInterval time.Duration) error
//
// * Debugging:
//
//	func (apiCtx *APIContext) DebugPrintResponseBody() error
//	func (apiCtx *APIContext) DebugStart() error
//	func (apiCtx *APIContext) DebugStop() error
//
// Here is example lib usage to test endpoint returning list of ducks gifs:
//
//		ac := gdutils.NewDefaultAPIContext(false, "")
//
//		if err := ac.RequestPrepare("GET", "https://random-d.uk/api/v2/list", "DUCK_GIFS_LIST"); err != nil {
//			return err
//		}
//
//		if err := ac.RequestSend("DUCK_GIFS_LIST"); err != nil {
//	   		return err
//		}
//
//		if err := ac.AssertStatusCodeIs(200); err != nil {
//	   		return err
//		}
//
//		if err := ac.AssertResponseFormatIs(format.JSON); err != nil {
//	   		return err
//		}
//
//		if err := ac.AssertNodeIsType(format.JSON, "$.gifs", types.Array); err != nil {
//	   		return err
//		}
package gdutils
