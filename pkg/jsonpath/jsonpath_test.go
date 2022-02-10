package jsonpath

import (
	"reflect"
	"testing"
)

var jsonBytes = []byte(`{
    "store": {
        "book": [
            {
                "category": "reference",
                "author": "Nigel Rees",
				"available": true,
                "title": "Sayings of the Century",
                "price": 8.95
            },
            {
                "category": "fiction",
                "author": "Evelyn Waugh",
				"available": true,
                "title": "Sword of Honour",
                "price": 12.99
            },
            {
                "category": "fiction",
                "author": "Herman Melville",
				"available": true,
                "title": "Moby Dick",
                "isbn": "0-553-21311-3",
                "price": 8.99
            },
            {
                "category": "fiction",
                "author": "J. R. R. Tolkien",
				"available": false,
                "title": "The Lord of the Rings",
                "isbn": "0-395-19395-8",
                "price": 22.99
            }
        ],
        "bicycle": {
            "color": "red",
            "price": 19.95
        }
    },
    "expensive": 10
}`)

func TestQJSONResolver_Resolve(t *testing.T) {
	type args struct {
		expr      string
		jsonBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{name: "no expression", args: args{
			expr:      "",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "no jsonBytes", args: args{
			expr:      "data",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "expression points to nothing", args: args{
			expr:      "data",
			jsonBytes: jsonBytes,
		}, want: nil, wantErr: true},
		{name: "successful fetch data #1", args: args{
			expr:      "store.book[0].category",
			jsonBytes: jsonBytes,
		}, want: interface{}("reference"), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "store.book[1].price",
			jsonBytes: jsonBytes,
		}, want: interface{}(float64(12.99)), wantErr: false},
		{name: "successful fetch data #3", args: args{
			expr:      "store.book[2].available",
			jsonBytes: jsonBytes,
		}, want: interface{}(true), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "store.bicycle",
			jsonBytes: jsonBytes,
		}, want: interface{}(map[string]interface{}{"color": "red", "price": float64(19.95)}), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Q := QJSONResolver{}
			got, err := Q.Resolve(tt.args.expr, tt.args.jsonBytes)
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

func TestOliveagleJSONpath_Resolve(t *testing.T) {
	type args struct {
		expr      string
		jsonBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{name: "no expression", args: args{
			expr:      "",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "no jsonBytes", args: args{
			expr:      "data",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "expression points to nothing", args: args{
			expr:      "data",
			jsonBytes: []byte(`{"a": "b"}`),
		}, want: nil, wantErr: true},
		{name: "no expression", args: args{
			expr:      "",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "no jsonBytes", args: args{
			expr:      "data",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "expression points to nothing", args: args{
			expr:      "data",
			jsonBytes: jsonBytes,
		}, want: nil, wantErr: true},
		{name: "successful fetch data #1", args: args{
			expr:      "$.expensive",
			jsonBytes: jsonBytes,
		}, want: interface{}(float64(10)), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "$.store.book[0].price",
			jsonBytes: jsonBytes,
		}, want: interface{}(8.95), wantErr: false},
		{name: "successful fetch data #3", args: args{
			expr:      "$.store.book[-1].isbn",
			jsonBytes: jsonBytes,
		}, want: "0-395-19395-8", wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "$.store.bicycle",
			jsonBytes: jsonBytes,
		}, want: interface{}(map[string]interface{}{"color": "red", "price": float64(19.95)}), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "$.store.book[?(@.price > 10)].title",
			jsonBytes: jsonBytes,
		}, want: []interface{}{"Sword of Honour", "The Lord of the Rings"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := OliveagleJSONpath{}
			got, err := o.Resolve(tt.args.expr, tt.args.jsonBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resolve() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDynamicJSONPathResolver_Resolve(t *testing.T) {
	type fields struct {
		qjson             QJSONResolver
		oliveagleJSONpath OliveagleJSONpath
	}
	type args struct {
		expr      string
		jsonBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{name: "no expression", args: args{
			expr:      "",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "no jsonBytes", args: args{
			expr:      "data",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "expression points to nothing", args: args{
			expr:      "data",
			jsonBytes: []byte(`{"a": "b"}`),
		}, want: nil, wantErr: true},
		{name: "no expression", args: args{
			expr:      "",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "no jsonBytes", args: args{
			expr:      "data",
			jsonBytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "expression points to nothing", args: args{
			expr:      "data",
			jsonBytes: jsonBytes,
		}, want: nil, wantErr: true},
		{name: "successful fetch data #1", args: args{
			expr:      "$.expensive",
			jsonBytes: jsonBytes,
		}, want: interface{}(float64(10)), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "$.store.book[0].price",
			jsonBytes: jsonBytes,
		}, want: interface{}(8.95), wantErr: false},
		{name: "successful fetch data #3", args: args{
			expr:      "$.store.book[-1].isbn",
			jsonBytes: jsonBytes,
		}, want: "0-395-19395-8", wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "$.store.bicycle",
			jsonBytes: jsonBytes,
		}, want: interface{}(map[string]interface{}{"color": "red", "price": float64(19.95)}), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "$.store.book[?(@.price > 10)].title",
			jsonBytes: jsonBytes,
		}, want: []interface{}{"Sword of Honour", "The Lord of the Rings"}, wantErr: false},
		{name: "successful fetch data #1", args: args{
			expr:      "store.book[0].category",
			jsonBytes: jsonBytes,
		}, want: interface{}("reference"), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "store.book[1].price",
			jsonBytes: jsonBytes,
		}, want: interface{}(float64(12.99)), wantErr: false},
		{name: "successful fetch data #3", args: args{
			expr:      "store.book[2].available",
			jsonBytes: jsonBytes,
		}, want: interface{}(true), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "store.bicycle",
			jsonBytes: jsonBytes,
		}, want: interface{}(map[string]interface{}{"color": "red", "price": float64(19.95)}), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DynamicJSONPathResolver{
				qjson:             tt.fields.qjson,
				oliveagleJSONpath: tt.fields.oliveagleJSONpath,
			}
			got, err := d.Resolve(tt.args.expr, tt.args.jsonBytes)
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
