package gdutils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/format"
	"github.com/pawelWritesCode/gdutils/pkg/httpcache"
	"github.com/pawelWritesCode/gdutils/pkg/mathutils"
	"github.com/pawelWritesCode/gdutils/pkg/stringutils"
	"github.com/pawelWritesCode/gdutils/pkg/template"
	"github.com/pawelWritesCode/gdutils/pkg/timeutils"
	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

type mockedHTTPContext struct {
	mock.Mock
}

type mockedJSONValidator struct {
	mock.Mock
}

type mockedTemplateEngine struct {
	mock.Mock
}

type mockedFormatter struct {
	mock.Mock
}

func (m *mockedFormatter) Deserialize(data []byte, v any) error {
	args := m.Called(data, v)

	return args.Error(0)
}

func (m *mockedFormatter) Serialize(v any) ([]byte, error) {
	args := m.Called(v)

	return args.Get(0).([]byte), args.Error(1)
}

type mockedJsonPathFinder struct {
	mock.Mock
}

func (m *mockedTemplateEngine) Replace(templateValue string, storage map[string]any) (string, error) {
	args := m.Called(templateValue, storage)

	return args.String(0), args.Error(1)
}

func (m *mockedJSONValidator) Validate(document, schemaPath string) error {
	args := m.Called(document, schemaPath)

	return args.Error(0)
}

func (m *mockedHTTPContext) GetHTTPClient() *http.Client {
	args := m.Called()

	return args.Get(0).(*http.Client)
}

func (m *mockedHTTPContext) GetLastResponse() (*http.Response, error) {
	args := m.Called()

	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *mockedHTTPContext) GetLastResponseBody() ([]byte, error) {
	args := m.Called()

	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockedJsonPathFinder) Find(expr string, jsonBytes []byte) (any, error) {
	args := m.Called(expr, jsonBytes)

	return args.Get(0).(any), args.Error(1)
}

func TestState_IPrepareNewRequestToAndSaveItAs(t *testing.T) {
	type args struct {
		method      string
		urlTemplate string
		cacheKey    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "successfully prepare new request",
			args:    args{method: http.MethodGet, urlTemplate: "/", cacheKey: "MY_GET_REQUEST"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")
			if err := s.IPrepareNewRequestToAndSaveItAs(tt.args.method, tt.args.urlTemplate, tt.args.cacheKey); (err != nil) != tt.wantErr {
				t.Errorf("IPrepareNewRequestToAndSaveItAs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				req, err := s.GetPreparedRequest(tt.args.cacheKey)
				if err != nil {
					t.Errorf("%v", err)
				}

				if req.Method != tt.args.method {
					t.Errorf("obtained request has different method: %s, expected: %s", req.Method, tt.args.method)
				}
			}
		})
	}
}

