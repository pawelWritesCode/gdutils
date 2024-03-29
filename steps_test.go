package gdutils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pawelWritesCode/charset"
	"github.com/pawelWritesCode/df"
	"github.com/stretchr/testify/mock"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/httpcache"
	"github.com/pawelWritesCode/gdutils/pkg/mathutils"
	"github.com/pawelWritesCode/gdutils/pkg/timeutils"
	"github.com/pawelWritesCode/gdutils/pkg/types"
	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

var bigJson = `{
	"number_1" : 210,
	"number_2" : -210,
	"number_3" : 21.05,
	"number_4" : 1.0E+2,
	"color": "purple",
	"isHorizontal": false,
	"isPopular": null,
	"user": {
		"name": "John",
		"height": 180,
		"isActive": true,
		"address": null,
		"roles": ["role_admin", "role_user"]
	}
}`

var bigYaml = `---
 doe: "a deer, a female deer"
 ray: null
 nakamoto: ~
 pi: 3.14159
 xmas: true
 french-hens: 3
 calling-birds:
   - huey
   - dewey
   - louie
   - fred
 xmas-fifth-day:
   calling-birds: four
   french-hens: 3
   golden-rings: 5
   partridges:
     count: 1
     location: "a pear tree"
   turtle-doves: two`

var bigXML = `<?xml version="1.0"?>
<weather xmlns="x-schema:weatherSchema.xml">
  <description>Weather info</description>
  <date>1970-09-30</date>
  <degrees>67.5</degrees>
  <humidity>95</humidity>
  <isHot>true</isHot>
</weather>
`

var bigHTML = `<html>
  <head>
    <title>Div Align Attribbute</title>
  </head>
  <body>
    <div align="left">
      a
    </div>
    <div align="right">
      b
    </div>
    <div align="center">
      c
    </div>
    <div align="justify">
      d
    </div>
	<div>
		<p>Hello world</p>
	</div>
	<div>
		<i>1</i>
		<i>2</i>
        <i>3</i>
	</div>
  </body>
</html>
`

func createRespBody() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(bigJson))
}

func createYamlRespBody() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(bigYaml))
}

func createXMLRespBody() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(bigXML))
}

func createHTMLRespBody() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(bigHTML))
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

func (m *mockedJsonPathFinder) Find(expr string, jsonBytes []byte) (any, error) {
	args := m.Called(expr, jsonBytes)

	return args.Get(0).(any), args.Error(1)
}

