package format

import "testing"

func TestIsJSON(t *testing.T) {
	testCases := []string{
		`{"key": "value"}`,
		`{"key1": "value", "key2": "value"}`,
		`{"key1": []}`,
		`{"key1": ["aa"]}`,
	}

	for _, testCase := range testCases {
		if !IsJSON([]byte(testCase)) {
			t.Errorf("response '%s' expected to be JSON", testCase)
		}
	}
}

func TestIsYAML(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "json", args: args{bytes: []byte(`{"name": "abc"}`)}, want: false},
		{name: "plain text", args: args{bytes: []byte(`abcd efgh`)}, want: false},
		{name: "xml", args: args{bytes: []byte(`<?xml version="1.0"?>
<catalog>
   <book id="bk101">
      <author>Gambardella, Matthew</author>
      <title>XML Developer's Guide</title>
      <genre>Computer</genre>
      <price>44.95</price>
      <publish_date>2000-10-01</publish_date>
      <description>An in-depth look at creating applications 
      with XML.</description>
   </book>
</catalog>`)}, want: false},
		{name: "yaml", args: args{bytes: []byte(`---
name: "abc"`)}, want: true},
		{name: "yaml", args: args{bytes: []byte(`---
- id: 1
  firstName: ЦЧдОМг
  lastName: "doe-Ӧ\U000131C8㎆⋃ང༑q\U0001D6A3\U0001302Eオ⠽"
  age: 39
  description: nJuadDs iJgg NEX
  friendSince: '2022-03-07T09:40:17Z'
  avatar: acuogn.gif
`)}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsYAML(tt.args.bytes); got != tt.want {
				t.Errorf("IsYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsXML(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "json", args: args{bytes: []byte(`{"name": "abc"}`)}, want: false},
		{name: "plain text", args: args{bytes: []byte(`abcd efgh`)}, want: false},
		{name: "xml", args: args{bytes: []byte(`<?xml version="1.0"?>
<catalog>
   <book id="bk101">
      <author>Gambardella, Matthew</author>
      <title>XML Developer's Guide</title>
      <genre>Computer</genre>
      <price>44.95</price>
      <publish_date>2000-10-01</publish_date>
      <description>An in-depth look at creating applications 
      with XML.</description>
   </book>
</catalog>`)}, want: true},
		{name: "yaml", args: args{bytes: []byte(`---
name: "abc"`)}, want: false},
		{name: "yaml", args: args{bytes: []byte(`<?xml version='1.0' encoding='us-ascii'?>

<!--  A SAMPLE set of slides  -->

<slideshow 
    title="Sample Slide Show"
    date="Date of publication"
    author="Yours Truly"
    >

    <!-- TITLE SLIDE -->
    <slide type="all">
      <title>Wake up to WonderWidgets!</title>
    </slide>

    <!-- OVERVIEW -->
    <slide type="all">
        <title>Overview</title>
        <item>Why <em>WonderWidgets</em> are great</item>
        <item/>
        <item>Who <em>buys</em> WonderWidgets</item>
    </slide>

</slideshow>
`)}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsXML(tt.args.bytes); got != tt.want {
				t.Errorf("IsXML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHTML(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "valid example #1", args: args{b: []byte(`<!DOCTYPE html>
<html>
<body>

<h1>My First Heading</h1>

<p>My first paragraph.</p>

</body>
</html>`)}, want: true},
		{name: "valid example #1", args: args{b: []byte(`<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Img Width Attribute</title>
  </head>
  <body>
    <img src="image.png" alt="Image" width="100" />
  </body>
</html>
`)}, want: true},
		{name: "valid example #2", args: args{b: []byte(`<html>
  <head>
    <title>Div Align Attribbute</title>
  </head>
  <body>
    <div align="left">
      Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut
      labore et dolore magna aliqua.
    </div>
    <div align="right">
      Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut
      labore et dolore magna aliqua.
    </div>
    <div align="center">
      Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut
      labore et dolore magna aliqua.
    </div>
    <div align="justify">
      Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut
      labore et dolore magna aliqua.
    </div>
  </body>
</html>
`)}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHTML(tt.args.b); got != tt.want {
				t.Errorf("IsHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPlainText(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "simple plain text #1", args: args{b: []byte("abc")}, want: true},
		{name: "simple plain text #2", args: args{b: []byte(`Lorem ipsum dolor sit amet
consectetur adipiscing elit. Lorem ipsum dolor sit amet`)}, want: true},
		{name: "simple plain text #3", args: args{b: []byte("☫ ☬ ☭ ⚒ ⚔")}, want: true},
		{name: "simple plain text #4", args: args{b: []byte(`Those are examples of different data types:
JSON:
{
  "user": "Ivo"
}

XML:
<?xml version="1.0"?>
<catalog>
   <book id="bk101">
      <author>Gambardella, Matthew</author>
      <title>XML Developer's Guide</title>
      <genre>Computer</genre>
      <price>44.95</price>
      <publish_date>2000-10-01</publish_date>
      <description>An in-depth look at creating applications 
      with XML.</description>
   </book>
</catalog>

HTML:
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Img Width Attribute</title>
  </head>
  <body>
    <img src="image.png" alt="Image" width="100" />
  </body>
</html>

YAML:
---
- id: 1
  firstName: ЦЧдОМг
  lastName: "doe-Ӧ\U000131C8㎆⋃ང༑q\U0001D6A3\U0001302Eオ⠽"
  age: 39
  description: nJuadDs iJgg NEX
  friendSince: '2022-03-07T09:40:17Z'
  avatar: acuogn.gif
`)}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPlainText(tt.args.b); got != tt.want {
				t.Errorf("IsPlainText() = %v, want %v", got, tt.want)
			}
		})
	}
}
