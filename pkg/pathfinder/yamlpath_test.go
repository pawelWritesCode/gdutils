package pathfinder

import (
	"reflect"
	"testing"
)

var yml = `
store:
  book:
    - author: john
      price: 10
    - author: ken
      price: 12
  bicycle:
    color: red
    price: 19.95
`

func TestGoccyGoYamlFinder_Find(t *testing.T) {
	type args struct {
		expr      string
		yamlBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{name: "no expression", args: args{
			expr:      "",
			yamlBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "no yaml bytes", args: args{
			expr:      "data",
			yamlBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "expression points to nothing", args: args{
			expr:      "data",
			yamlBytes: []byte(yml),
		}, want: nil, wantErr: true},
		{name: "successful fetch data #1", args: args{
			expr:      "$.store.book[0].author",
			yamlBytes: []byte(yml),
		}, want: any("john"), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "$.store.book[0].price",
			yamlBytes: []byte(yml),
		}, want: any(uint64(10)), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "$.store.bicycle.color",
			yamlBytes: []byte(yml),
		}, want: any("red"), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "$.store.book",
			yamlBytes: []byte(yml),
		}, want: []any{map[string]any{"author": "john", "price": uint64(10)}, map[string]any{"author": "ken", "price": uint64(12)}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := GoccyGoYamlFinder{}
			got, err := g.Find(tt.args.expr, tt.args.yamlBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %+v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