func TestState_RequestPrepare(t *testing.T) {
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
			if err := s.RequestPrepare(tt.args.method, tt.args.urlTemplate, tt.args.cacheKey); (err != nil) != tt.wantErr {
				t.Errorf("RequestPrepare() error = %v, wantErr %v", err, tt.wantErr)
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

func TestState_RequestSetHeaders(t *testing.T) {
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

			err := s.RequestPrepare(tt.fields.reqMethod, tt.fields.reqUri, tt.fields.cacheKey)
			if err != nil {
				t.Errorf("%v", err)
			}

			if err = s.RequestSetHeaders(tt.args.cacheKey, tt.args.headersTemplate); (err != nil) != tt.wantErr {
				t.Errorf("RequestSetHeaders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_RequestSetBody(t *testing.T) {
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
			err := s.RequestPrepare(tt.fields.reqMethod, tt.fields.reqUri, tt.fields.cacheKey)
			if err != nil {
				t.Errorf("%v", err)
			}

			if err := s.RequestSetBody(tt.args.cacheKey, tt.args.bodyTemplate); (err != nil) != tt.wantErr {
				t.Errorf("RequestSetBody() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_RequestSetCookies(t *testing.T) {
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
				s.SetJSONSerializer(mFormatter)
				s.SetTemplateEngine(mTemplateEngine)
			}
			tt.fields.mockFunc()

			if err = s.RequestPrepare("GET", "https://www.example.com", tt.args.cacheKey); err != nil {
				t.Errorf("%s", err.Error())
			}

			s.Cache.Save("NOW", tm)

			if err := s.RequestSetCookies(tt.args.cacheKey, tt.args.cookiesTemplate); (err != nil) != tt.wantErr {
				t.Errorf("RequestSetCookies() error = %v, wantErr %v", err, tt.wantErr)
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

func TestState_AssertStatusCodeIs(t *testing.T) {
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

			if err := s.AssertStatusCodeIs(tt.args.code); (err != nil) != tt.wantErr {
				t.Errorf("AssertStatusCodeIs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertStatusCodeIs() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{StatusCode: 201})

	err := apiCtx.AssertStatusCodeIs(200)
	fmt.Println(err)

	// Output:
	// expected status code 200, but got 201
}

func TestAPIContext_AssertStatusCodeIsNot(t *testing.T) {
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
		// Successful
		{name: "different code #1, code less than 200", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 1}, wantErr: false},
		{name: "different code #2, code over 599", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 600}, wantErr: false},
		{name: "different code #3, different code expected", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 400}, wantErr: false},

		// Failed
		{name: "same code #1", fields: fields{lastResponse: &http.Response{StatusCode: 200}}, args: args{code: 200}, wantErr: true},
		{name: "same code #1", fields: fields{lastResponse: &http.Response{StatusCode: 400}}, args: args{code: 400}, wantErr: true},
		{name: "same code #1", fields: fields{lastResponse: &http.Response{StatusCode: 511}}, args: args{code: 511}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(tt.fields.isDebug, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)
			if err := s.AssertStatusCodeIsNot(tt.args.code); (err != nil) != tt.wantErr {
				t.Errorf("AssertStatusCodeIsNot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertStatusCodeIsNot() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{StatusCode: 200})

	err := apiCtx.AssertStatusCodeIsNot(200)
	fmt.Println(err)

	// Output:
	// expected status code different than 200, but got 200
}

func TestState_AssertResponseFormatIs(t *testing.T) {
	yaml := `
---
user:
   name: four
   last_name: b
`

	xml := `<?xml version="1.0" encoding="UTF-8"?><this>is xml</this>`
	plainText := `this is plain text`
	json := `{"this_is": "json"}`
	html := `<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Img Width Attribute</title>
  </head>
  <body>
    <img src="image.png" alt="Image" width="100" />
  </body>
</html>
`
	type fields struct {
		body []byte
	}

	type args struct {
		dataFormat df.DataFormat
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		// JSON
		{name: "response body has JSON format and JSON format expected", args: args{dataFormat: df.JSON}, fields: fields{body: []byte(json)}, wantErr: false},
		{name: "response body has plain text format but JSON format expected", args: args{dataFormat: df.JSON}, fields: fields{body: []byte(plainText)}, wantErr: true},
		{name: "response body has XML format but JSON format expected", args: args{dataFormat: df.JSON}, fields: fields{body: []byte(xml)}, wantErr: true},
		{name: "response body has YAML format but JSON format expected", args: args{dataFormat: df.JSON}, fields: fields{body: []byte(yaml)}, wantErr: true},
		{name: "response body has HTML format but JSON format expected", args: args{dataFormat: df.JSON}, fields: fields{body: []byte(html)}, wantErr: true},

		// YAML
		{name: "response body has JSON format but YAML format expected", args: args{dataFormat: df.YAML}, fields: fields{body: []byte(json)}, wantErr: true},
		{name: "response body has plain text format but YAML format expected", args: args{dataFormat: df.YAML}, fields: fields{body: []byte(plainText)}, wantErr: true},
		{name: "response body has XML format but YAML format expected", args: args{dataFormat: df.YAML}, fields: fields{body: []byte(xml)}, wantErr: true},
		{name: "response body has YAML format and YAML format expected", args: args{dataFormat: df.YAML}, fields: fields{body: []byte(yaml)}, wantErr: false},
		{name: "response body has HTML format and YAML format expected", args: args{dataFormat: df.YAML}, fields: fields{body: []byte(html)}, wantErr: true},

		// XML
		{name: "response body has JSON format but XML format expected", args: args{dataFormat: df.XML}, fields: fields{body: []byte(json)}, wantErr: true},
		{name: "response body has plain text format but XML format expected", args: args{dataFormat: df.XML}, fields: fields{body: []byte(plainText)}, wantErr: true},
		{name: "response body has XML format and XML format expected", args: args{dataFormat: df.XML}, fields: fields{body: []byte(xml)}, wantErr: false},
		{name: "response body has YAML format but XML format expected", args: args{dataFormat: df.XML}, fields: fields{body: []byte(yaml)}, wantErr: true},

		// plain text
		{name: "response body has JSON format but plain text format expected", args: args{dataFormat: df.PlainText}, fields: fields{body: []byte(json)}, wantErr: true},
		{name: "response body has plain text format and plain text format expected", args: args{dataFormat: df.PlainText}, fields: fields{body: []byte(plainText)}, wantErr: false},
		{name: "response body has XML format but plain text format expected", args: args{dataFormat: df.PlainText}, fields: fields{body: []byte(xml)}, wantErr: false},
		{name: "response body has YAML format but plain text format expected", args: args{dataFormat: df.PlainText}, fields: fields{body: []byte(yaml)}, wantErr: false},

		//HTML
		{name: "response body has HTML format and html format expected", args: args{dataFormat: df.HTML}, fields: fields{body: []byte(html)}, wantErr: false},
		{name: "response body has JSON format but HTML format expected", args: args{dataFormat: df.HTML}, fields: fields{body: []byte(json)}, wantErr: true},
		{name: "response body has plain text format but HTML format expected", args: args{dataFormat: df.HTML}, fields: fields{body: []byte(plainText)}, wantErr: true},
		{name: "response body has YAML format but HTML format expected", args: args{dataFormat: df.HTML}, fields: fields{body: []byte(yaml)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})

			if err := s.AssertResponseFormatIs(tt.args.dataFormat); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseFormatIs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertResponseFormatIs() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader("abc"))})

	err := apiCtx.AssertResponseFormatIs(df.JSON)
	fmt.Println(err)

	// Output:
	// response body doesn't have format json
}

func TestAPIContext_AssertResponseFormatIsNot(t *testing.T) {
	yaml := `
---
user:
   name: four
   last_name: b
`

	xml := `<this>is xml</this>`
	plainText := `this is plain text`
	json := `{"this_is": "json"}`
	html := `<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Img Width Attribute</title>
  </head>
  <body>
    <img src="image.png" alt="Image" width="100" />
  </body>
</html>
`

	type fields struct {
		body []byte
	}
	type args struct {
		dataFormat df.DataFormat
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// JSON
		{name: "response body has JSON format and JSON format is not expected", args: args{dataFormat: df.JSON}, fields: fields{body: []byte(json)}, wantErr: true},
		{name: "response body has plain text format but JSON format is not expected", args: args{dataFormat: df.JSON}, fields: fields{body: []byte(plainText)}, wantErr: false},
		{name: "response body has XML format but JSON format is not expected", args: args{dataFormat: df.JSON}, fields: fields{body: []byte(xml)}, wantErr: false},
		{name: "response body has YAML format but JSON format is not expected", args: args{dataFormat: df.JSON}, fields: fields{body: []byte(yaml)}, wantErr: false},

		// YAML
		{name: "response body has JSON format but YAML format is not expected", args: args{dataFormat: df.YAML}, fields: fields{body: []byte(json)}, wantErr: false},
		{name: "response body has plain text format but YAML format is not expected", args: args{dataFormat: df.YAML}, fields: fields{body: []byte(plainText)}, wantErr: false},
		{name: "response body has XML format but YAML format is not expected", args: args{dataFormat: df.YAML}, fields: fields{body: []byte(xml)}, wantErr: false},
		{name: "response body has YAML format and YAML format is not expected", args: args{dataFormat: df.YAML}, fields: fields{body: []byte(yaml)}, wantErr: true},

		// XML
		{name: "response body has JSON format but XML format is not expected", args: args{dataFormat: df.XML}, fields: fields{body: []byte(json)}, wantErr: false},
		{name: "response body has plain text format but XML format is not expected", args: args{dataFormat: df.XML}, fields: fields{body: []byte(plainText)}, wantErr: false},
		{name: "response body has XML format and XML format is not expected", args: args{dataFormat: df.XML}, fields: fields{body: []byte(xml)}, wantErr: true},
		{name: "response body has YAML format but XML format is not expected", args: args{dataFormat: df.XML}, fields: fields{body: []byte(yaml)}, wantErr: false},

		// plain text
		{name: "response body has JSON format but plain text format is not expected", args: args{dataFormat: df.PlainText}, fields: fields{body: []byte(json)}, wantErr: false},
		{name: "response body has plain text format and plain text format is not expected", args: args{dataFormat: df.PlainText}, fields: fields{body: []byte(plainText)}, wantErr: true},
		{name: "response body has XML format but plain text format is not expected", args: args{dataFormat: df.PlainText}, fields: fields{body: []byte(xml)}, wantErr: true},
		{name: "response body has YAML format but plain text format is not expected", args: args{dataFormat: df.PlainText}, fields: fields{body: []byte(yaml)}, wantErr: true},

		//HTML
		{name: "response body has HTML format and html format expected", args: args{dataFormat: df.HTML}, fields: fields{body: []byte(html)}, wantErr: true},
		{name: "response body has JSON format but HTML format expected", args: args{dataFormat: df.HTML}, fields: fields{body: []byte(json)}, wantErr: false},
		{name: "response body has plain text format but HTML format expected", args: args{dataFormat: df.HTML}, fields: fields{body: []byte(plainText)}, wantErr: false},
		{name: "response body has YAML format but HTML format expected", args: args{dataFormat: df.HTML}, fields: fields{body: []byte(yaml)}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})
			if err := s.AssertResponseFormatIsNot(tt.args.dataFormat); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseFormatIsNot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertResponseFormatIsNot() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader("{}"))})

	err := apiCtx.AssertResponseFormatIsNot(df.JSON)
	fmt.Println(err)

	// Output:
	// response body has format json
}

func TestState_GenerateRandomInt(t *testing.T) {
	s := NewDefaultAPIContext(false, "")
	for i := 0; i < 100; i++ {
		if err := s.GenerateRandomInt(0, 100000, "RANDOM_INT"); (err != nil) != false {
			t.Errorf("GenerateRandomInt() error = %v, wantErr %v", err, false)
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

func ExampleAPIContext_GenerateRandomInt() {
	apiCtx := NewDefaultAPIContext(false, "")

	if err := apiCtx.GenerateRandomInt(2, 2, "MYINT"); err != nil {
		log.Fatal(err)

		return
	}

	myInt, err := apiCtx.Cache.GetSaved("MYINT")
	if err != nil {
		log.Fatal(err)

		return
	}

	fmt.Println(myInt)
	// Output:
	// 2
}

func ExampleAPIContext_GenerateFloat64() {
	apiCtx := NewDefaultAPIContext(false, "")

	if err := apiCtx.GenerateFloat64(0, 0, "MYFLOAT"); err != nil {
		log.Fatal(err)

		return
	}

	myFloat, err := apiCtx.Cache.GetSaved("MYFLOAT")
	if err != nil {
		log.Fatal(err)

		return
	}

	fmt.Println(myFloat)
	// Output:
	// 0
}

func TestState_GeneratorRandomRunes(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	rndStringASCII := s.GeneratorRandomRunes(charset.ASCII)
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

	rndStringUnicode := s.GeneratorRandomRunes(charset.Unicode)
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

func ExampleAPIContext_GeneratorRandomRunes() {
	apiCtx := NewDefaultAPIContext(false, "")

	generateA := apiCtx.GeneratorRandomRunes("A")
	err := generateA(2, 2, "TWO_A")
	if err != nil {
		log.Fatal(err)

		return
	}

	myA, err := apiCtx.Cache.GetSaved("TWO_A")
	if err != nil {
		log.Fatal(err)

		return
	}

	fmt.Println(myA)
	// Output:
	// AA
}

func TestState_GeneratorRandomSentence_ASCII(t *testing.T) {
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

func ExampleAPIContext_GeneratorRandomSentence() {
	apiCtx := NewDefaultAPIContext(false, "")

	generateA := apiCtx.GeneratorRandomSentence("A", 2, 2)
	err := generateA(2, 2, "TWO_WORDS_A")
	if err != nil {
		log.Fatal(err)

		return
	}

	myA, err := apiCtx.Cache.GetSaved("TWO_WORDS_A")
	if err != nil {
		log.Fatal(err)

		return
	}

	fmt.Println(myA)
	// Output:
	// AA AA
}

func TestState_GeneratorRandomSentence_Unicode(t *testing.T) {
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

func TestState_GetTimeAndTravel(t *testing.T) {
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

			if err := s.GetTimeAndTravel(tt.args.t, tt.args.timeDirection, tt.args.timeDuration, tt.args.cacheKey); (err != nil) != tt.wantErr {
				t.Errorf("GetTimeAndTravel() error = %v, wantErr %v", err, tt.wantErr)
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

func TestState_AssertNodeExists(t *testing.T) {
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

	html := `<html>
  <head>
    <title>Div Align Attribbute</title>
  </head>
  <body>
    <div align="left">
      a
    </div>
    <div align="right">
      b
    </div>
    <div align="center">
      c
    </div>
    <div align="justify">
      d
    </div>
	<div>
		<p>Hello world</p>
	</div>
	<div>
		<i>1</i>
		<i>2</i>
        <i>3</i>
	</div>
  </body>
</html>
`

	type fields struct {
		body      []byte
		cacheData map[string]any
	}
	type args struct {
		dataFormat  df.DataFormat
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
			dataFormat:  df.JSON,
			expressions: "age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - gjson expression", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "users",
		}, wantErr: false},
		{name: "last response body has expected key #2 - Oliveagle expression", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "$.users[0]",
		}, wantErr: false},
		{name: "last response body has expected key #3 - gjson expression ", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "users.1.name",
		}, wantErr: false},
		{name: "last response body has expected key #4 - gjson expression with template value", fields: fields{body: []byte(json), cacheData: map[string]any{
			"USER_ID": 1,
		}}, args: args{
			dataFormat:  df.JSON,
			expressions: "users.{{.USER_ID}}.name",
		}, wantErr: false},

		// XML
		{name: "last response body is missing expected key in XML", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//users",
		}, wantErr: false},
		{name: "last response body has expected key #2 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//users//user[1]",
		}, wantErr: false},
		{name: "last response body has expected key #3 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//user[2]//name",
		}, wantErr: false},
		{name: "last response body has expected key #4 - Antchfx expression with template value", fields: fields{body: []byte(xml), cacheData: map[string]any{
			"USER_ID": 2,
		}}, args: args{
			dataFormat:  df.XML,
			expressions: "//user[{{.USER_ID}}]//name",
		}, wantErr: false},

		// YAML
		{name: "last response body is missing expected key in YAML", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - Goccy expression", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users",
		}, wantErr: false},
		{name: "last response body has expected key #2 - Goccy expression", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users[0].name",
		}, wantErr: false},
		{name: "last response body has expected key #3 - Goccy expression with template value", fields: fields{body: []byte(yaml), cacheData: map[string]any{
			"USER_ID": 0,
		}}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users[{{.USER_ID}}].name",
		}, wantErr: false},

		{name: "last response body is missing expected node in HTML", fields: fields{body: []byte(html)}, args: args{
			dataFormat:  df.HTML,
			expressions: "//strong",
		}, wantErr: true},
		{name: "last response body has expected node #1 in HTML", fields: fields{body: []byte(html)}, args: args{
			dataFormat:  df.HTML,
			expressions: "//i",
		}, wantErr: false},
		{name: "last response body has expected node #2 in HTML", fields: fields{body: []byte(html)}, args: args{
			dataFormat:  df.HTML,
			expressions: "//div[@align=\"justify\"]",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			for k, v := range tt.fields.cacheData {
				s.Cache.Save(k, v)
			}

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})

			if err := s.AssertNodeExists(tt.args.dataFormat, tt.args.expressions); (err != nil) != tt.wantErr {
				t.Errorf("AssertNodeExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeExists() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader("{}"))})

	err := apiCtx.AssertNodeExists(df.JSON, "user")
	fmt.Println(err)

	// Output:
	// node 'user' could not be found within last response body, reason: could not find node using provided expression: 'user', err: could not find node, using expression user
}

func TestAPIContext_AssertNodeNotExists(t *testing.T) {
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

	html := `<html>
  <head>
    <title>Div Align Attribbute</title>
  </head>
  <body>
    <div align="left">
      a
    </div>
    <div align="right">
      b
    </div>
    <div align="center">
      c
    </div>
    <div align="justify">
      d
    </div>
	<div>
		<p>Hello world</p>
	</div>
	<div>
		<i>1</i>
		<i>2</i>
        <i>3</i>
	</div>
  </body>
</html>
`

	type fields struct {
		body      []byte
		cacheData map[string]any
	}
	type args struct {
		dataFormat  df.DataFormat
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
			dataFormat:  df.JSON,
			expressions: "age",
		}, wantErr: false},
		{name: "last response body has expected key #1 - gjson expression", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "users",
		}, wantErr: true},
		{name: "last response body has expected key #2 - Oliveagle expression", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "$.users[0]",
		}, wantErr: true},
		{name: "last response body has expected key #3 - gjson expression ", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "users.1.name",
		}, wantErr: true},
		{name: "last response body has expected key #4 - gjson expression with template value", fields: fields{body: []byte(json), cacheData: map[string]any{
			"USER_ID": 1,
		}}, args: args{
			dataFormat:  df.JSON,
			expressions: "users.{{.USER_ID}}.name",
		}, wantErr: true},

		// XML
		{name: "last response body is missing expected key in XML", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "age",
		}, wantErr: false},
		{name: "last response body has expected key #1 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//users",
		}, wantErr: true},
		{name: "last response body has expected key #2 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//users//user[1]",
		}, wantErr: true},
		{name: "last response body has expected key #3 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//user[2]//name",
		}, wantErr: true},
		{name: "last response body has expected key #4 - Antchfx expression with template value", fields: fields{body: []byte(xml), cacheData: map[string]any{
			"USER_ID": 2,
		}}, args: args{
			dataFormat:  df.XML,
			expressions: "//user[{{.USER_ID}}]//name",
		}, wantErr: true},

		// YAML
		{name: "last response body is missing expected key in YAML", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.age",
		}, wantErr: false},
		{name: "last response body has expected key #1 - Goccy expression", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users",
		}, wantErr: true},
		{name: "last response body has expected key #2 - Goccy expression", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users[0].name",
		}, wantErr: true},
		{name: "last response body has expected key #3 - Goccy expression with template value", fields: fields{body: []byte(yaml), cacheData: map[string]any{
			"USER_ID": 0,
		}}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users[{{.USER_ID}}].name",
		}, wantErr: true},

		// HTML
		{name: "last response body is missing expected node in HTML", fields: fields{body: []byte(html)}, args: args{
			dataFormat:  df.HTML,
			expressions: "//strong",
		}, wantErr: false},
		{name: "last response body has expected node #1 in HTML", fields: fields{body: []byte(html)}, args: args{
			dataFormat:  df.HTML,
			expressions: "//i",
		}, wantErr: true},
		{name: "last response body has expected node #2 in HTML", fields: fields{body: []byte(html)}, args: args{
			dataFormat:  df.HTML,
			expressions: "//div[@align=\"justify\"]",
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			for k, v := range tt.fields.cacheData {
				s.Cache.Save(k, v)
			}

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})
			if err := s.AssertNodeNotExists(tt.args.dataFormat, tt.args.expressions); (err != nil) != tt.wantErr {
				t.Errorf("AssertNodeNotExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeNotExists() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user": "abc"}`))})

	err := apiCtx.AssertNodeNotExists(df.JSON, "user")
	fmt.Println(err)

	// Output:
	// json node 'user' exists
}

func TestState_AssertNodesExist(t *testing.T) {
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
		dataFormat  df.DataFormat
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
			dataFormat:  df.JSON,
			expressions: "age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - gjson expression", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "users",
		}, wantErr: false},
		{name: "last response body has expected keys #2 - different expressions", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "users, $.users",
		}, wantErr: false},
		{name: "last response body has expected keys #3 - gjson expressions", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "users, users.0",
		}, wantErr: false},
		{name: "last response body has expected keys #4 - gjson expressions", fields: fields{body: []byte(json)}, args: args{
			dataFormat:  df.JSON,
			expressions: "users.0.name, users.1.name",
		}, wantErr: false},
		{name: "last response body has expected keys #1 - gjson expressions with template values", fields: fields{body: []byte(json), cacheKeys: map[string]any{
			"USER1_ID": 0,
			"USER2_ID": 1,
		}}, args: args{
			dataFormat:  df.JSON,
			expressions: "users.{{.USER1_ID}}.name, users.{{.USER2_ID}}.name",
		}, wantErr: false},

		// XML
		{name: "last response body is missing key", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - Antchfx expression", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//users",
		}, wantErr: false},
		{name: "last response body has expected keys #2 - Antchfx expressions", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//users, //users",
		}, wantErr: false},
		{name: "last response body has expected keys #3 - Antchfx expressions", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//users, //users//user[1]",
		}, wantErr: false},
		{name: "last response body has expected keys #4 - Antchfx expressions", fields: fields{body: []byte(xml)}, args: args{
			dataFormat:  df.XML,
			expressions: "//user[1]//name, //user[2]//name",
		}, wantErr: false},
		{name: "last response body has expected keys #5 - Antchfx expressions with template values", fields: fields{body: []byte(xml), cacheKeys: map[string]any{
			"USER1_ID": 1,
			"USER2_ID": 2,
		}}, args: args{
			dataFormat:  df.XML,
			expressions: "//user[{{.USER1_ID}}]//name, //user[{{.USER2_ID}}]//name",
		}, wantErr: false},

		// YAML
		{name: "last response body is missing yaml key", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.age",
		}, wantErr: true},
		{name: "last response body has expected key #1 - Goccy expression", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users",
		}, wantErr: false},
		{name: "last response body has expected keys #2 - Goccy expressions", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users, $.users",
		}, wantErr: false},
		{name: "last response body has expected keys #3 - Goccy expressions", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users, $.users[0]",
		}, wantErr: false},
		{name: "last response body has expected keys #4 - Goccy expressions", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat:  df.YAML,
			expressions: "$.users[0].name, $.users[1].name",
		}, wantErr: false},
		{name: "last response body has expected keys #5 - Goccy expressions with template values", fields: fields{body: []byte(yaml), cacheKeys: map[string]any{
			"USER1_ID": 0,
			"USER2_ID": 1,
		}}, args: args{
			dataFormat:  df.YAML,
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

			if err := s.AssertNodesExist(tt.args.dataFormat, tt.args.expressions); (err != nil) != tt.wantErr {
				t.Errorf("AssertNodesExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodesExist() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user":"abc"}`))})

	err := apiCtx.AssertNodesExist(df.JSON, "user, address, roles")
	fmt.Println(err)

	// Output:
	// node 'address', err: could not find node using provided expression: 'address', err: could not find node, using expression address
	//node 'roles', err: could not find node using provided expression: 'roles', err: could not find node, using expression roles
}

