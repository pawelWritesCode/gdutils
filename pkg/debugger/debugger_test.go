package debugger

import (
	"bytes"
	"io"
	"math"
	"os"
	"testing"
)

func TestDebuggerService_IsOn(t *testing.T) {
	d := New(false, true, math.MaxUint16, os.Stdout)
	if d.IsOn() {
		t.Errorf("IsOn should be false")
	}
}

func TestDebuggerService_TurnOn(t *testing.T) {
	d := New(false, true, math.MaxUint16, os.Stdout)

	if d.IsOn() {
		t.Errorf("IsOn should be false")
	}

	d.TurnOn()

	if !d.IsOn() {
		t.Errorf("IsOn should be true")
	}
}

func TestDebuggerService_TurnOff(t *testing.T) {
	d := New(false, true, math.MaxUint16, os.Stdout)

	if d.IsOn() {
		t.Errorf("IsOn should be false")
	}

	d.TurnOn()

	if !d.IsOn() {
		t.Errorf("IsOn should be true")
	}

	d.TurnOff()

	if d.IsOn() {
		t.Errorf("IsOn should be false again")
	}
}

func TestDebuggerService_Reset(t *testing.T) {
	d := New(false, true, math.MaxUint16, os.Stdout)

	if d.IsOn() {
		t.Errorf("IsOn should be false")
	}

	d.TurnOn()

	if !d.IsOn() {
		t.Errorf("IsOn should be true")
	}

	d.Reset(false)
	if d.IsOn() {
		t.Errorf("IsOn should be false again")
	}
}

func TestDebuggerService_prepareMessage(t *testing.T) {
	type fields struct {
		actualState bool
		limit       uint16
		isColored   bool
		writer      io.Writer
	}
	type args struct {
		info string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "incoming info has zero length",
			fields: fields{
				actualState: true,
				isColored:   false,
				limit:       3072,
				writer:      bytes.NewBuffer(make([]byte, 3072)),
			},
			args: args{info: ""},
			want: "",
		},
		{
			name: "incoming info has length greater than limit",
			fields: fields{
				actualState: true,
				isColored:   true,
				limit:       2,
				writer:      bytes.NewBuffer(make([]byte, 3072)),
			},
			args: args{info: "abc"},
			want: "ab",
		},
		{
			name: "incoming info has length less than limit and has format plain-text",
			fields: fields{
				actualState: true,
				isColored:   true,
				limit:       4,
				writer:      bytes.NewBuffer(make([]byte, 3072)),
			},
			args: args{info: "abc"},
			want: "abc",
		},
		{
			name: "incoming info has length less than limit and has format JSON",
			fields: fields{
				actualState: true,
				isColored:   false,
				limit:       100,
				writer:      bytes.NewBuffer(make([]byte, 3072)),
			},
			args: args{info: `{"user": "abc"}`},
			want: "{\n\t\"user\": \"abc\"\n}",
		},
		{
			name: "incoming info has length less than limit and has format HTML",
			fields: fields{
				actualState: true,
				isColored:   true,
				limit:       1000,
				writer:      bytes.NewBuffer(make([]byte, 3072)),
			},
			args: args{info: `<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Img Width Attribute</title>
  </head>
  <body>
    <img src="image.png" alt="Image" width="100" />
  </body>
</html>
`},
			want: `<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Img Width Attribute</title>
  </head>
  <body>
    <img src="image.png" alt="Image" width="100" />
  </body>
</html>
`,
		},
		{
			name: "incoming info has length less than limit and has format XML",
			fields: fields{
				actualState: true,
				isColored:   true,
				limit:       1000,
				writer:      bytes.NewBuffer(make([]byte, 3072)),
			},
			args: args{info: `<?xml version="1.0"?>
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
</catalog>`},
			want: `<?xml version="1.0"?>
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
</catalog>`,
		},
		{
			name: "incoming info has length less than limit and has format YAML",
			fields: fields{
				actualState: true,
				isColored:   true,
				limit:       1000,
				writer:      bytes.NewBuffer(make([]byte, 3072)),
			},
			args: args{info: `---
- id: 1
  firstName: ЦЧдОМг
  lastName: "doe-Ӧ\U000131C8㎆⋃ང༑q\U0001D6A3\U0001302Eオ⠽"
  age: 39
  description: nJuadDs iJgg NEX
  friendSince: '2022-03-07T09:40:17Z'
  avatar: acuogn.gif
`},
			want: `---
- id: 1
  firstName: ЦЧдОМг
  lastName: "doe-Ӧ\U000131C8㎆⋃ང༑q\U0001D6A3\U0001302Eオ⠽"
  age: 39
  description: nJuadDs iJgg NEX
  friendSince: '2022-03-07T09:40:17Z'
  avatar: acuogn.gif
`,
		},
		{
			name: "incoming info has length less than limit and has format JSON #2",
			fields: fields{
				actualState: true,
				isColored:   true,
				limit:       100,
				writer:      bytes.NewBuffer(make([]byte, 3072)),
			},
			args: args{info: `{"user": "abc"}`},
			want: "{\n  \"user\": \"abc\"\n}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DebuggerService{
				actualState: tt.fields.actualState,
				limit:       tt.fields.limit,
				writer:      tt.fields.writer,
				isColored:   tt.fields.isColored,
			}

			if got := d.prepareMessage(tt.args.info); got != tt.want {
				t.Errorf("prepareMessage()\n %#v\n want:\n %#v", got, tt.want)
			}
		})
	}
}
