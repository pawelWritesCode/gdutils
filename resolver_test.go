package gdutils

import (
	"reflect"
	"testing"
)

func TestResolve(t *testing.T) {
	type args struct {
		expr     string
		respBody []byte
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{name: "basic example object #1", args: args{
			expr: "body",
			respBody: []byte(`{
	"body": "this is value of key body"
}`),
		}, want: interface{}("this is value of key body"), wantErr: false},
		{name: "basic example array #1", args: args{
			expr: "data[0]",
			respBody: []byte(`{
	"data": [1, 2, 3]
}`),
		}, want: interface{}(float64(1)), wantErr: false},
		{name: "basic example two objects #1", args: args{
			expr: "project.user",
			respBody: []byte(`{
	"project": {
		"user": "adam"
	}
}`),
		}, want: interface{}("adam"), wantErr: false},
		{name: "basic example two objects #2", args: args{
			expr: "project.user",
			respBody: []byte(`{
	"project": {
		"anotherKey": 1,
		"user": "adam"
	}
}`),
		}, want: interface{}("adam"), wantErr: false},
		{name: "object with array", args: args{
			expr: "project.user[1].name",
			respBody: []byte(`{
	"project": {
		"user": [
			{
				"name": "abc"
			},
			{
				"name": "cde"
			}
		]
	}
}`),
		}, want: interface{}("cde"), wantErr: false},
		{name: "only array", args: args{
			expr: "root[0].name",
			respBody: []byte(`[
	{
		"name": "xxx"
	},
	{
		"name": "yyy"
	}
]`),
		}, want: interface{}("xxx"), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(tt.args.expr, tt.args.respBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resolve() got = %v, want %v", got, tt.want)
			}
		})
	}
}