func TestAPIContext_AssertNodeIsNotType(t *testing.T) {
	type fields struct {
		saved        map[string]any
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		df     df.DataFormat
		node   string
		goType types.DataType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// JSON related types against json data
		{
			name:    "selected node does not exists",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "abc", goType: types.Null},
			wantErr: true,
		},
		{
			name:    "selected node value is null - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isPopular", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is null - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isPopular", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is boolean - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isHorizontal", goType: types.Null},
			wantErr: false,
		},
		{
			name:    "selected node value is boolean - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isHorizontal", goType: types.Null},
			wantErr: false,
		},
		{
			name:    "selected node value is string - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "color", goType: types.Boolean},
			wantErr: false,
		},
		{
			name:    "selected node value is string - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.color", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #1 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_1", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is number #1 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_1", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is number #2 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_2", goType: types.Boolean},
			wantErr: false,
		},
		{
			name:    "selected node value is number #2 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_2", goType: types.Object},
			wantErr: false,
		},
		{
			name:    "selected node value is number #3 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_3", goType: types.Array},
			wantErr: false,
		},
		{
			name:    "selected node value is number #3 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_3", goType: types.Array},
			wantErr: false,
		},
		{
			name:    "selected node value is number #4 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_4", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is number #4 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_4", goType: types.Boolean},
			wantErr: false,
		},
		{
			name:    "selected node value is number #5 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user.height", goType: types.Array},
			wantErr: false,
		},
		{
			name:    "selected node value is number #5- - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user.height", goType: types.Object},
			wantErr: false,
		},
		{
			name:    "selected node value is object - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is object - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user", goType: types.Boolean},
			wantErr: false,
		},
		{
			name:    "selected node value is array - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user.roles", goType: types.Object},
			wantErr: false,
		},
		{
			name:    "selected node value is array - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user.roles", goType: types.Boolean},
			wantErr: false,
		},

		// YAML related data types against YAML data
		{
			name:    "selected node does not exist",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.abc", goType: types.Null},
			wantErr: true,
		},
		{
			name:    "selected node value is scalar - null #1",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.ray", goType: types.Sequence},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - null #2",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.nakamoto", goType: types.Mapping},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - string #1",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.doe", goType: types.Sequence},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - string #2",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas-fifth-day.calling-birds", goType: types.Sequence},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - float",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.pi", goType: types.Mapping},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - int #1",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.french-hens", goType: types.Mapping},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - int #2",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas-fifth-day.french-hens", goType: types.Mapping},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - boolean",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas", goType: types.Sequence},
			wantErr: false,
		},
		{
			name:    "selected node value is sequence",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.calling-birds", goType: types.Scalar},
			wantErr: false,
		},
		{
			name:    "selected node value is mapping",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas-fifth-day", goType: types.Scalar},
			wantErr: false,
		},

		//GO related types against JSON data
		{
			name:    "selected node value is nil - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isPopular", goType: types.Slice},
			wantErr: false,
		},
		{
			name:    "selected node value is nil - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isPopular", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is int #1 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_1", goType: types.Float},
			wantErr: false,
		},
		{
			name:    "selected node value is int #1 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_1", goType: types.Float},
			wantErr: false,
		},
		{
			name:    "selected node value is float - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_3", goType: types.Int},
			wantErr: false,
		},
		{
			name:    "selected node value is float - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_3", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is string - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "color", goType: types.Bool},
			wantErr: false,
		},
		{
			name:    "selected node value is string - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.color", goType: types.Int},
			wantErr: false,
		},
		{
			name:    "selected node value is bool - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isHorizontal", goType: types.Slice},
			wantErr: false,
		},
		{
			name:    "selected node value is bool - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isHorizontal", goType: types.Map},
			wantErr: false,
		},
		{
			name:    "selected node value is map - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is map - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user", goType: types.Slice},
			wantErr: false,
		},
		{
			name:    "selected node value is slice - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user.roles", goType: types.Bool},
			wantErr: false,
		},
		{
			name:    "selected node value is slice - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user.roles", goType: types.Float},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := NewDefaultAPIContext(tt.fields.isDebug, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := af.AssertNodeIsNotType(tt.args.df, tt.args.node, tt.args.goType); (err != nil) != tt.wantErr {
				t.Errorf("AssertNodeIsNotType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeIsNotType() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user":"abc"}`))})

	err := apiCtx.AssertNodeIsNotType(df.JSON, "user", types.String)
	fmt.Println(err)

	// Output:
	// node 'user' has type 'string', but expected not to be
}

func TestState_AssertNodeIsType(t *testing.T) {
	type fields struct {
		saved        map[string]any
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		df     df.DataFormat
		node   string
		goType types.DataType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// JSON related types against json data
		{
			name:    "selected node does not exists",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "abc", goType: types.Null},
			wantErr: true,
		},
		{
			name:    "selected node value is null - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isPopular", goType: types.Null},
			wantErr: false,
		},
		{
			name:    "selected node value is null - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isPopular", goType: types.Null},
			wantErr: false,
		},
		{
			name:    "selected node value is boolean - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isHorizontal", goType: types.Boolean},
			wantErr: false,
		},
		{
			name:    "selected node value is boolean - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isHorizontal", goType: types.Boolean},
			wantErr: false,
		},
		{
			name:    "selected node value is string - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "color", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is string - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.color", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is number #1 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_1", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #1 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_1", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #2 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_2", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #2 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_2", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #3 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_3", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #3 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_3", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #4 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_4", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #4 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_4", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #5 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user.height", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is number #5- - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user.height", goType: types.Number},
			wantErr: false,
		},
		{
			name:    "selected node value is object - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user", goType: types.Object},
			wantErr: false,
		},
		{
			name:    "selected node value is object - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user", goType: types.Object},
			wantErr: false,
		},
		{
			name:    "selected node value is array - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user.roles", goType: types.Array},
			wantErr: false,
		},
		{
			name:    "selected node value is array - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user.roles", goType: types.Array},
			wantErr: false,
		},

		// YAML related data types against YAML data
		{
			name:    "selected node does not exist",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.abc", goType: types.Null},
			wantErr: true,
		},
		{
			name:    "selected node value is scalar - null #1",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.ray", goType: types.Null},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - null #2",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.nakamoto", goType: types.Null},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - string #1",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.doe", goType: types.Scalar},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - string #2",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas-fifth-day.calling-birds", goType: types.Scalar},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - float",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.pi", goType: types.Scalar},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - int #1",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.french-hens", goType: types.Scalar},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - int #2",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas-fifth-day.french-hens", goType: types.Scalar},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - boolean",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas", goType: types.Scalar},
			wantErr: false,
		},
		{
			name:    "selected node value is sequence",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.calling-birds", goType: types.Sequence},
			wantErr: false,
		},
		{
			name:    "selected node value is mapping",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas-fifth-day", goType: types.Mapping},
			wantErr: false,
		},

		//GO related types against JSON data
		{
			name:    "selected node value is nil - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isPopular", goType: types.Nil},
			wantErr: false,
		},
		{
			name:    "selected node value is nil - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isPopular", goType: types.Nil},
			wantErr: false,
		},
		{
			name:    "selected node value is int #1 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_1", goType: types.Int},
			wantErr: false,
		},
		{
			name:    "selected node value is int #1 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_1", goType: types.Int},
			wantErr: false,
		},
		{
			name:    "selected node value is float - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_3", goType: types.Float},
			wantErr: false,
		},
		{
			name:    "selected node value is float - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_3", goType: types.Float},
			wantErr: false,
		},
		{
			name:    "selected node value is string - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "color", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is string - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.color", goType: types.String},
			wantErr: false,
		},
		{
			name:    "selected node value is bool - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isHorizontal", goType: types.Bool},
			wantErr: false,
		},
		{
			name:    "selected node value is bool - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isHorizontal", goType: types.Bool},
			wantErr: false,
		},
		{
			name:    "selected node value is map - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user", goType: types.Map},
			wantErr: false,
		},
		{
			name:    "selected node value is map - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user", goType: types.Map},
			wantErr: false,
		},
		{
			name:    "selected node value is slice - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user.roles", goType: types.Slice},
			wantErr: false,
		},
		{
			name:    "selected node value is slice - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user.roles", goType: types.Slice},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := NewDefaultAPIContext(tt.fields.isDebug, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := af.AssertNodeIsType(tt.args.df, tt.args.node, tt.args.goType); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeIsType() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user":"abc"}`))})

	err := apiCtx.AssertNodeIsType(df.JSON, "user", types.Number)
	fmt.Println(err)

	// Output:
	// expected node 'user' to be 'number', but node value is detected as 'string' in terms of JSON and 'string' in terms of Go
}

