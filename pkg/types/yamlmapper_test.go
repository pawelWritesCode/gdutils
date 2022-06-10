package types

import "testing"

func TestYAMLTypeMapper_Map(t *testing.T) {
	type args struct {
		data any
	}
	tests := []struct {
		name string
		args args
		want DataType
	}{
		{name: "scalar - string", args: args{data: "abc"}, want: Scalar},
		{name: "scalar - int", args: args{data: 1}, want: Scalar},
		{name: "scalar - float", args: args{data: 1.1}, want: Scalar},
		{name: "scalar - boolean", args: args{data: true}, want: Scalar},
		{name: "object - map", args: args{data: make(map[string]int)}, want: Mapping},
		{name: "object - struct", args: args{data: struct{ key string }{key: "value"}}, want: Mapping},
		{name: "sequence - slice", args: args{data: []int{1}}, want: Sequence},
		{name: "sequence - array", args: args{data: [...]int{1}}, want: Sequence},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Y := YAMLTypeMapper{}
			if got := Y.Map(tt.args.data); got != tt.want {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}
