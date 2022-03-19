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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsXML(tt.args.bytes); got != tt.want {
				t.Errorf("IsXML() = %v, want %v", got, tt.want)
			}
		})
	}
}