func TestState_AssertNodeSliceLengthIs(t *testing.T) {
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
		dataFormat df.DataFormat
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
			dataFormat: df.JSON,
			expr:       "count",
			length:     0,
		}, wantErr: true},
		{name: "expression does point at slice but expected invalid length", fields: fields{body: []byte(json)}, args: args{
			dataFormat: df.JSON,
			expr:       "data",
			length:     1,
		}, wantErr: true},
		{name: "expression does point at slice and expected proper length", fields: fields{body: []byte(json)}, args: args{
			dataFormat: df.JSON,
			expr:       "data",
			length:     2,
		}, wantErr: false},

		//YAML
		{name: "expression does not point at slice", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat: df.YAML,
			expr:       "$.count",
			length:     0,
		}, wantErr: true},
		{name: "expression does point at slice but expected invalid length", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat: df.YAML,
			expr:       "$.data",
			length:     1,
		}, wantErr: true},
		{name: "expression does point at slice and expected proper length", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat: df.YAML,
			expr:       "$.data",
			length:     2,
		}, wantErr: false},

		//XML
		{name: "expression does not point at slice", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: df.XML,
			expr:       "//count",
			length:     0,
		}, wantErr: true},
		{name: "expression does point at slice but expected invalid length", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: df.XML,
			expr:       "//data//name",
			length:     1,
		}, wantErr: true},
		{name: "expression does point at slice and expected proper length #1", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: df.XML,
			expr:       "//data//name",
			length:     2,
		}, wantErr: false},
		{name: "expression does point at slice and expected proper length #2", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: df.XML,
			expr:       "//user",
			length:     2,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})

			if err := s.AssertNodeSliceLengthIs(tt.args.dataFormat, tt.args.expr, tt.args.length); (err != nil) != tt.wantErr {
				t.Errorf("AssertNodeSliceLengthIs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeSliceLengthIs() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user":[]}`))})

	err := apiCtx.AssertNodeSliceLengthIs(df.JSON, "user", 1)
	fmt.Println(err)

	// Output:
	// node 'user' contains slice(array) which has length: 0, but expected: 1
}

func TestAPIContext_AssertNodeSliceLengthIsNot(t *testing.T) {
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
		dataFormat df.DataFormat
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
			dataFormat: df.JSON,
			expr:       "count",
			length:     0,
		}, wantErr: true},
		{name: "expression does point at slice and expected invalid length", fields: fields{body: []byte(json)}, args: args{
			dataFormat: df.JSON,
			expr:       "data",
			length:     1,
		}, wantErr: false},
		{name: "expression does point at slice but expected proper length", fields: fields{body: []byte(json)}, args: args{
			dataFormat: df.JSON,
			expr:       "data",
			length:     2,
		}, wantErr: true},

		//YAML
		{name: "expression does not point at slice", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat: df.YAML,
			expr:       "$.count",
			length:     0,
		}, wantErr: true},
		{name: "expression does point at slice and expected invalid length", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat: df.YAML,
			expr:       "$.data",
			length:     1,
		}, wantErr: false},
		{name: "expression does point at slice but expected proper length", fields: fields{body: []byte(yaml)}, args: args{
			dataFormat: df.YAML,
			expr:       "$.data",
			length:     2,
		}, wantErr: true},

		//XML
		{name: "expression does not point at slice", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: df.XML,
			expr:       "//count",
			length:     0,
		}, wantErr: true},
		{name: "expression does point at slice and expected invalid length", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: df.XML,
			expr:       "//data//name",
			length:     1,
		}, wantErr: false},
		{name: "expression does point at slice but expected proper length #1", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: df.XML,
			expr:       "//data//name",
			length:     2,
		}, wantErr: true},
		{name: "expression does point at slice but expected proper length #2", fields: fields{body: []byte(xml)}, args: args{
			dataFormat: df.XML,
			expr:       "//user",
			length:     2,
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: ioutil.NopCloser(bytes.NewReader(tt.fields.body))})

			if err := s.AssertNodeSliceLengthIsNot(tt.args.dataFormat, tt.args.expr, tt.args.length); (err != nil) != tt.wantErr {
				t.Errorf("AssertNodeSliceLengthIsNot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeSliceLengthIsNot() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user":[]}`))})

	err := apiCtx.AssertNodeSliceLengthIsNot(df.JSON, "user", 0)
	fmt.Println(err)

	// Output:
	// node 'user' contains slice(array) which has length: 0, but expected not to have it
}

