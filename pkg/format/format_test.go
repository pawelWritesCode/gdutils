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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsYAML(tt.args.bytes); got != tt.want {
				t.Errorf("IsYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}
