package types

import "testing"

func TestGoTypeMapper_Map(t *testing.T) {
	type args struct {
		data any
	}
	tests := []struct {
		name string
		args args
		want DataType
	}{
		{name: "nil", args: args{nil}, want: Nil},
		{name: "string", args: args{"abc"}, want: String},
		{name: "int", args: args{34}, want: Int},
		{name: "float", args: args{3.14}, want: Float},
		{name: "bool", args: args{true}, want: Bool},
		{name: "map", args: args{make(map[string]int)}, want: Map},
		{name: "slice", args: args{make([]int, 1)}, want: Slice},
		{name: "slice", args: args{[...]int{1}}, want: Slice},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := GoTypeMapper{}
			if got := g.Map(tt.args.data); got != tt.want {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}