func TestAPIContext_AssertNodeIsTypeAndValue(t *testing.T) {
	type fields struct {
		saved        map[string]any
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		df     df.DataFormat
		node   string
		goType types.DataType
		val    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// JSON related types against json data
		{
			name:    "selected node does not exists",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "abc", goType: types.Null, val: "abc"},
			wantErr: true,
		},
		{
			name:    "selected node value is boolean - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isHorizontal", goType: types.Boolean, val: "false"},
			wantErr: false,
		},
		{
			name:    "selected node value is boolean - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isHorizontal", goType: types.Boolean, val: "false"},
			wantErr: false,
		},
		{
			name:    "selected node value is string - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "color", goType: types.String, val: "purple"},
			wantErr: false,
		},
		{
			name:    "selected node value is string - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.color", goType: types.String, val: "purple"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #1 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_1", goType: types.Number, val: "210"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #1 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_1", goType: types.Number, val: "210"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #2 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_2", goType: types.Number, val: "-210"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #2 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_2", goType: types.Number, val: "-210"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #3 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_3", goType: types.Number, val: "21.05"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #3 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_3", goType: types.Number, val: "21.05"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #4 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_4", goType: types.Number, val: "1.0E+2"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #4 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_4", goType: types.Number, val: "1.0E+2"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #5 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "user.height", goType: types.Number, val: "180"},
			wantErr: false,
		},
		{
			name:    "selected node value is number #5- - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.user.height", goType: types.Number, val: "180"},
			wantErr: false,
		},

		// YAML related data types against YAML data
		{
			name:    "selected node does not exist",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.abc", goType: types.Null},
			wantErr: true,
		},
		{
			name:    "selected node value is scalar - string #1",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.doe", goType: types.Scalar, val: "a deer, a female deer"},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - string #2",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas-fifth-day.calling-birds", goType: types.Scalar, val: "four"},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - float",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.pi", goType: types.Scalar, val: "3.14159"},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - int #1",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.french-hens", goType: types.Scalar, val: "3"},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - int #2",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas-fifth-day.french-hens", goType: types.Scalar, val: "3"},
			wantErr: false,
		},
		{
			name:    "selected node value is scalar - boolean",
			fields:  fields{lastResponse: &http.Response{Body: createYamlRespBody()}},
			args:    args{df: df.YAML, node: "$.xmas", goType: types.Scalar, val: "true"},
			wantErr: false,
		},

		//GO related types against JSON data
		{
			name:    "selected node value is int #1 - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_1", goType: types.Int, val: "210"},
			wantErr: false,
		},
		{
			name:    "selected node value is int #1 - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_1", goType: types.Int, val: "210"},
			wantErr: false,
		},
		{
			name:    "selected node value is float - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "number_3", goType: types.Float, val: "21.05"},
			wantErr: false,
		},
		{
			name:    "selected node value is float - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.number_3", goType: types.Float, val: "21.05"},
			wantErr: false,
		},
		{
			name:    "selected node value is string - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "color", goType: types.String, val: "purple"},
			wantErr: false,
		},
		{
			name:    "selected node value is string - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.color", goType: types.String, val: "purple"},
			wantErr: false,
		},
		{
			name:    "selected node value is bool - gjson",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "isHorizontal", goType: types.Bool, val: "false"},
			wantErr: false,
		},
		{
			name:    "selected node value is bool - oliveagle",
			fields:  fields{lastResponse: &http.Response{Body: createRespBody()}},
			args:    args{df: df.JSON, node: "$.isHorizontal", goType: types.Bool, val: "false"},
			wantErr: false,
		},

		// XML
		{
			name:    "selected node does not exist",
			fields:  fields{lastResponse: &http.Response{Body: createXMLRespBody()}},
			args:    args{df: df.XML, node: "//abc", goType: types.Nil, val: "true"},
			wantErr: true,
		},
		{
			name:    "selected node value is string ",
			fields:  fields{lastResponse: &http.Response{Body: createXMLRespBody()}},
			args:    args{df: df.XML, node: "//description", goType: types.String, val: "Weather info"},
			wantErr: false,
		},
		{
			name:    "selected node value is boolean ",
			fields:  fields{lastResponse: &http.Response{Body: createXMLRespBody()}},
			args:    args{df: df.XML, node: "//isHot", goType: types.Boolean, val: "true"},
			wantErr: false,
		},
		{
			name:    "selected node value is float ",
			fields:  fields{lastResponse: &http.Response{Body: createXMLRespBody()}},
			args:    args{df: df.XML, node: "//weather//degrees", goType: types.Float, val: "67.5"},
			wantErr: false,
		},
		{
			name:    "selected node value is integer ",
			fields:  fields{lastResponse: &http.Response{Body: createXMLRespBody()}},
			args:    args{df: df.XML, node: "//weather//humidity", goType: types.Integer, val: "95"},
			wantErr: false,
		},

		// HTML
		{
			name:    "selected HTML node does not exist",
			fields:  fields{lastResponse: &http.Response{Body: createHTMLRespBody()}},
			args:    args{df: df.HTML, node: "//strong", goType: types.String, val: "true"},
			wantErr: true,
		},
		{
			name:    "selected HTML node exist but its value is different",
			fields:  fields{lastResponse: &http.Response{Body: createHTMLRespBody()}},
			args:    args{df: df.HTML, node: "//i[2]", goType: types.String, val: "3"},
			wantErr: true,
		},
		{
			name:    "selected HTML node exist and its value is correct",
			fields:  fields{lastResponse: &http.Response{Body: createHTMLRespBody()}},
			args:    args{df: df.HTML, node: "//i[2]", goType: types.String, val: "2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := NewDefaultAPIContext(tt.fields.isDebug, "")

			af.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			if err := af.AssertNodeIsTypeAndValue(tt.args.df, tt.args.node, tt.args.goType, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("AssertNodeIsTypeAndValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeIsTypeAndValue() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user": "abc"}`))})

	err := apiCtx.AssertNodeIsTypeAndValue(df.JSON, "user", types.String, "def")
	fmt.Println(err)

	// Output:
	// node 'user' has string value: 'abc', but expected: 'def'
}

func TestAPIContext_AssertNodeIsTypeAndHasOneOfValues(t *testing.T) {
	type fields struct {
		lastResponse *http.Response
		cacheData    map[string]any
	}
	type args struct {
		dataFormat      df.DataFormat
		exprTemplate    string
		dataType        types.DataType
		valuesTemplates string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "template engine can't fulfill valuesTemplates because of missing cache data",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.user.name",
				dataType:        types.String,
				valuesTemplates: "{{.NOT_EXISTING_VALUE}}",
			},
			wantErr: true,
		},
		{
			name: "template engine can't fulfill exprTemplate because of missing cache data",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.user.{{.PROPERTY}}",
				dataType:        types.String,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "missing last response body",
			fields: fields{
				lastResponse: nil,
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.user.name",
				dataType:        types.String,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at node that doesn't exist",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.abc",
				dataType:        types.String,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at node that has different type than expected",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.color",
				dataType:        types.Number,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at node and one of values is present #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.color",
				dataType:        types.String,
				valuesTemplates: "purple",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.color",
				dataType:        types.String,
				valuesTemplates: "red, purple",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.user.name",
				dataType:        types.String,
				valuesTemplates: "Ivo, John, Kristine",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #4",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.number_1",
				dataType:        types.Number,
				valuesTemplates: "210, 20, -30",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #4",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.isHorizontal",
				dataType:        types.Boolean,
				valuesTemplates: "true, false",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and none of values is present #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.number_2",
				dataType:        types.Int,
				valuesTemplates: "1, 2, -3",
			},
			wantErr: true,
		},
		{
			name: "expression points at node and none of values is present #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.user.roles[0]",
				dataType:        types.String,
				valuesTemplates: "a, b",
			},
			wantErr: true,
		},
		{
			name: "expression points at node and none of values is present #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.JSON,
				exprTemplate:    "$.isHorizontal",
				dataType:        types.Boolean,
				valuesTemplates: "true",
			},
			wantErr: true,
		},

		// YAML
		{
			name: "expression points at node that doesn't exist",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.abc",
				dataType:        types.Scalar,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at node that has different type than expected",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.pi",
				dataType:        types.Mapping,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at node and one of values is present #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.xmas-fifth-day.calling-birds",
				dataType:        types.Scalar,
				valuesTemplates: "four",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.xmas-fifth-day.calling-birds",
				dataType:        types.String,
				valuesTemplates: "three, four",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.xmas-fifth-day.calling-birds",
				dataType:        types.String,
				valuesTemplates: "one, four, three",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #4",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.pi",
				dataType:        types.Scalar,
				valuesTemplates: "3, 3.14, 3.14159",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #4",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.xmas",
				dataType:        types.Scalar,
				valuesTemplates: "true, false",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and none of values is present #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.french-hens",
				dataType:        types.Scalar,
				valuesTemplates: "1, 2, -3",
			},
			wantErr: true,
		},
		{
			name: "expression points at node and none of values is present #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.doe",
				dataType:        types.String,
				valuesTemplates: "a, b",
			},
			wantErr: true,
		},
		{
			name: "expression points at node and none of values is present #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.YAML,
				exprTemplate:    "$.xmas",
				dataType:        types.Scalar,
				valuesTemplates: "false",
			},
			wantErr: true,
		},
		// XML
		{
			name: "expression points at node that doesn't exist",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//abc",
				dataType:        types.String,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at node that has different type than expected",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//date",
				dataType:        types.Mapping,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at node and one of values is present #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//description",
				dataType:        types.String,
				valuesTemplates: "Weather info",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//description",
				dataType:        types.String,
				valuesTemplates: "three, Weather info",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//description",
				dataType:        types.String,
				valuesTemplates: "one, four, Weather info",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #4",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//humidity",
				dataType:        types.Int,
				valuesTemplates: "3, 95, 15, 32",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and one of values is present #4",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//degrees",
				dataType:        types.Float,
				valuesTemplates: "10.5, 67.5,13,34",
			},
			wantErr: false,
		},
		{
			name: "expression points at node and none of values is present #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//description",
				dataType:        types.String,
				valuesTemplates: "a, b, c",
			},
			wantErr: true,
		},
		{
			name: "expression points at node and none of values is present #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//degrees",
				dataType:        types.Int,
				valuesTemplates: "1,2,3",
			},
			wantErr: true,
		},
		{
			name: "expression points at node and none of values is present #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.XML,
				exprTemplate:    "//isHot",
				dataType:        types.Boolean,
				valuesTemplates: "false",
			},
			wantErr: true,
		},

		// HTML
		{
			name: "expression points at HTML node that doesn't exist",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.HTML,
				exprTemplate:    "//abc",
				dataType:        types.String,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at HTML node that exists but has different type",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.HTML,
				exprTemplate:    "//p",
				dataType:        types.Int,
				valuesTemplates: "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at HTML node that exists but none of values are correct",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.HTML,
				exprTemplate:    "//i",
				dataType:        types.Int,
				valuesTemplates: "2,3,4",
			},
			wantErr: true,
		},
		{
			name: "expression points at HTML node that exists and one of values is correct",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
				cacheData:    nil,
			},
			args: args{
				dataFormat:      df.HTML,
				exprTemplate:    "//i",
				dataType:        types.Int,
				valuesTemplates: "2,3,1,4",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCtx := NewDefaultAPIContext(false, "")

			apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			for key, val := range tt.fields.cacheData {
				apiCtx.Cache.Save(key, val)
			}

			if err := apiCtx.AssertNodeIsTypeAndHasOneOfValues(tt.args.dataFormat, tt.args.exprTemplate, tt.args.dataType, tt.args.valuesTemplates); (err != nil) != tt.wantErr {
				t.Errorf("AssertNodeIsTypeAndHasOneOfValues() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeIsTypeAndHasOneOfValues() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user": "abc"}`))})

	err := apiCtx.AssertNodeIsTypeAndHasOneOfValues(df.JSON, "user", types.String, "def, ghi")
	fmt.Println(err)

	// Output:
	// node 'user' doesn't contain any of: []string{"def", "ghi"}
}