func TestState_ISetFollowingHeadersForPreparedRequest(t *testing.T) {
	type fields struct {
		reqMethod string
		reqUri    string
		cacheKey  string
	}
	type args struct {
		cacheKey        string
		headersTemplate string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// fails
		{
			name:    "headers have invalid format",
			fields:  fields{},
			args:    args{cacheKey: "", headersTemplate: "abc"},
			wantErr: true,
		},
		{
			name:    "there is no prepared request to set headers for",
			fields:  fields{},
			args:    args{cacheKey: "abc", headersTemplate: `{"Content-Type": "application/json"}`},
			wantErr: true,
		},
		{
			name:    "cache key does not point at request",
			fields:  fields{reqMethod: "GET", reqUri: "/", cacheKey: "abc"},
			args:    args{cacheKey: "xxx", headersTemplate: `{"Content-Type": "application/json"}`},
			wantErr: true,
		},
		{
			name:    "headers are in unsupported XML format",
			fields:  fields{reqMethod: "GET", reqUri: "/", cacheKey: "abc"},
			args:    args{cacheKey: "abc", headersTemplate: `<data>abc</data>`},
			wantErr: true,
		},

		//success
		{
			name:    "successfully set request headers in JSON format",
			fields:  fields{reqMethod: "GET", reqUri: "/", cacheKey: "MY_REQUEST"},
			args:    args{cacheKey: "MY_REQUEST", headersTemplate: `{"Content-Type": "application/json"}`},
			wantErr: false,
		},
		{
			name:   "successfully set request header with YAML format",
			fields: fields{reqMethod: "GET", reqUri: "/", cacheKey: "MY_REQUEST"},
			args: args{cacheKey: "MY_REQUEST", headersTemplate: `---
Content-Type: application/json`},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			err := s.IPrepareNewRequestToAndSaveItAs(tt.fields.reqMethod, tt.fields.reqUri, tt.fields.cacheKey)
			if err != nil {
				t.Errorf("%v", err)
			}

			if err = s.ISetFollowingHeadersForPreparedRequest(tt.args.cacheKey, tt.args.headersTemplate); (err != nil) != tt.wantErr {
				t.Errorf("ISetFollowingHeadersForPreparedRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_ISetFollowingBodyForPreparedRequest(t *testing.T) {
	type fields struct {
		reqMethod string
		reqUri    string
		cacheKey  string
	}
	type args struct {
		cacheKey     string
		bodyTemplate string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// Failed
		{
			name:    "missing prepared request",
			fields:  fields{},
			args:    args{cacheKey: "abc", bodyTemplate: `{"Content-Type": "application/json"}`},
			wantErr: true,
		},
		{
			name:    "cache key does not point at request",
			fields:  fields{reqMethod: "GET", reqUri: "/", cacheKey: "MY_REQUEST"},
			args:    args{cacheKey: "OTHER_REQUEST", bodyTemplate: `{"Content-Type": "application/json"}`},
			wantErr: true,
		},

		// Successful
		{
			name:    "successfully set request body in JSON format",
			fields:  fields{reqMethod: "GET", reqUri: "/", cacheKey: "MY_REQUEST"},
			args:    args{cacheKey: "MY_REQUEST", bodyTemplate: `{"a": "b"}`},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")
			err := s.IPrepareNewRequestToAndSaveItAs(tt.fields.reqMethod, tt.fields.reqUri, tt.fields.cacheKey)
			if err != nil {
				t.Errorf("%v", err)
			}

			if err := s.ISetFollowingBodyForPreparedRequest(tt.args.cacheKey, tt.args.bodyTemplate); (err != nil) != tt.wantErr {
				t.Errorf("ISetFollowingBodyForPreparedRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_ISetFollowingCookiesForPreparedRequest(t *testing.T) {
	layout := "Jan 2, 2006 at 3:04pm (MST)"
	tm, err := time.Parse(layout, "Feb 4, 2014 at 6:05pm (PST)")
	if err != nil {
		t.Errorf("invalid time parsing")
	}

	mTemplateEngine := new(mockedTemplateEngine)
	mFormatter := new(mockedFormatter)

	type fields struct {
		mockFunc func()
	}

	type args struct {
		cacheKey        string
		cookiesTemplate string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantCookies []http.Cookie
	}{
		// failed
		{
			name: "template engine returns error",
			fields: fields{mockFunc: func() {
				mTemplateEngine.On("Replace", "https://www.example.com", mock.Anything).
					Return("https://www.example.com", nil).Once()

				mTemplateEngine.On("Replace", "", mock.Anything).
					Return("", errors.New("abc")).Once()
			}},
			args:    args{cacheKey: "a", cookiesTemplate: ""},
			wantErr: true,
		},
		{
			name: "unsupported XML format",
			fields: fields{mockFunc: func() {
				mTemplateEngine.On("Replace", "b", mock.Anything).
					Return("<data>xml</data>", nil).Once()

				mTemplateEngine.On("Replace", "https://www.example.com", mock.Anything).
					Return("https://www.example.com", nil).Once()
			}}, args: args{cacheKey: "a", cookiesTemplate: "b"},
			wantErr: true,
		},
		{
			name: "deserializer returns error",
			fields: fields{mockFunc: func() {
				mTemplateEngine.On("Replace", "", mock.Anything).
					Return("", nil).Once()

				mTemplateEngine.On("Replace", "https://www.example.com", mock.Anything).
					Return("https://www.example.com", nil).Once()

				mFormatter.On("Deserialize", []byte(""), mock.Anything).Return(errors.New("abc"))

			}}, args: args{cacheKey: "a", cookiesTemplate: ""},
			wantErr: true,
		},

		// successful
		{
			name: "successfully set cookies #1",
			fields: fields{mockFunc: func() {
			}}, args: args{
				cacheKey: "MY_REQ",
				cookiesTemplate: `[
	{
		"name": "token",
		"value": "abc"
	}
]`,
			}, wantErr: false,
			wantCookies: []http.Cookie{{Name: "token", Value: "abc"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			if tt.wantErr {
				s.SetJSONFormatter(mFormatter)
				s.SetTemplateEngine(mTemplateEngine)
			}
			tt.fields.mockFunc()

			if err = s.IPrepareNewRequestToAndSaveItAs("GET", "https://www.example.com", tt.args.cacheKey); err != nil {
				t.Errorf("%s", err.Error())
			}

			s.Cache.Save("NOW", tm)

			if err := s.ISetFollowingCookiesForPreparedRequest(tt.args.cacheKey, tt.args.cookiesTemplate); (err != nil) != tt.wantErr {
				t.Errorf("ISetFollowingCookiesForPreparedRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				reqWithCookies, err := s.GetPreparedRequest(tt.args.cacheKey)
				if err != nil {
					t.Errorf("err: %s", err.Error())
				}

				for _, wantCookie := range tt.wantCookies {
					exists := false
					for _, cookie := range reqWithCookies.Cookies() {
						if cookie.Name == wantCookie.Name && cookie.Value == wantCookie.Value && cookie.HttpOnly == wantCookie.HttpOnly && cookie.Secure == wantCookie.Secure {
							exists = true
						}
					}

					if !exists {
						t.Errorf("cookie %v does not exists in http.Request. All cookies: %+v", wantCookie, reqWithCookies.Cookies())
					}
				}
			}
		})
	}
}

func TestState_TheResponseStatusCodeShouldBe(t *testing.T) {
	type fields struct {
		cache        map[string]any
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		code int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// Failed
		{name: "invalid code #1, code less than 200", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 1}, wantErr: true},
		{name: "invalid code #2, code over 599", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 600}, wantErr: true},
		{name: "invalid code #3, different code expected", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 400}, wantErr: true},

		// Successful
		{name: "valid code #1", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 200}, wantErr: false},
		{name: "valid code #1", fields: fields{lastResponse: &http.Response{StatusCode: 400}}, args: args{code: 400}, wantErr: false},
		{name: "valid code #1", fields: fields{lastResponse: &http.Response{StatusCode: 511}}, args: args{code: 511}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(tt.fields.isDebug, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := s.TheResponseStatusCodeShouldBe(tt.args.code); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseStatusCodeShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheResponseBodyShouldHaveFormat(t *testing.T) {
	yaml := `
---
user:
   name: four
   last_name: b
`

	xml := `<this>is xml</this>`
	plainText := `this is plain text`
	json := `{"this_is": "json"}`
	type fields struct {
		body []byte
	}

	type args struct {
		dataFormat format.DataFormat
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		// JSON
		{name: "response body has JSON format and JSON format expected", args: args{dataFormat: format.JSON}, fields: fields{body: []byte(json)}, wantErr: false},
		{name: "response body has plain text format but JSON format expected", args: args{dataFormat: format.JSON}, fields: fields{body: []byte(plainText)}, wantErr: true},
		{name: "response body has XML format but JSON format expected", args: args{dataFormat: format.JSON}, fields: fields{body: []byte(xml)}, wantErr: true},
		{name: "response body has YAML format but JSON format expected", args: args{dataFormat: format.JSON}, fields: fields{body: []byte(yaml)}, wantErr: true},

		// YAML
		{name: "response body has JSON format but YAML format expected", args: args{dataFormat: format.YAML}, fields: fields{body: []byte(json)}, wantErr: true},
		{name: "response body has plain text format but YAML format expected", args: args{dataFormat: format.YAML}, fields: fields{body: []byte(plainText)}, wantErr: true},
		{name: "response body has XML format but YAML format expected", args: args{dataFormat: format.YAML}, fields: fields{body: []byte(xml)}, wantErr: true},
		{name: "response body has YAML format and YAML format expected", args: args{dataFormat: format.YAML}, fields: fields{body: []byte(yaml)}, wantErr: false},

		// XML
		{name: "response body has JSON format but XML format expected", args: args{dataFormat: format.XML}, fields: fields{body: []byte(json)}, wantErr: true},
		{name: "response body has plain text format but XML format expected", args: args{dataFormat: format.XML}, fields: fields{body: []byte(plainText)}, wantErr: true},
		{name: "response body has XML format and XML format expected", args: args{dataFormat: format.XML}, fields: fields{body: []byte(xml)}, wantErr: false},
		{name: "response body has YAML format but XML format expected", args: args{dataFormat: format.XML}, fields: fields{body: []byte(yaml)}, wantErr: true},

		// plain text
		{name: "response body has JSON format but plain text format expected", args: args{dataFormat: format.PlainText}, fields: fields{body: []byte(json)}, wantErr: true},
		{name: "response body has plain text format and plain text format expected", args: args{dataFormat: format.PlainText}, fields: fields{body: []byte(plainText)}, wantErr: false},
		{name: "response body has XML format but plain text format expected", args: args{dataFormat: format.PlainText}, fields: fields{body: []byte(xml)}, wantErr: true},
		{name: "response body has YAML format but plain text format expected", args: args{dataFormat: format.PlainText}, fields: fields{body: []byte(yaml)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})

			if err := s.TheResponseBodyShouldHaveFormat(tt.args.dataFormat); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseBodyShouldHaveFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_IGenerateARandomIntInTheRangeToAndSaveItAs(t *testing.T) {
	s := NewDefaultAPIContext(false, "")
	for i := 0; i < 100; i++ {
		if err := s.IGenerateARandomIntInTheRangeToAndSaveItAs(0, 100000, "RANDOM_INT"); (err != nil) != false {
			t.Errorf("IGenerateARandomIntInTheRangeToAndSaveItAs() error = %v, wantErr %v", err, false)
		}

		randomInteger, err := s.Cache.GetSaved("RANDOM_INT")
		if err != nil {
			t.Errorf("%v", err)
		}

		randomInt := randomInteger.(int)
		if randomInt < 0 {
			t.Errorf("randomInt should not be less than 0")
		}

		if randomInt > 100000 {
			t.Errorf("randomInt should not be greater than 100000")
		}
	}
}

func TestState_IGenerateARandomRunesWithoutUnicodeCharactersInTheRangeToAndSaveItAs(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	rndStringASCII := s.GeneratorRandomRunes(stringutils.CharsetASCII)
	for i := 0; i < 10; i++ {
		key := "TEST_" + strconv.Itoa(i)
		if err := rndStringASCII(5, 10, key); err != nil {
			t.Errorf(err.Error())
		}

		strI, err := s.Cache.GetSaved(key)
		if err != nil {
			t.Errorf(err.Error())
		}

		str, ok := strI.(string)
		if !ok {
			t.Errorf("%+v is not string", strI)
		}

		rStr := []rune(str)

		if !(len(rStr) >= 5 && len(rStr) <= 10) {
			t.Errorf("%v should have length between 5 - 10, got: %d", str, len(rStr))
		}
	}

	rndStringUnicode := s.GeneratorRandomRunes(stringutils.CharsetUnicode)
	for i := 0; i < 10; i++ {
		key := "TEST_" + strconv.Itoa(i)
		if err := rndStringUnicode(5, 10, key); err != nil {
			t.Errorf(err.Error())
		}

		strI, err := s.Cache.GetSaved(key)
		if err != nil {
			t.Errorf(err.Error())
		}

		str, ok := strI.(string)
		if !ok {
			t.Errorf("%+v is not string", strI)
		}

		rStr := []rune(str)

		if !(len(rStr) >= 5 && len(rStr) <= 10) {
			t.Errorf("%v should have length between 5 - 10, got: %d", str, len(rStr))
		}
	}
}

func TestState_IGenerateArandomSentenceInTheRangeFromToWordsAndSaveItAs_ASCII(t *testing.T) {
	s := NewDefaultAPIContext(false, "")
	sentenceGen := s.GeneratorRandomSentence("ab", 1, 1)

	for i := 0; i < 10; i++ {
		rndNumberOfWords, _ := mathutils.RandomInt(2, 10)
		cacheKeyRnd := fmt.Sprintf("TEST_%d", i)
		if err := sentenceGen(2, rndNumberOfWords, cacheKeyRnd); err != nil {
			t.Errorf("error during sentence generation, err: %s", err.Error())
		}

		sentenceFromCache, err := s.Cache.GetSaved(cacheKeyRnd)
		if err != nil {
			t.Errorf("error during obtaining sentence from cache, err: %s", err.Error())
		}

		obtainedSentence, ok := sentenceFromCache.(string)
		if !ok {
			t.Errorf("error during type checking. Expected %+v to be string", obtainedSentence)
		}

		words := strings.Split(obtainedSentence, " ")
		if len(words) < 2 || len(words) > rndNumberOfWords {
			t.Errorf("expected sentence to have between (%d, %d) words, got %d, sentence: %s", 2, rndNumberOfWords, len(words), obtainedSentence)
		}
	}
}

func TestState_IGenerateArandomSentenceInTheRangeFromToWordsAndSaveItAs_Unicode(t *testing.T) {
	s := NewDefaultAPIContext(false, "")
	sentenceGen := s.GeneratorRandomSentence("🤡🤖🧟🏋🥇", 1, 1)

	for i := 0; i < 10; i++ {
		rndNumberOfWords, _ := mathutils.RandomInt(2, 10)
		cacheKeyRnd := fmt.Sprintf("TEST_%d", i)
		if err := sentenceGen(2, rndNumberOfWords, cacheKeyRnd); err != nil {
			t.Errorf("error during sentence generation, err: %s", err.Error())
		}

		sentenceFromCache, err := s.Cache.GetSaved(cacheKeyRnd)
		if err != nil {
			t.Errorf("error during obtaining sentence from cache, err: %s", err.Error())
		}

		obtainedSentence, ok := sentenceFromCache.(string)
		if !ok {
			t.Errorf("error during type checking. Expected %+v to be string", obtainedSentence)
		}

		words := strings.Split(obtainedSentence, " ")
		if len(words) < 2 || len(words) > rndNumberOfWords {
			t.Errorf("expected sentence to have between (%d, %d) words, got %d, sentence: %s", 2, rndNumberOfWords, len(words), obtainedSentence)
		}
	}
}

func TestState_IGetTimeAndTravelByAndSaveItAs(t *testing.T) {
	layout := "Jan 2, 2006 at 3:04pm (MST)"
	tm, err := time.Parse(layout, "Feb 4, 2014 at 6:05pm (PST)")
	if err != nil {
		t.Errorf("invalid time parsing")
	}

	type args struct {
		t             time.Time
		timeDirection timeutils.TimeDirection
		timeDuration  time.Duration
		cacheKey      string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantTime string
	}{
		// failed
		{name: "unknown time direction", args: args{
			t:             tm,
			timeDirection: timeutils.TimeDirection("abc"),
			timeDuration:  5 * time.Minute,
			cacheKey:      "TIME",
		}, wantErr: true, wantTime: "Jan 1, 0001 at 12:00am (UTC)"},

		// successful
		{name: "Move forward 5 min", args: args{
			t:             tm,
			timeDirection: timeutils.TimeDirectionForward,
			timeDuration:  5 * time.Minute,
			cacheKey:      "TIME",
		}, wantErr: false, wantTime: "Feb 4, 2014 at 6:10pm (PST)"},
		{name: "Move forward 0 sec", args: args{
			t:             tm,
			timeDirection: timeutils.TimeDirectionForward,
			timeDuration:  0 * time.Second,
			cacheKey:      "TIME",
		}, wantErr: false, wantTime: "Feb 4, 2014 at 6:05pm (PST)"},
		{name: "Move backward 0 sec", args: args{
			t:             tm,
			timeDirection: timeutils.TimeDirectionBackward,
			timeDuration:  0 * time.Second,
			cacheKey:      "TIME",
		}, wantErr: false, wantTime: "Feb 4, 2014 at 6:05pm (PST)"},
		{name: "Move backward 5 min", args: args{
			t:             tm,
			timeDirection: timeutils.TimeDirectionBackward,
			timeDuration:  5 * time.Minute,
			cacheKey:      "TIME",
		}, wantErr: false, wantTime: "Feb 4, 2014 at 6:00pm (PST)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, ".")

			if err := s.IGetTimeAndTravelByAndSaveItAs(tt.args.t, tt.args.timeDirection, tt.args.timeDuration, tt.args.cacheKey); (err != nil) != tt.wantErr {
				t.Errorf("IGetTimeAndTravelByAndSaveItAs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				ti, err := s.Cache.GetSaved(tt.args.cacheKey)
				if err != nil {
					t.Errorf("err: %v", err)
				}

				tim, ok := ti.(time.Time)
				if !ok {
					t.Errorf("cache key %s is not time.Time", tt.args.cacheKey)
				}

				if tim.Format(layout) != tt.wantTime {
					t.Errorf("expected: %s, got: %s", tt.wantTime, tim.Format(layout))
				}
			}
		})
	}
}

func TestState_TheResponseShouldHaveNode(t *testing.T) {
	json := `{
	"users": [
		{
			"name": "abc"
		},
		{
			"name": "xxx"
		}
	]
}`

	yaml := `---
users:
- name: abc
- name: xxx
`

	xml := `<?xml version="1.0"?>
<users>
	<user>
		<name>abc</name>
	</user>
	<user>
		<name>xxx</name>
	</user>
</users>`

	type fields struct {
		body      []byte
		cacheData map[string]any
	}
	type args struct {
		dataFormat  format.DataFormat
		expressions string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// JSON
		{name: "missing last response body", fields: fields{}, args: args{}, wantErr: true},
		{name: "last response body is missing expected key in JSON", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  format.JSON,
			expressions: "age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - qjson expression", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  format.JSON,
			expressions: "users",
		}, wantErr: false},
		{name: "last response body has expected key #2 - Oliveagle expression", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  format.JSON,
			expressions: "$.users[0]",
		}, wantErr: false},
		{name: "last response body has expected key #3 - qjson expression ", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  format.JSON,
			expressions: "users[1].name",
		}, wantErr: false},
		{name: "last response body has expected key #4 - qjson expression with template value", fields: fields{body: []byte(json), cacheData: map[string]any{
			"USER_ID": 1,
		}}, args: args{
			dataFormat:  format.JSON,
			expressions: "users[{{.USER_ID}}].name",
		}, wantErr: false},

		// XML
		{name: "last response body is missing expected key in XML", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  format.XML,
			expressions: "age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  format.XML,
			expressions: "//users",
		}, wantErr: false},
		{name: "last response body has expected key #2 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  format.XML,
			expressions: "//users//user[1]",
		}, wantErr: false},
		{name: "last response body has expected key #3 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  format.XML,
			expressions: "//user[2]//name",
		}, wantErr: false},
		{name: "last response body has expected key #4 - Antchfx expression with template value", fields: fields{body: []byte(xml), cacheData: map[string]any{
			"USER_ID": 2,
		}}, args: args{
			dataFormat:  format.XML,
			expressions: "//user[{{.USER_ID}}]//name",
		}, wantErr: false},

		// YAML
		{name: "last response body is missing expected key in YAML", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - Goccy expression", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.users",
		}, wantErr: false},
		{name: "last response body has expected key #2 - Goccy expression", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.users[0].name",
		}, wantErr: false},
		{name: "last response body has expected key #3 - Goccy expression with template value", fields: fields{body: []byte(yaml), cacheData: map[string]any{
			"USER_ID": 0,
		}}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.users[{{.USER_ID}}].name",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			for k, v := range tt.fields.cacheData {
				s.Cache.Save(k, v)
			}

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})

			if err := s.TheResponseShouldHaveNode(tt.args.dataFormat, tt.args.expressions); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldHaveNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheResponseShouldHaveNodes(t *testing.T) {
	json := `{
	"users": [
		{
			"name": "abc"
		},
		{
			"name": "xxx"
		}
	]
}`

	yaml := `---
users:
- name: abc
- name: xxx
`

	xml := `<?xml version="1.0"?>
<users>
	<user>
		<name>abc</name>
	</user>
	<user>
		<name>xxx</name>
	</user>
</users>`

	type fields struct {
		body      []byte
		cacheKeys map[string]any
	}
	type args struct {
		dataFormat  format.DataFormat
		expressions string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// JSON
		{name: "missing last response body", fields: fields{}, args: args{}, wantErr: true},
		{name: "last response body is missing key", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  format.JSON,
			expressions: "age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - qjson expression", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  format.JSON,
			expressions: "users",
		}, wantErr: false},
		{name: "last response body has expected keys #2 - different expressions", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  format.JSON,
			expressions: "users, $.users",
		}, wantErr: false},
		{name: "last response body has expected keys #3 - qjson expressions", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  format.JSON,
			expressions: "users, users[0]",
		}, wantErr: false},
		{name: "last response body has expected keys #4 - qjson expressions", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  format.JSON,
			expressions: "users[0].name, users[1].name",
		}, wantErr: false},
		{name: "last response body has expected keys #1 - qjson expressions with template values", fields: fields{body: []byte(json), cacheKeys: map[string]any{
			"USER1_ID": 0,
			"USER2_ID": 1,
		}}, args: args{
			dataFormat:  format.JSON,
			expressions: "users[{{.USER1_ID}}].name, users[{{.USER2_ID}}].name",
		}, wantErr: false},

		// XML
		{name: "last response body is missing key", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  format.XML,
			expressions: "age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  format.XML,
			expressions: "//users",
		}, wantErr: false},
		{name: "last response body has expected keys #2 - Antchfx expressions", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  format.XML,
			expressions: "//users, //users",
		}, wantErr: false},
		{name: "last response body has expected keys #3 - Antchfx expressions", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  format.XML,
			expressions: "//users, //users//user[1]",
		}, wantErr: false},
		{name: "last response body has expected keys #4 - Antchfx expressions", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  format.XML,
			expressions: "//user[1]//name, //user[2]//name",
		}, wantErr: false},
		{name: "last response body has expected keys #5 - Antchfx expressions with template values", fields: fields{body: []byte(xml), cacheKeys: map[string]any{
			"USER1_ID": 1,
			"USER2_ID": 2,
		}}, args: args{
			dataFormat:  format.XML,
			expressions: "//user[{{.USER1_ID}}]//name, //user[{{.USER2_ID}}]//name",
		}, wantErr: false},

		// YAML
		{name: "last response body is missing yaml key", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - Goccy expression", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.users",
		}, wantErr: false},
		{name: "last response body has expected keys #2 - Goccy expressions", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.users, $.users",
		}, wantErr: false},
		{name: "last response body has expected keys #3 - Goccy expressions", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.users, $.users[0]",
		}, wantErr: false},
		{name: "last response body has expected keys #4 - Goccy expressions", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.users[0].name, $.users[1].name",
		}, wantErr: false},
		{name: "last response body has expected keys #5 - Goccy expressions with template values", fields: fields{body: []byte(yaml), cacheKeys: map[string]any{
			"USER1_ID": 0,
			"USER2_ID": 1,
		}}, args: args{
			dataFormat:  format.YAML,
			expressions: "$.users[{{.USER1_ID}}].name, $.users[{{.USER2_ID}}].name",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			for k, v := range tt.fields.cacheKeys {
				s.Cache.Save(k, v)
			}

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})

			if err := s.TheResponseShouldHaveNodes(tt.args.dataFormat, tt.args.expressions); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldHaveNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheNodeShouldNotBe(t *testing.T) {
	type fields struct {
		saved        map[string]any
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		df     format.DataFormat
		node   string
		goType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// JSON
		{name: "is not nil value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "nil"}, wantErr: false},
		{name: "is nil value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": nil
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "nil"}, wantErr: true},
		{name: "is null value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "nil"}, wantErr: true},
		{name: "is not string #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "string"}, wantErr: false},
		{name: "is not string #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "string"}, wantErr: false},
		{name: "is string", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "string"}, wantErr: true},
		{name: "is not int #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: false},
		{name: "is not int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2.1
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: false},
		{name: "is int #1 <- special case", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2.0
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: true},
		{name: "is int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: true},
		{name: "is float", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1.0
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: true},
		{name: "is not float #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "float"}, wantErr: false},
		{name: "is not float #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "float"}, wantErr: false},
		{name: "is not float #3", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": true
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "float"}, wantErr: false},
		{name: "is bool", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": true
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "bool"}, wantErr: true},
		{name: "is not bool #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "bool"}, wantErr: false},
		{name: "is not bool #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "false"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "bool"}, wantErr: false},
		{name: "is map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "map"}, wantErr: true},
		{name: "is map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {"name": "pawel"}
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "map"}, wantErr: true},
		{name: "is not map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "map"}, wantErr: false},
		{name: "is not map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "pawel"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "map"}, wantErr: false},
		{name: "is slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": []
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "slice"}, wantErr: true},
		{name: "is slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": ["1"]
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "slice"}, wantErr: true},
		{name: "is not slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "xxx"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "slice"}, wantErr: false},
		{name: "is not slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "slice"}, wantErr: false},
		{name: "unknown type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{node: "user", goType: "xxx"}, wantErr: true},

		//YAML
		{name: "is not nil value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: abc`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "nil"}, wantErr: false},
		{name: "is null value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "nil"}, wantErr: true},
		{name: "is not string #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "string"}, wantErr: false},
		{name: "is not string #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: 2`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "string"}, wantErr: false},
		{name: "is string", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: abc`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "string"}, wantErr: true},
		{name: "is not int #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: false},
		{name: "is not int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: 2.1`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: false},
		{name: "is int #1 <- special case", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: 2.0`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: true},
		{name: "is int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: -1`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: true},
		{name: "is float", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: -1.0`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: true},
		{name: "is not float #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: -1`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "float"}, wantErr: false},
		{name: "is not float #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "float"}, wantErr: false},
		{name: "is not float #3", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: true`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "float"}, wantErr: false},
		{name: "is bool", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: true`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "bool"}, wantErr: true},
		{name: "is not bool #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "bool"}, wantErr: false},
		{name: "is not bool #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: "false"`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "bool"}, wantErr: false},
		{name: "is map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
foo: bar
     pleh: help
     stuff:
       foo: bar
       bar: foo`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "map"}, wantErr: true},
		{name: "is map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
xmas-fifth-day:
   calling-birds: four
   french-hens: 3
   golden-rings: 5`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "map"}, wantErr: true},
		{name: "is not map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "map"}, wantErr: false},
		{name: "is not map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: "pawel"`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "map"}, wantErr: false},
		{name: "is slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
users:
	- huey
	- dewey`))},
		}, args: args{df: format.YAML, node: "$.users", goType: "slice"}, wantErr: true},
		{name: "is slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
recipe:
	ingredients:
		- oil
		- tomato`))},
		}, args: args{df: format.YAML, node: "$.recipe.ingredients", goType: "slice"}, wantErr: true},
		{name: "is not slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "slice"}, wantErr: false},
		{name: "is not slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
xmas-fifth-day:
   calling-birds: four
   french-hens: 3
   golden-rings: 5`))},
		}, args: args{df: format.YAML, node: "$.xmas-fifth-day", goType: "slice"}, wantErr: false},
		{name: "unknown type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
users:
	- huey
	- dewey`))},
		}, args: args{node: "$.users", goType: "xxx"}, wantErr: true},
		{name: "format XML is not supported", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(``))},
		}, args: args{df: format.XML, node: "$.users", goType: "xxx"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := NewDefaultAPIContext(tt.fields.isDebug, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := af.TheNodeShouldNotBe(tt.args.df, tt.args.node, tt.args.goType); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldNotBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheNodeShouldBe(t *testing.T) {
	type fields struct {
		saved        map[string]any
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		df     format.DataFormat
		node   string
		goType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "is not nil value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "nil"}, wantErr: true},
		{name: "is nil value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": nil
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "nil"}, wantErr: false},
		{name: "is null value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "nil"}, wantErr: false},
		{name: "is not string #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "string"}, wantErr: true},
		{name: "is not string #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "string"}, wantErr: true},
		{name: "is string", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "string"}, wantErr: false},
		{name: "is not int #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: true},
		{name: "is not int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2.1
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: true},
		{name: "is int #1 <- special case", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": 2.0
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: false},
		{name: "is int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: false},
		{name: "is float", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1.0
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "int"}, wantErr: false},
		{name: "is not float #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": -1
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "float"}, wantErr: true},
		{name: "is not float #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "float"}, wantErr: true},
		{name: "is not float #3", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": true
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "float"}, wantErr: true},
		{name: "is bool", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": true
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "bool"}, wantErr: false},
		{name: "is not bool #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "bool"}, wantErr: true},
		{name: "is not bool #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "false"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "bool"}, wantErr: true},
		{name: "is map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "map"}, wantErr: false},
		{name: "is map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {"name": "pawel"}
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "map"}, wantErr: false},
		{name: "is not map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": null
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "map"}, wantErr: true},
		{name: "is not map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "pawel"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "map"}, wantErr: true},
		{name: "is slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": []
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "slice"}, wantErr: false},
		{name: "is slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": ["1"]
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "slice"}, wantErr: false},
		{name: "is not slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "xxx"
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "slice"}, wantErr: true},
		{name: "is not slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "slice"}, wantErr: true},
		{name: "unknown type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {}
}`))},
		}, args: args{df: format.JSON, node: "user", goType: "xxx"}, wantErr: true},

		//YAML
		{name: "is not nil value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: abc`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "nil"}, wantErr: true},
		{name: "is null value", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "nil"}, wantErr: false},
		{name: "is not string #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "string"}, wantErr: true},
		{name: "is not string #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: 2`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "string"}, wantErr: true},
		{name: "is string", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: abc`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "string"}, wantErr: false},
		{name: "is not int #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: true},
		{name: "is not int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: 2.1`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: true},
		{name: "is int #1 <- special case", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: 2.0`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: false},
		{name: "is int #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: -1`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: false},
		{name: "is float", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: -1.0`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "int"}, wantErr: false},
		{name: "is not float #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: -1`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "float"}, wantErr: true},
		{name: "is not float #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "float"}, wantErr: true},
		{name: "is not float #3", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: true`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "float"}, wantErr: true},
		{name: "is bool", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: true`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "bool"}, wantErr: false},
		{name: "is not bool #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "bool"}, wantErr: true},
		{name: "is not bool #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: "false"`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "bool"}, wantErr: true},
		{name: "is map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
foo: bar
     pleh: help
     stuff:
       foo: bar
       bar: foo`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "map"}, wantErr: false},
		{name: "is map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
xmas-fifth-day:
   calling-birds: four
   french-hens: 3
   golden-rings: 5`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "map"}, wantErr: false},
		{name: "is not map #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "map"}, wantErr: true},
		{name: "is not map #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: "pawel"`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "map"}, wantErr: true},
		{name: "is slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
users:
	- huey
	- dewey`))},
		}, args: args{df: format.YAML, node: "$.users", goType: "slice"}, wantErr: false},
		{name: "is slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
recipe:
	ingredients:
		- oil
		- tomato`))},
		}, args: args{df: format.YAML, node: "$.recipe.ingredients", goType: "slice"}, wantErr: false},
		{name: "is not slice #1", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
user: null`))},
		}, args: args{df: format.YAML, node: "$.user", goType: "slice"}, wantErr: true},
		{name: "is not slice #2", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
xmas-fifth-day:
   calling-birds: four
   french-hens: 3
   golden-rings: 5`))},
		}, args: args{df: format.YAML, node: "$.xmas-fifth-day", goType: "slice"}, wantErr: true},
		{name: "unknown type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`---
users:
	- huey
	- dewey`))},
		}, args: args{node: "$.users", goType: "xxx"}, wantErr: true},
		{name: "format XML is not supported", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(``))},
		}, args: args{df: format.XML, node: "$.users", goType: "xxx"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := NewDefaultAPIContext(tt.fields.isDebug, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := af.TheNodeShouldBe(tt.args.df, tt.args.node, tt.args.goType); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheNodeShouldBeSliceOfLength(t *testing.T) {
	json := `{
	"count": 2,
	"data": [
		{
			"name": "abc"
		},
		{
			"name": "xxx"
		}
	]
}`

	yaml := `---
count: 2
data:
	- name: abc
	- name: xxx`

	xml := `<?xml version="1.0"?>
<root>
	<count>2</count>
	<data>
		<user>
			<name>abc</name>
		</user>
		<user>
			<name>xxx</name>
		</user>
	</data>
</root>`

	type fields struct {
		body []byte
	}
	type args struct {
		dataFormat format.DataFormat
		expr       string
		length     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//JSON
		{name: "no body", fields: fields{body: nil}, args: args{}, wantErr: true},
		{name: "expression does not point at slice", fields: fields{body: []byte(json)}, args: args{
			dataFormat: format.JSON,
			expr:       "count",
			length:     0,
		}, wantErr: true},
		{name: "expression does point at slice but expected invalid length", fields: fields{body: []byte(json)}, args: args{
			dataFormat: format.JSON,
			expr:       "data",
			length:     1,
		}, wantErr: true},
		{name: "expression does point at slice and expected proper length", fields: fields{body: []byte(json)}, args: args{
			dataFormat: format.JSON,
			expr:       "data",
			length:     2,
		}, wantErr: false},

		//YAML
		{name: "expression does not point at slice", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat: format.YAML,
			expr:       "$.count",
			length:     0,
		}, wantErr: true},
		{name: "expression does point at slice but expected invalid length", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat: format.YAML,
			expr:       "$.data",
			length:     1,
		}, wantErr: true},
		{name: "expression does point at slice and expected proper length", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat: format.YAML,
			expr:       "$.data",
			length:     2,
		}, wantErr: false},

		//XML
		{name: "expression does not point at slice", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: format.XML,
			expr:       "//count",
			length:     0,
		}, wantErr: true},
		{name: "expression does point at slice but expected invalid length", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: format.XML,
			expr:       "//data//name",
			length:     1,
		}, wantErr: true},
		{name: "expression does point at slice and expected proper length #1", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: format.XML,
			expr:       "//data//name",
			length:     2,
		}, wantErr: false},
		{name: "expression does point at slice and expected proper length #2", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: format.XML,
			expr:       "//user",
			length:     2,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})

			if err := s.TheNodeShouldBeSliceOfLength(tt.args.dataFormat, tt.args.expr, tt.args.length); (err != nil) != tt.wantErr {
				t.Errorf("TheNodeShouldBeSliceOfLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: Add YAML cases
func TestState_TheNodeShouldBeOfValue(t *testing.T) {
	type fields struct {
		lastResponse *http.Response
	}
	type args struct {
		df        format.DataFormat
		expr      string
		dataType  string
		dataValue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "empty json", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(``))},
		}, args: args{
			df:        format.JSON,
			expr:      "name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: true},
		{name: "json with first level field with string data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"name": "ivo"
}`))},
		}, args: args{
			df:        format.JSON,
			expr:      "name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: false},
		{name: "json with first level field with int data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"number": 10
}`))},
		}, args: args{
			df:        format.JSON,
			expr:      "number",
			dataType:  "int",
			dataValue: "10",
		}, wantErr: false},
		{name: "json with first level field with float64 data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"number": 10.1
}`))},
		}, args: args{
			df:        format.JSON,
			expr:      "number",
			dataType:  "float",
			dataValue: "10.1",
		}, wantErr: false},
		{name: "json with first level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"is": true
}`))},
		}, args: args{
			df:        format.JSON,
			expr:      "is",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data": {
		"name": "Is empty",
		"value": true
	}
}`))},
		}, args: args{
			df:        format.JSON,
			expr:      "data.value",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data":	[
			{
				"name": "Is empty",
				"value": true
			},
			{
				"name": "Is big",
				"value": false
			}
		]
}`))},
		}, args: args{
			df:        format.JSON,
			expr:      "data[1].value",
			dataType:  "bool",
			dataValue: "false",
		}, wantErr: false},
		{name: "json with second level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data":	[
			true,
			false
		]
}`))},
		}, args: args{
			df:        format.JSON,
			expr:      "data[1]",
			dataType:  "bool",
			dataValue: "false",
		}, wantErr: false},

		{name: "empty yaml", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(``))},
		}, args: args{
			df:        format.YAML,
			expr:      "$.name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: true},
		{name: "yaml with first level field with string data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
---
name: "ivo"
`))},
		}, args: args{
			df:        format.YAML,
			expr:      "$.name",
			dataType:  "string",
			dataValue: "ivo",
		}, wantErr: false},
		{name: "yaml with first level field with int data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"number": 10
}`))},
		}, args: args{
			df:        format.YAML,
			expr:      "$.number",
			dataType:  "int",
			dataValue: "10",
		}, wantErr: false},
		{name: "yaml with first level field with float64 data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"number": 10.1
}`))},
		}, args: args{
			df:        format.YAML,
			expr:      "$.number",
			dataType:  "float",
			dataValue: "10.1",
		}, wantErr: false},
		{name: "yaml with first level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"is": true
}`))},
		}, args: args{
			df:        format.YAML,
			expr:      "$.is",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "yaml with second level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data": {
		"name": "Is empty",
		"value": true
	}
}`))},
		}, args: args{
			df:        format.YAML,
			expr:      "$.data.value",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "yaml with second level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data":	[
			{
				"name": "Is empty",
				"value": true
			},
			{
				"name": "Is big",
				"value": false
			}
		]
}`))},
		}, args: args{
			df:        format.YAML,
			expr:      "$.data[1].value",
			dataType:  "bool",
			dataValue: "false",
		}, wantErr: false},
		{name: "yaml with second level field with bool data type", fields: fields{
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"data":	[
			{"is": false},
			{"is": true}
	]
}`))},
		}, args: args{
			df:        format.YAML,
			expr:      "$.data[1].is",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},

		{name: "XML node string", fields: fields{lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<root>
	<name>aa</name>
</root>`))}}, args: args{
			df:        format.XML,
			expr:      "//name",
			dataType:  "string",
			dataValue: "aa",
		}, wantErr: false},
		{name: "XML node bool", fields: fields{lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<root>
	<isLow>true</isLow>
</root>`))}}, args: args{
			df:        format.XML,
			expr:      "//isLow",
			dataType:  "bool",
			dataValue: "true",
		}, wantErr: false},
		{name: "XML node int", fields: fields{lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<root>
	<height>10</height>
</root>`))}}, args: args{
			df:        format.XML,
			expr:      "//height",
			dataType:  "int",
			dataValue: "10",
		}, wantErr: false},
		{name: "XML node float", fields: fields{lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<root>
	<height>10.02</height>
</root>`))}}, args: args{
			df:        format.XML,
			expr:      "//height",
			dataType:  "float",
			dataValue: "10.02",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := NewDefaultAPIContext(false, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := af.TheNodeShouldBeOfValue(tt.args.df, tt.args.expr, tt.args.dataType, tt.args.dataValue); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBeOfValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheNodeShouldMatchRegExp(t *testing.T) {
	mTemplateEngine := new(mockedTemplateEngine)
	mJsonPathResolver := new(mockedJsonPathFinder)

	type fields struct {
		cacheKeys      map[string]any
		templateEngine template.Engine
		pathResolvers  PathFinders
		respBody       string
		mockFunc       func()
	}
	type args struct {
		df             format.DataFormat
		expr           string
		regExpTemplate string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "template engine error", fields: fields{
			cacheKeys:      nil,
			templateEngine: mTemplateEngine,
			pathResolvers: PathFinders{
				JSON: nil,
				YAML: nil,
			},
			respBody: "",
			mockFunc: func() {
				mTemplateEngine.On("Replace", "abc", mock.Anything).
					Return("", errors.New("err")).Once()
			},
		}, args: args{
			df:             format.JSON,
			expr:           "",
			regExpTemplate: "abc",
		}, wantErr: true},
		{name: "json path finder error", fields: fields{
			cacheKeys:      nil,
			templateEngine: mTemplateEngine,
			pathResolvers:  PathFinders{JSON: mJsonPathResolver},
			respBody:       "",
			mockFunc: func() {
				mTemplateEngine.On("Replace", "abc", mock.Anything).
					Return("abc", nil).Once()

				mTemplateEngine.On("Replace", "xxx", mock.Anything).
					Return("xxx", nil).Once()

				mJsonPathResolver.On("Find", "xxx", []byte("")).
					Return(any(""), errors.New("err")).Once()
			},
		}, args: args{
			df:             format.JSON,
			expr:           "xxx",
			regExpTemplate: "abc",
		}, wantErr: true},
		{name: "node value does not match provided regExp #1", fields: fields{
			respBody: `{
	"name": "abcdef"
}`,
			mockFunc: func() {},
		}, args: args{
			df:             format.JSON,
			expr:           "name",
			regExpTemplate: "dd.*",
		}, wantErr: true},

		{name: "node value matches provided regExp #1", fields: fields{
			respBody: `{
	"name": "abcdef"
}`,
			mockFunc: func() {},
		}, args: args{
			df:             format.JSON,
			expr:           "name",
			regExpTemplate: "abc.*",
		}, wantErr: false},

		{name: "node value does not match provided regExp #2", fields: fields{
			respBody: `---
name: abcdef`,
			mockFunc: func() {},
		}, args: args{
			df:             format.YAML,
			expr:           "$.name",
			regExpTemplate: "dd.*",
		}, wantErr: true},

		{name: "node value matches provided regExp #2", fields: fields{
			respBody: `---
name: abcdef`,
			mockFunc: func() {},
		}, args: args{
			df:             format.YAML,
			expr:           "$.name",
			regExpTemplate: "abc.*",
		}, wantErr: false},
		{name: "node value matches provided regExp #3", fields: fields{
			respBody: `
<name>abcdef</name>`,
			mockFunc: func() {},
		}, args: args{
			df:             format.XML,
			expr:           "//name",
			regExpTemplate: "abc.*",
		}, wantErr: false},
		{name: "node value does not match provided regExp #3", fields: fields{
			respBody: `
<name>abcdef</name>`,
			mockFunc: func() {},
		}, args: args{
			df:             format.XML,
			expr:           "//name",
			regExpTemplate: "dd.*",
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			if len(tt.fields.cacheKeys) > 0 {
				for key, val := range tt.fields.cacheKeys {
					s.Cache.Save(key, val)
				}
			}

			if tt.fields.templateEngine != nil {
				s.SetTemplateEngine(tt.fields.templateEngine)
			}

			if tt.fields.pathResolvers.JSON != nil {
				s.SetJSONPathFinder(tt.fields.pathResolvers.JSON)
			}

			r := &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(tt.fields.respBody))}
			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, r)

			tt.fields.mockFunc()

			if err := s.TheNodeShouldMatchRegExp(tt.args.df, tt.args.expr, tt.args.regExpTemplate); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldMatchRegExp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheResponseShouldHaveHeader(t *testing.T) {
	type fields struct {
		cache        map[string]any
		lastResponse *http.Response
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "no headers in request", fields: fields{lastResponse: &http.Response{Header: map[string][]string{}}}, args: args{name: "Content-Type"}, wantErr: true},
		{name: "empty string provided as header name", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: ""}, wantErr: true},
		{name: "matching header #1 - case insensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "content-type"}, wantErr: false},
		{name: "matching header #2 - case sensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "Content-Type"}, wantErr: false},
		{name: "matching header #3", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{
				"Content-Length": {"30"},
				"Content-Type":   {"application/json"},
			},
		},
		}, args: args{name: "Content-Type"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := s.TheResponseShouldHaveHeader(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldHaveHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheResponseShouldHaveHeaderOfValue(t *testing.T) {
	type fields struct {
		cache        map[string]any
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "no headers in request", fields: fields{lastResponse: &http.Response{Header: map[string][]string{}}}, args: args{name: "Content-Type", value: "application/json"}, wantErr: true},
		{name: "empty string provided as header name", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "", value: "application/json"}, wantErr: true},
		{name: "matching header but improper value", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "content-type", value: "application/xml"}, wantErr: true},
		{name: "matching header #1 - case insensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "content-type", value: "application/json"}, wantErr: false},
		{name: "matching header #2 - case sensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "Content-Type", value: "application/json"}, wantErr: false},
		{name: "matching header #3", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{
				"Content-Length": {"30"},
				"Content-Type":   {"application/json"},
			},
		},
		}, args: args{name: "Content-Type", value: "application/json"}, wantErr: false},
		{name: "matching header using template value #3", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{
				"Content-Length": {"30"},
				"Content-Type":   {"application/json"},
			},
		},
			cache: map[string]any{"CONTENT_TYPE_JSON": "application/json"},
		}, args: args{name: "Content-Type", value: "{{.CONTENT_TYPE_JSON}}"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(tt.fields.isDebug, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)
			if tt.fields.cache != nil {
				for key, val := range tt.fields.cache {
					s.Cache.Save(key, val)
				}
			}

			if err := s.TheResponseShouldHaveHeaderOfValue(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldHaveHeaderOfValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_IValidateLastResponseBodyWithSchemaReference(t *testing.T) {
	type fields struct {
		resp      *http.Response
		validator validator.SchemaValidator
		mockFunc  func()
	}
	type args struct {
		schemaPath string
	}

	mJSONValidator := new(mockedJSONValidator)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "missing last response body", fields: fields{
			resp:      nil,
			validator: mJSONValidator,
			mockFunc: func() {

			},
		}, args: args{schemaPath: ""}, wantErr: true},
		{name: "validator fails", fields: fields{
			resp:      &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(""))},
			validator: mJSONValidator,
			mockFunc: func() {
				mJSONValidator.On("Validate", "", "").Return(errors.New("abc")).Once()
			},
		}, args: args{schemaPath: ""}, wantErr: true},
		{name: "validator succeeded", fields: fields{
			resp:      &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(""))},
			validator: mJSONValidator,
			mockFunc: func() {
				mJSONValidator.On("Validate", "", "").Return(nil).Once()
			},
		}, args: args{schemaPath: ""}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")
			s.SetSchemaReferenceValidator(tt.fields.validator)

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.resp)

			tt.fields.mockFunc()

			if err := s.IValidateLastResponseBodyWithSchemaReference(tt.args.schemaPath); (err != nil) != tt.wantErr {
				t.Errorf("IValidateLastResponseBodyWithSchemaReference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_IValidateLastResponseBodyWithSchemaString(t *testing.T) {
	type fields struct {
		resp      *http.Response
		validator validator.SchemaValidator
		mockFunc  func()
	}
	type args struct {
		jsonSchema string
	}

	mJSONValidator := new(mockedJSONValidator)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "missing last response body", fields: fields{
			resp:      nil,
			validator: mJSONValidator,
			mockFunc: func() {

			},
		}, args: args{jsonSchema: ""}, wantErr: true},
		{name: "validator fails", fields: fields{
			resp:      &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(""))},
			validator: mJSONValidator,
			mockFunc: func() {
				mJSONValidator.On("Validate", "", "").Return(errors.New("abc")).Once()
			},
		}, args: args{jsonSchema: ""}, wantErr: true},
		{name: "validator succeeded", fields: fields{
			resp:      &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(""))},
			validator: mJSONValidator,
			mockFunc: func() {
				mJSONValidator.On("Validate", "", "").Return(nil).Once()
			},
		}, args: args{jsonSchema: ""}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")
			s.SetSchemaStringValidator(tt.fields.validator)

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.resp)

			tt.fields.mockFunc()

			if err := s.IValidateLastResponseBodyWithSchemaString(tt.args.jsonSchema); (err != nil) != tt.wantErr {
				t.Errorf("IValidateLastResponseBodyWithSchemaReference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_IValidateNodeWithSchemaString(t *testing.T) {
	jsonData := `{
	"count": 2,
	"data": [
		{
			"name": "a",
			"age": 2
		},
		{
			"name": "b",
			"age": 3
		}
	]
}`

	yamlData := `---
user:
   name: four
   age: 4
`

	type fields struct {
		response *http.Response
	}

	type args struct {
		df         format.DataFormat
		expr       string
		jsonSchema string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//unsuccessful examples:
		{name: "no response body", fields: fields{response: nil}, args: args{
			df:         format.JSON,
			expr:       "",
			jsonSchema: "",
		}, wantErr: true},
		{name: "resolver could not resolve json path", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{}`))}}, args: args{
			df:         format.JSON,
			expr:       "",
			jsonSchema: "",
		}, wantErr: true},
		{name: "invalid json schema", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(jsonData))}}, args: args{
			df:   format.JSON,
			expr: "data",
			jsonSchema: `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "users data",
    "description": "users data",
    "type": "object",
    "properties": {
		"not_existing_key": {
			"type": "string"
		}
    }
}`,
		}, wantErr: true},

		// Successful examples:
		{name: "node value is valid according to provided JSON schema #1", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(jsonData))}}, args: args{
			df:   format.JSON,
			expr: "data",
			jsonSchema: `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "users data",
    "description": "users data",
    "type": "array",
    "items": {
        "type": "object",
        "properties": {
            "name": {
                "type": "string"
            },
			"age": {
                "type": "integer"
            }
        }
    }
}`,
		}, wantErr: false},
		{name: "node value is valid according to provided JSON schema #2", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(jsonData))}}, args: args{
			df:   format.JSON,
			expr: "count",
			jsonSchema: `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "count integer",
    "description": "number of items in data array",
    "type": "integer"
}`,
		}, wantErr: false},

		{name: "yaml node value is valid according to provided JSON schema #1", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(yamlData))}}, args: args{
			df:   format.YAML,
			expr: "$.user.name",
			jsonSchema: `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "user name",
    "description": "name of user",
    "type": "string"
}`,
		}, wantErr: false},
		{name: "yaml node value is valid according to provided JSON schema #2", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(yamlData))}}, args: args{
			df:   format.YAML,
			expr: "$.user",
			jsonSchema: `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "user",
    "description": "user data",
    "type": "object",
	"properties": {
		"name": {
			"type": "string"
		},
		"age": {
			"type": "integer"
		}
	}
}`,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.response)

			if err := s.IValidateNodeWithSchemaString(tt.args.df, tt.args.expr, tt.args.jsonSchema); (err != nil) != tt.wantErr {
				t.Errorf("IValidateJSONNodeWithSchemaString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TimeBetweenLastHTTPRequestResponseShouldBeLessThan(t *testing.T) {
	type fields struct {
		req *time.Time
		res *time.Time
	}
	type args struct {
		timeInterval string
	}

	currTime := time.Now()
	currTimePlusOneSec := currTime.Add(1 * time.Second)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "time passed between request and response is greater than expected", fields: fields{
			req: &currTime,
			res: &currTimePlusOneSec,
		}, args: args{timeInterval: "1ms"}, wantErr: true},
		{name: "time passed between request and response is equal to expected", fields: fields{
			req: &currTime,
			res: &currTimePlusOneSec,
		}, args: args{timeInterval: "1s"}, wantErr: false},
		{name: "time passed between request and response is less to expected", fields: fields{
			req: &currTime,
			res: &currTimePlusOneSec,
		}, args: args{timeInterval: "2s"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPRequestTimestamp, *tt.fields.req)
			s.Cache.Save(httpcache.LastHTTPResponseTimestamp, *tt.fields.res)

			td, err := time.ParseDuration(tt.args.timeInterval)
			if err != nil {
				t.Errorf("could not parse timeInterval: %s", tt.args.timeInterval)
			}

			if err := s.TimeBetweenLastHTTPRequestResponseShouldBeLessThanOrEqualTo(td); (err != nil) != tt.wantErr {
				t.Errorf("TimeBetweenLastHTTPRequestResponseShouldBeLessThan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheResponseShouldHaveCookieOfValue(t *testing.T) {
	mTemplateEngine := new(mockedTemplateEngine)

	type fields struct {
		TemplateEngine template.Engine
		response       *http.Response
		mockFunc       func()
	}
	type args struct {
		name          string
		valueTemplate string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "missing last response", fields: fields{
			TemplateEngine: mTemplateEngine,
			response:       nil,
			mockFunc:       func() {},
		}, args: args{
			name:          "",
			valueTemplate: "",
		}, wantErr: true},
		{name: "template engine returns error", fields: fields{
			TemplateEngine: mTemplateEngine,
			response:       &http.Response{},
			mockFunc: func() {
				mTemplateEngine.On("Replace", "a", mock.Anything).Return("", errors.New("abc")).Once()
			},
		}, args: args{
			name:          "",
			valueTemplate: "a",
		}, wantErr: true},
		{name: "cookies are not found in response", fields: fields{
			TemplateEngine: mTemplateEngine,
			response:       &http.Response{},
			mockFunc: func() {
				mTemplateEngine.On("Replace", "a", mock.Anything).Return("a", nil).Once()
			},
		}, args: args{
			name:          "a",
			valueTemplate: "a",
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")
			s.SetTemplateEngine(tt.fields.TemplateEngine)
			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.response)

			tt.fields.mockFunc()

			if err := s.TheResponseShouldHaveCookieOfValue(tt.args.name, tt.args.valueTemplate); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldHaveCookieOfValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_TheResponseShouldHaveCookie(t *testing.T) {
	mTemplateEngine := new(mockedTemplateEngine)

	type fields struct {
		TemplateEngine template.Engine
		response       *http.Response
		mockFunc       func()
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "missing last response", fields: fields{
			TemplateEngine: mTemplateEngine,
			response:       nil,
			mockFunc:       func() {},
		}, args: args{
			name: "",
		}, wantErr: true},
		{name: "template engine returns error", fields: fields{
			TemplateEngine: mTemplateEngine,
			response:       &http.Response{},
			mockFunc: func() {
				mTemplateEngine.On("Replace", "a", mock.Anything).Return("", errors.New("abc")).Once()
			},
		}, args: args{
			name: "",
		}, wantErr: true},
		{name: "cookies are not found in response", fields: fields{
			TemplateEngine: mTemplateEngine,
			response:       &http.Response{},
			mockFunc: func() {
				mTemplateEngine.On("Replace", "a", mock.Anything).Return("a", nil).Once()
			},
		}, args: args{
			name: "a",
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")
			s.SetTemplateEngine(tt.fields.TemplateEngine)
			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.response)

			tt.fields.mockFunc()

			if err := s.TheResponseShouldHaveCookie(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("TheResponseShouldHaveCookie() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_ISaveAs(t *testing.T) {
	type fields struct {
		cacheData map[string]any
	}
	type args struct {
		value    string
		cacheKey string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    string
	}{
		{name: "value and cacheKey should be not empty string", fields: fields{}, args: args{
			value:    "",
			cacheKey: "",
		}, wantErr: true},
		{name: "value should be not empty string", fields: fields{}, args: args{
			value:    "",
			cacheKey: "a",
		}, wantErr: true},
		{name: "cacheKey should be not empty string", fields: fields{}, args: args{
			value:    "a",
			cacheKey: "",
		}, wantErr: true},
		{name: "valid value", fields: fields{}, args: args{
			value:    "a",
			cacheKey: "a",
		}, wantErr: false, want: "a"},
		{name: "template value", fields: fields{
			cacheData: map[string]any{
				"FIRST_NAME": "ABC",
				"LAST_NAME":  "XXX",
			},
		}, args: args{
			value:    "{{.FIRST_NAME}} {{.LAST_NAME}}",
			cacheKey: "a",
		}, wantErr: false, want: "ABC XXX"},
		{name: "template value", fields: fields{
			cacheData: map[string]any{
				"HEIGHT": 10,
			},
		}, args: args{
			value:    "my height is {{.HEIGHT}}",
			cacheKey: "a",
		}, wantErr: false, want: "my height is 10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			for k, v := range tt.fields.cacheData {
				s.Cache.Save(k, v)
			}

			if err := s.ISaveAs(tt.args.value, tt.args.cacheKey); (err != nil) != tt.wantErr {
				t.Errorf("ISaveAs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				v, err := s.Cache.GetSaved(tt.args.cacheKey)
				if err != nil {
					t.Errorf("%s", err)
				}

				vStr, ok := v.(string)
				if !ok {
					t.Errorf("%+v value is not string", v)
				}

				if vStr != tt.want {
					t.Errorf("expected %s, got %s", tt.want, vStr)
				}
			}
		})
	}
}

func TestState_ISaveFromTheLastResponseNodeAs(t *testing.T) {
	type fields struct {
		cache        cache.Cache
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		df           format.DataFormat
		node         string
		variableName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "invalid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{df: format.JSON, node: "token", variableName: "TOKEN"}, wantErr: true},
		{name: "invalid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{df: format.JSON, node: "last_name", variableName: "LAST_NAME"}, wantErr: true},
		{name: "valid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{df: format.JSON, node: "user.last_name", variableName: "LAST_NAME"}, wantErr: false},
		{name: "valid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{df: format.JSON, node: "$.user", variableName: "USER"}, wantErr: false},

		//YAML
		{name: "invalid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{df: format.YAML, node: "$.token", variableName: "TOKEN"}, wantErr: true},
		{name: "invalid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
---
user:
   name: four
   last_name: b
`))},
		}, args: args{df: format.YAML, node: "$.last_name", variableName: "LAST_NAME"}, wantErr: true},
		{name: "valid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
---
user:
   name: four
   last_name: b
`))},
		}, args: args{df: format.YAML, node: "$.user.last_name", variableName: "LAST_NAME"}, wantErr: false},
		{name: "valid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
---
user:
   name: four
   last_name: b
`))},
		}, args: args{df: format.YAML, node: "$.user", variableName: "USER"}, wantErr: false},

		//XML
		{name: "invalid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<user>
	<name>abc</name>
	<lastName>xxx</lastName>
</user>`))},
		}, args: args{df: format.XML, node: "//token", variableName: "TOKEN"}, wantErr: true},
		{name: "invalid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<user>
	<name>abc</name>
	<lastName>xxx</lastName>
</user>`))},
		}, args: args{df: format.XML, node: "//lastname", variableName: "LAST_NAME"}, wantErr: true},
		{name: "valid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<user>
	<name>abc</name>
	<lastName>xxx</lastName>
</user>`))},
		}, args: args{df: format.XML, node: "//lastName", variableName: "LAST_NAME"}, wantErr: false},
		{name: "valid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<user>
	<name>abc</name>
	<lastName>xxx</lastName>
</user>`))},
		}, args: args{df: format.XML, node: "//user[1]//name", variableName: "USER"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(tt.fields.isDebug, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			err := s.ISaveFromTheLastResponseNodeAs(tt.args.df, tt.args.node, tt.args.variableName)

			if (err != nil) != tt.wantErr {
				t.Errorf("ISaveFromTheLastResponseJSONNodeAs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if _, err := s.Cache.GetSaved(tt.args.variableName); err != nil {
					t.Errorf("%s was not saved to Cache", tt.args.node)
				}
			}
		})
	}
}
