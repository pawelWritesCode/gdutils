package pathfinder

import (
	"reflect"
	"testing"
)

func TestAntchfxHTMLFinder_Find(t *testing.T) {
	htmlBytes := []byte(`<html>
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
`)

	type args struct {
		expr string
		b    []byte
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{name: "title from head section", args: args{
			expr: "//head//title",
			b:    htmlBytes,
		}, want: any("Div Align Attribbute"), wantErr: false},
		{name: "div with align justify", args: args{
			expr: "//div[@align=\"justify\"]",
			b:    htmlBytes,
		}, want: any("d"), wantErr: false},
		{name: "second div", args: args{
			expr: "//div[2]",
			b:    htmlBytes,
		}, want: any("b"), wantErr: false},
		{name: "paragraph text ", args: args{
			expr: "//p",
			b:    htmlBytes,
		}, want: any("Hello world"), wantErr: false},
		{name: "second i", args: args{
			expr: "//i[2]",
			b:    htmlBytes,
		}, want: any("2"), wantErr: false},
		{name: "there are many nodes matching expression", args: args{
			expr: "//i",
			b:    htmlBytes,
		}, want: any("1"), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AntchfxHTMLFinder{}
			got, err := a.Find(tt.args.expr, tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