func TestAPIContext_AsserNodeContainsSubString(t *testing.T) {
	type fields struct {
		lastResponse *http.Response
		cacheData    map[string]any
	}
	type args struct {
		dataFormat   df.DataFormat
		exprTemplate string
		sub          string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "template engine can't fulfill template because cache is missing data",
			fields: fields{
				lastResponse: &http.Response{Body: createRespBody()},
				cacheData:    nil,
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.{{.USER_PROPERTY}}",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "missing last response body",
			fields: fields{
				lastResponse: nil,
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.{{.USER_PROPERTY}}",
				sub:          "John",
			},
			wantErr: true,
		},
		// JSON related
		{
			name: "expression points at not existing node",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.abc",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, but it doesn't contain string value",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.number_1",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value, but it does not contain substring",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.name",
				sub:          "Ivo",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.name",
				sub:          "Jo",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    map[string]any{"USER_PROPERTY": "name"},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.{{.USER_PROPERTY}}",
				sub:          "Jo",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    map[string]any{"USER_PROPERTY": "name", "USER_NAME_FIRST_LETTER": "J"},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.{{.USER_PROPERTY}}",
				sub:          "{{.USER_NAME_FIRST_LETTER}}",
			},
			wantErr: false,
		},

		// YAML RELATED
		{
			name: "expression points at not existing node",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.abc",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, but it doesn't contain string value",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.pi",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value, but it does not contain substring",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.doe",
				sub:          "Ivo",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.doe",
				sub:          "deer",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    map[string]any{"PROPERTY": "doe"},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.{{.PROPERTY}}",
				sub:          "deer",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    map[string]any{"PROPERTY": "doe", "PROPERTY_SUBSTRING": "deer"},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.{{.PROPERTY}}",
				sub:          "{{.PROPERTY_SUBSTRING}}",
			},
			wantErr: false,
		},

		// XML RELATED
		{
			name: "expression points at not existing node",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//abc",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, but it doesn't contain string value",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//date",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value, but it does not contain substring",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//description",
				sub:          "Ivo",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//description",
				sub:          "Weather",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    map[string]any{"PROPERTY": "description"},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//{{.PROPERTY}}",
				sub:          "Weather",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    map[string]any{"PROPERTY": "description", "PROPERTY_SUBSTRING": "Weather"},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//{{.PROPERTY}}",
				sub:          "{{.PROPERTY_SUBSTRING}}",
			},
			wantErr: false,
		},

		// HTML
		{
			name: "expression points at not existing HTML node",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
			},
			args: args{
				dataFormat:   df.HTML,
				exprTemplate: "//abc",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing HTML node but substring is not present",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
			},
			args: args{
				dataFormat:   df.HTML,
				exprTemplate: "//p",
				sub:          "abc",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing HTML node and substring is present",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
			},
			args: args{
				dataFormat:   df.HTML,
				exprTemplate: "//p",
				sub:          "world",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCtx := NewDefaultAPIContext(false, "")

			apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			for key, val := range tt.fields.cacheData {
				apiCtx.Cache.Save(key, val)
			}

			if err := apiCtx.AssertNodeContainsSubString(tt.args.dataFormat, tt.args.exprTemplate, tt.args.sub); (err != nil) != tt.wantErr {
				t.Errorf("AsserNodeContainsSubString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeContainsSubString() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user": "abc"}`))})

	err := apiCtx.AssertNodeContainsSubString(df.JSON, "user", "abcd")
	fmt.Println(err)

	// Output:
	// node 'user' string value doesn't contain any occurrence of 'abcd'
}

func TestAPIContext_AsserNodeNotContainsSubString(t *testing.T) {
	type fields struct {
		lastResponse *http.Response
		cacheData    map[string]any
	}
	type args struct {
		dataFormat   df.DataFormat
		exprTemplate string
		sub          string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "template engine can't fulfill template because cache is missing data",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    nil,
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.{{.USER_PROPERTY}}",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "missing last response body",
			fields: fields{
				lastResponse: nil,
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.{{.USER_PROPERTY}}",
				sub:          "John",
			},
			wantErr: true,
		},
		// JSON related
		{
			name: "expression points at not existing node",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.abc",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, but it doesn't contain string value",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.number_1",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value, but it does not contain substring",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.name",
				sub:          "Ivo",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.name",
				sub:          "Jo",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    map[string]any{"USER_PROPERTY": "name"},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.{{.USER_PROPERTY}}",
				sub:          "Jo",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigJson))},
				cacheData:    map[string]any{"USER_PROPERTY": "name", "USER_NAME_FIRST_LETTER": "J"},
			},
			args: args{
				dataFormat:   df.JSON,
				exprTemplate: "$.user.{{.USER_PROPERTY}}",
				sub:          "{{.USER_NAME_FIRST_LETTER}}",
			},
			wantErr: true,
		},

		// YAML RELATED
		{
			name: "expression points at not existing node",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.abc",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, but it doesn't contain string value",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.pi",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value, but it does not contain substring",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.doe",
				sub:          "Ivo",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.doe",
				sub:          "deer",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    map[string]any{"PROPERTY": "doe"},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.{{.PROPERTY}}",
				sub:          "deer",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigYaml))},
				cacheData:    map[string]any{"PROPERTY": "doe", "PROPERTY_SUBSTRING": "deer"},
			},
			args: args{
				dataFormat:   df.YAML,
				exprTemplate: "$.{{.PROPERTY}}",
				sub:          "{{.PROPERTY_SUBSTRING}}",
			},
			wantErr: true,
		},

		// XML RELATED
		{
			name: "expression points at not existing node",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//abc",
				sub:          "John",
			},
			wantErr: true,
		},
		//TODO: Until there is no mapper, XML nodes will be always seen as strings
		//{
		//	name: "expression points at existing node, but it doesn't contain string value",
		//	fields: fields{
		//		lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
		//	},
		//	args: args{
		//		dataFormat:   df.XML,
		//		exprTemplate: "//degrees",
		//		sub:          "John",
		//	},
		//	wantErr: true,
		//},
		{
			name: "expression points at existing node, with string value, but it does not contain substring",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//description",
				sub:          "Ivo",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #1",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//description",
				sub:          "Weather",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #2",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    map[string]any{"PROPERTY": "description"},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//{{.PROPERTY}}",
				sub:          "Weather",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing node, with string value and it contains substring #3",
			fields: fields{
				lastResponse: &http.Response{Body: io.NopCloser(bytes.NewBufferString(bigXML))},
				cacheData:    map[string]any{"PROPERTY": "description", "PROPERTY_SUBSTRING": "Weather"},
			},
			args: args{
				dataFormat:   df.XML,
				exprTemplate: "//{{.PROPERTY}}",
				sub:          "{{.PROPERTY_SUBSTRING}}",
			},
			wantErr: true,
		},

		// HTML
		{
			name: "expression points at not existing HTML node",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
			},
			args: args{
				dataFormat:   df.HTML,
				exprTemplate: "//abc",
				sub:          "John",
			},
			wantErr: true,
		},
		{
			name: "expression points at existing HTML node but substring is not present",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
			},
			args: args{
				dataFormat:   df.HTML,
				exprTemplate: "//p",
				sub:          "abc",
			},
			wantErr: false,
		},
		{
			name: "expression points at existing HTML node and substring is present",
			fields: fields{
				lastResponse: &http.Response{Body: createHTMLRespBody()},
			},
			args: args{
				dataFormat:   df.HTML,
				exprTemplate: "//p",
				sub:          "world",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCtx := NewDefaultAPIContext(false, "")

			apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			for key, val := range tt.fields.cacheData {
				apiCtx.Cache.Save(key, val)
			}

			if err := apiCtx.AssertNodeNotContainsSubString(tt.args.dataFormat, tt.args.exprTemplate, tt.args.sub); (err != nil) != tt.wantErr {
				t.Errorf("AsserNodeNotContainsSubString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeNotContainsSubString() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user": "abc"}`))})

	err := apiCtx.AssertNodeNotContainsSubString(df.JSON, "user", "a")
	fmt.Println(err)

	// Output:
	// node 'user' string value contain some 'a', but expected not to
}

