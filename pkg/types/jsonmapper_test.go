package types

import "testing"

func TestJSONTypeMapper_Map(t *testing.T) {
	type args struct {
		data any
	}
	tests := []struct {
		name string
		args args
		want DataType
	}{
		{name: "null #1", args: args{data: nil}, want: Null},
		{name: "string", args: args{data: "abc"}, want: String},
		{name: "number - int", args: args{data: 1}, want: Number},
		{name: "number - float", args: args{data: 1.1}, want: Number},
		{name: "boolean", args: args{data: true}, want: Boolean},
		{name: "object - map", args: args{data: make(map[string]int)}, want: Object},
		{name: "object - struct", args: args{data: struct{ key string }{key: "value"}}, want: Object},
		{name: "array - slice", args: args{data: []int{1}}, want: Array},
		{name: "array - array", args: args{data: [...]int{1}}, want: Array},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			J := JSONTypeMapper{}
			if got := J.Map(tt.args.data); got != tt.want {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}