func TestState_AssertNodeMatchesRegExp(t *testing.T) {
	mTemplateEngine := new(mockedTemplateEngine)
	mJsonPathResolver := new(mockedJsonPathFinder)

	type fields struct {
		cacheKeys      map[string]any
		templateEngine templateEngine
		pathResolvers  PathFinders
		respBody       string
		mockFunc       func()
	}
	type args struct {
		df             df.DataFormat
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
			df:             df.JSON,
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
			df:             df.JSON,
			expr:           "xxx",
			regExpTemplate: "abc",
		}, wantErr: true},
		{name: "node value does not match provided regExp #1", fields: fields{
			respBody: `{
	"name": "abcdef"
}`,
			mockFunc: func() {},
		}, args: args{
			df:             df.JSON,
			expr:           "name",
			regExpTemplate: "dd.*",
		}, wantErr: true},

		{name: "node value matches provided regExp #1", fields: fields{
			respBody: `{
	"name": "abcdef"
}`,
			mockFunc: func() {},
		}, args: args{
			df:             df.JSON,
			expr:           "name",
			regExpTemplate: "abc.*",
		}, wantErr: false},

		{name: "node value does not match provided regExp #2", fields: fields{
			respBody: `---
name: abcdef`,
			mockFunc: func() {},
		}, args: args{
			df:             df.YAML,
			expr:           "$.name",
			regExpTemplate: "dd.*",
		}, wantErr: true},

		{name: "node value matches provided regExp #2", fields: fields{
			respBody: `---
name: abcdef`,
			mockFunc: func() {},
		}, args: args{
			df:             df.YAML,
			expr:           "$.name",
			regExpTemplate: "abc.*",
		}, wantErr: false},
		{name: "node value matches provided regExp #3", fields: fields{
			respBody: `
<name>abcdef</name>`,
			mockFunc: func() {},
		}, args: args{
			df:             df.XML,
			expr:           "//name",
			regExpTemplate: "abc.*",
		}, wantErr: false},
		{name: "node value does not match provided regExp #3", fields: fields{
			respBody: `
<name>abcdef</name>`,
			mockFunc: func() {},
		}, args: args{
			df:             df.XML,
			expr:           "//name",
			regExpTemplate: "dd.*",
		}, wantErr: true},
		{name: "node value matches provided regExp #4", fields: fields{
			respBody: `<html>
  <head>
    <title>Div Align Attribbute</title>
  </head>
  <body>
    <div align="left">
      a
    </div>
    <div align="right">
      b
    </div>
    <div align="center">
      c
    </div>
    <div align="justify">
      d
    </div>
	<div>
		<p>Hello world</p>
	</div>
	<div>
		<i>1</i>
		<i>2</i>
        <i>3</i>
	</div>
  </body>
</html>`,
			mockFunc: func() {},
		}, args: args{
			df:             df.HTML,
			expr:           "//p",
			regExpTemplate: "worl.*",
		}, wantErr: false},
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

			if err := s.AssertNodeMatchesRegExp(tt.args.df, tt.args.expr, tt.args.regExpTemplate); (err != nil) != tt.wantErr {
				t.Errorf("TheJSONNodeShouldMatchRegExp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeMatchesRegExp() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user": "abc"}`))})

	err := apiCtx.AssertNodeMatchesRegExp(df.JSON, "user", `ac.*`)
	fmt.Println(err)

	// Output:
	// node 'user' does not match regExp: 'ac.*'
}

func TestAPIContext_AssertNodeNotMatchesRegExp(t *testing.T) {
	mTemplateEngine := new(mockedTemplateEngine)
	mJsonPathResolver := new(mockedJsonPathFinder)

	type fields struct {
		cacheKeys      map[string]any
		templateEngine templateEngine
		pathResolvers  PathFinders
		respBody       string
		mockFunc       func()
	}
	type args struct {
		df             df.DataFormat
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
			df:             df.JSON,
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
			df:             df.JSON,
			expr:           "xxx",
			regExpTemplate: "abc",
		}, wantErr: true},
		{name: "node value does not match provided regExp #1", fields: fields{
			respBody: `{
	"name": "abcdef"
}`,
			mockFunc: func() {},
		}, args: args{
			df:             df.JSON,
			expr:           "name",
			regExpTemplate: "dd.*",
		}, wantErr: false},

		{name: "node value matches provided regExp #1", fields: fields{
			respBody: `{
	"name": "abcdef"
}`,
			mockFunc: func() {},
		}, args: args{
			df:             df.JSON,
			expr:           "name",
			regExpTemplate: "abc.*",
		}, wantErr: true},

		{name: "node value does not match provided regExp #2", fields: fields{
			respBody: `---
name: abcdef`,
			mockFunc: func() {},
		}, args: args{
			df:             df.YAML,
			expr:           "$.name",
			regExpTemplate: "dd.*",
		}, wantErr: false},

		{name: "node value matches provided regExp #2", fields: fields{
			respBody: `---
name: abcdef`,
			mockFunc: func() {},
		}, args: args{
			df:             df.YAML,
			expr:           "$.name",
			regExpTemplate: "abc.*",
		}, wantErr: true},
		{name: "node value matches provided regExp #3", fields: fields{
			respBody: `
<name>abcdef</name>`,
			mockFunc: func() {},
		}, args: args{
			df:             df.XML,
			expr:           "//name",
			regExpTemplate: "abc.*",
		}, wantErr: true},
		{name: "node value does not match provided regExp #3", fields: fields{
			respBody: `
<name>abcdef</name>`,
			mockFunc: func() {},
		}, args: args{
			df:             df.XML,
			expr:           "//name",
			regExpTemplate: "dd.*",
		}, wantErr: false},
		{name: "node value does not match provided regExp #4", fields: fields{
			respBody: `<html>
  <head>
    <title>Div Align Attribbute</title>
  </head>
  <body>
    <div align="left">
      a
    </div>
    <div align="right">
      b
    </div>
    <div align="center">
      c
    </div>
    <div align="justify">
      d
    </div>
	<div>
		<p>Hello world</p>
	</div>
	<div>
		<i>1</i>
		<i>2</i>
        <i>3</i>
	</div>
  </body>
</html>`,
			mockFunc: func() {},
		}, args: args{
			df:             df.HTML,
			expr:           "//p",
			regExpTemplate: "World.*",
		}, wantErr: false},
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
			if err := s.AssertNodeNotMatchesRegExp(tt.args.df, tt.args.expr, tt.args.regExpTemplate); (err != nil) != tt.wantErr {
				t.Errorf("AssertNodeNotMatchesRegExp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertNodeNotMatchesRegExp() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user": "abc"}`))})

	err := apiCtx.AssertNodeNotMatchesRegExp(df.JSON, "user", `ab.*`)
	fmt.Println(err)

	// Output:
	// node 'user' matches regExp: 'ab.*', but expected not to
}

func TestState_AssertResponseHeaderExists(t *testing.T) {
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

			if err := s.AssertResponseHeaderExists(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseHeaderExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertResponseHeaderExists() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Header: map[string][]string{"Content-Type": {"application/json"}}})

	err := apiCtx.AssertResponseHeaderExists("Content-Length")
	fmt.Println(err)

	// Output:
	// could not find header 'Content-Length' in last HTTP response
}

func TestAPIContext_AssertResponseHeaderNotExists(t *testing.T) {
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
		{name: "no headers in request", fields: fields{lastResponse: &http.Response{Header: map[string][]string{}}}, args: args{name: "Content-Type"}, wantErr: false},
		{name: "empty string provided as header name", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: ""}, wantErr: false},
		{name: "matching header #1 - case insensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "content-type"}, wantErr: true},
		{name: "matching header #2 - case sensitive", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}}},
		}, args: args{name: "Content-Type"}, wantErr: true},
		{name: "matching header #3", fields: fields{lastResponse: &http.Response{
			Header: map[string][]string{
				"Content-Length": {"30"},
				"Content-Type":   {"application/json"},
			},
		},
		}, args: args{name: "Content-Type"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)
			if err := s.AssertResponseHeaderNotExists(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseHeaderNotExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertResponseHeaderNotExists() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Header: map[string][]string{"Content-Type": {"application/json"}}})

	err := apiCtx.AssertResponseHeaderNotExists("content-type")
	fmt.Println(err)

	// Output:
	// last HTTP(s) response has header 'content-type', but expected not to
}

func TestState_AssertResponseHeaderValueIs(t *testing.T) {
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

			if err := s.AssertResponseHeaderValueIs(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseHeaderValueIs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleAPIContext_AssertResponseHeaderValueIs() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Header: map[string][]string{"Content-Type": {"application/json"}}})

	err := apiCtx.AssertResponseHeaderValueIs("content-type", "application/json+ld")
	fmt.Println(err)

	// Output:
	// last HTTP(s) response contains header 'content-type', but it's expected value: 'application/json+ld', is not equal to actual value: 'application/json'
}

func TestState_AssertResponseMatchesSchemaByReference(t *testing.T) {
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

			if err := s.AssertResponseMatchesSchemaByReference(tt.args.schemaPath); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseMatchesSchemaByReference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_AssertResponseMatchesSchemaByString(t *testing.T) {
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

			if err := s.AssertResponseMatchesSchemaByString(tt.args.jsonSchema); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseMatchesSchemaByReference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_AssertNodeMatchesSchemaByString(t *testing.T) {
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
		df         df.DataFormat
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
			df:         df.JSON,
			expr:       "",
			jsonSchema: "",
		}, wantErr: true},
		{name: "resolver could not resolve json path", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{}`))}}, args: args{
			df:         df.JSON,
			expr:       "",
			jsonSchema: "",
		}, wantErr: true},
		{name: "invalid json schema", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(jsonData))}}, args: args{
			df:   df.JSON,
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
			df:   df.JSON,
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
			df:   df.JSON,
			expr: "count",
			jsonSchema: `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "count integer",
    "description": "number of items in data array",
    "type": "integer"
}`,
		}, wantErr: false},

		{name: "yaml node value is valid according to provided JSON schema #1", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(yamlData))}}, args: args{
			df:   df.YAML,
			expr: "$.user.name",
			jsonSchema: `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "user name",
    "description": "name of user",
    "type": "string"
}`,
		}, wantErr: false},
		{name: "yaml node value is valid according to provided JSON schema #2", fields: fields{response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(yamlData))}}, args: args{
			df:   df.YAML,
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

			if err := s.AssertNodeMatchesSchemaByString(tt.args.df, tt.args.expr, tt.args.jsonSchema); (err != nil) != tt.wantErr {
				t.Errorf("IValidateJSONNodeWithSchemaString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_AssertTimeBetweenRequestAndResponseIs(t *testing.T) {
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

			if err := s.AssertTimeBetweenRequestAndResponseIs(td); (err != nil) != tt.wantErr {
				t.Errorf("TimeBetweenLastHTTPRequestResponseShouldBeLessThan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_AssertResponseCookieValueIs(t *testing.T) {
	mTemplateEngine := new(mockedTemplateEngine)

	type fields struct {
		TemplateEngine templateEngine
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

			if err := s.AssertResponseCookieValueIs(tt.args.name, tt.args.valueTemplate); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseCookieValueIs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_AssertResponseCookieExists(t *testing.T) {
	mTemplateEngine := new(mockedTemplateEngine)

	type fields struct {
		TemplateEngine templateEngine
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

			if err := s.AssertResponseCookieExists(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseCookieExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPIContext_AssertResponseCookieValueMatchesRegExp(t *testing.T) {
	mTemplateEngine := new(mockedTemplateEngine)

	type fields struct {
		TemplateEngine templateEngine
		response       *http.Response
		mockFunc       func()
	}
	type args struct {
		name           string
		regExpTemplate string
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
			name:           "",
			regExpTemplate: "",
		}, wantErr: true},
		{name: "template engine returns error", fields: fields{
			TemplateEngine: mTemplateEngine,
			response:       &http.Response{},
			mockFunc: func() {
				mTemplateEngine.On("Replace", "a", mock.Anything).Return("", errors.New("abc")).Once()
			},
		}, args: args{
			name:           "",
			regExpTemplate: "a",
		}, wantErr: true},
		{name: "cookies are not found in response", fields: fields{
			TemplateEngine: mTemplateEngine,
			response:       &http.Response{},
			mockFunc: func() {
				mTemplateEngine.On("Replace", "a", mock.Anything).Return("a", nil).Once()
			},
		}, args: args{
			name:           "a",
			regExpTemplate: "a",
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")
			s.SetTemplateEngine(tt.fields.TemplateEngine)
			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.response)

			tt.fields.mockFunc()

			if err := s.AssertResponseCookieValueMatchesRegExp(tt.args.name, tt.args.regExpTemplate); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseCookieValueMatchesRegExp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPIContext_AssertResponseCookieNotExists(t *testing.T) {
	mTemplateEngine := new(mockedTemplateEngine)

	type fields struct {
		TemplateEngine templateEngine
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
		{name: "cookies are not found in response", fields: fields{
			TemplateEngine: mTemplateEngine,
			response:       &http.Response{},
			mockFunc: func() {
			},
		}, args: args{
			name: "a",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(false, "")
			s.SetTemplateEngine(tt.fields.TemplateEngine)
			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.response)

			tt.fields.mockFunc()
			if err := s.AssertResponseCookieNotExists(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("AssertResponseCookieNotExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_Save(t *testing.T) {
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

			if err := s.Save(tt.args.value, tt.args.cacheKey); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
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

func ExampleAPIContext_Save() {
	apiCtx := NewDefaultAPIContext(false, "")

	apiCtx.Cache.Save("MY_KEY", "MY_VAL")

	val, err := apiCtx.Cache.GetSaved("MY_KEY")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(val)
	// Output:
	// MY_VAL
}

func TestState_SaveNode(t *testing.T) {
	type fields struct {
		cache        cacheable
		lastResponse *http.Response
		isDebug      bool
	}
	type args struct {
		df           df.DataFormat
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
		}, args: args{df: df.JSON, node: "token", variableName: "TOKEN"}, wantErr: true},
		{name: "invalid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{df: df.JSON, node: "last_name", variableName: "LAST_NAME"}, wantErr: true},
		{name: "valid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{df: df.JSON, node: "user.last_name", variableName: "LAST_NAME"}, wantErr: false},
		{name: "valid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": {
		"name": "a",
		"last_name": "b"
	}
}`))},
		}, args: args{df: df.JSON, node: "$.user", variableName: "USER"}, wantErr: false},

		//YAML
		{name: "invalid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{
	"user": "abc"
}`))},
		}, args: args{df: df.YAML, node: "$.token", variableName: "TOKEN"}, wantErr: true},
		{name: "invalid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
---
user:
   name: four
   last_name: b
`))},
		}, args: args{df: df.YAML, node: "$.last_name", variableName: "LAST_NAME"}, wantErr: true},
		{name: "valid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
---
user:
   name: four
   last_name: b
`))},
		}, args: args{df: df.YAML, node: "$.user.last_name", variableName: "LAST_NAME"}, wantErr: false},
		{name: "valid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`
---
user:
   name: four
   last_name: b
`))},
		}, args: args{df: df.YAML, node: "$.user", variableName: "USER"}, wantErr: false},

		//XML
		{name: "invalid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<user>
	<name>abc</name>
	<lastName>xxx</lastName>
</user>`))},
		}, args: args{df: df.XML, node: "//token", variableName: "TOKEN"}, wantErr: true},
		{name: "invalid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<user>
	<name>abc</name>
	<lastName>xxx</lastName>
</user>`))},
		}, args: args{df: df.XML, node: "//lastname", variableName: "LAST_NAME"}, wantErr: true},
		{name: "valid node #1", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<user>
	<name>abc</name>
	<lastName>xxx</lastName>
</user>`))},
		}, args: args{df: df.XML, node: "//lastName", variableName: "LAST_NAME"}, wantErr: false},
		{name: "valid node #2", fields: fields{
			cache: cache.NewConcurrentCache(),
			lastResponse: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`<?xml version="1.0"?>
<user>
	<name>abc</name>
	<lastName>xxx</lastName>
</user>`))},
		}, args: args{df: df.XML, node: "//user[1]//name", variableName: "USER"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultAPIContext(tt.fields.isDebug, "")

			s.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.lastResponse)

			err := s.SaveNode(tt.args.df, tt.args.node, tt.args.variableName)

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

func ExampleAPIContext_SaveNode() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Body: io.NopCloser(strings.NewReader(`{"user": "abc"}`))})

	err := apiCtx.SaveNode(df.JSON, "user", "USER_NAME")
	if err != nil {
		fmt.Println(err)

		return
	}

	val, err := apiCtx.Cache.GetSaved("USER_NAME")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(val)
	// Output:
	// abc
}

func TestAPIContext_SaveHeader(t *testing.T) {
	type fields struct {
		resp *http.Response
	}

	type args struct {
		name     string
		cacheKey string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    any
	}{
		{name: "missing last response", fields: fields{resp: nil}, args: args{
			name:     "Content-Length",
			cacheKey: "LENGTH_OF_CONTENT",
		}, wantErr: true, want: nil},
		{name: "last response has no headers", fields: fields{resp: &http.Response{}}, args: args{
			name:     "Content-Length",
			cacheKey: "LENGTH_OF_CONTENT",
		}, wantErr: true, want: nil},
		{name: "response has header from name field #1", fields: fields{resp: &http.Response{}}, args: args{
			name:     "Content-Length",
			cacheKey: "LENGTH_OF_CONTENT",
		}, wantErr: true, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCtx := NewDefaultAPIContext(false, "")

			apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, tt.fields.resp)

			if err := apiCtx.SaveHeader(tt.args.name, tt.args.cacheKey); (err != nil) != tt.wantErr {
				t.Errorf("SaveHeader() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				savedHeader, err := apiCtx.Cache.GetSaved(tt.args.cacheKey)
				if err != nil {
					t.Errorf("%s", err)
				}

				fmt.Println(savedHeader)
			}
		})
	}
}

func ExampleAPIContext_SaveHeader() {
	apiCtx := NewDefaultAPIContext(false, "")

	// instead of sending real HTTP(s) request with apiCtx.RequestSend
	// we simply mock last HTTP(s) request's response
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, &http.Response{Header: map[string][]string{"Content-Type": {"application/json"}}})

	if err := apiCtx.SaveHeader("content-type", "CONTENT_TYPE_VALUE"); err != nil {
		fmt.Println(err)

		return
	}

	val, err := apiCtx.Cache.GetSaved("CONTENT_TYPE_VALUE")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(val)

	// Output:
	// application/json
}
