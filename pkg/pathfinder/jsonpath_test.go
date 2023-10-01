package pathfinder

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

func TestOliveagleJSONFInder_Find(t *testing.T) {
	type args struct {
		expr      string
		jsonBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    any
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
		}, want: any(float64(10)), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "$.store.book[0].price",
			jsonBytes: jsonBytes,
		}, want: any(8.95), wantErr: false},
		{name: "successful fetch data #3", args: args{
			expr:      "$.store.book[-1].isbn",
			jsonBytes: jsonBytes,
		}, want: "0-395-19395-8", wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "$.store.bicycle",
			jsonBytes: jsonBytes,
		}, want: any(map[string]any{"color": "red", "price": float64(19.95)}), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "$.store.book[?(@.price > 10)].title",
			jsonBytes: jsonBytes,
		}, want: []any{"Sword of Honour", "The Lord of the Rings"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := OliveagleJSONFinder{}
			got, err := o.Find(tt.args.expr, tt.args.jsonBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDynamicJSONPathResolver_Find(t *testing.T) {
	type args struct {
		expr      string
		jsonBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    any
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
		}, want: any(float64(10)), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "$.store.book[0].price",
			jsonBytes: jsonBytes,
		}, want: any(8.95), wantErr: false},
		{name: "successful fetch data #3", args: args{
			expr:      "$.store.book[-1].isbn",
			jsonBytes: jsonBytes,
		}, want: "0-395-19395-8", wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "$.store.bicycle",
			jsonBytes: jsonBytes,
		}, want: any(map[string]any{"color": "red", "price": float64(19.95)}), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "$.store.book[?(@.price > 10)].title",
			jsonBytes: jsonBytes,
		}, want: []any{"Sword of Honour", "The Lord of the Rings"}, wantErr: false},
		{name: "successful fetch data #1", args: args{
			expr:      "store.book.0.category",
			jsonBytes: jsonBytes,
		}, want: any("reference"), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "store.book.1.price",
			jsonBytes: jsonBytes,
		}, want: any(float64(12.99)), wantErr: false},
		{name: "successful fetch data #3", args: args{
			expr:      "store.book.2.available",
			jsonBytes: jsonBytes,
		}, want: any(true), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "store.bicycle",
			jsonBytes: jsonBytes,
		}, want: any(map[string]any{"color": "red", "price": float64(19.95)}), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:      "/store/bicycle",
			jsonBytes: jsonBytes,
		}, want: any(map[string]any{"color": "red", "price": float64(19.95)}), wantErr: false},
		{name: "successful fetch data #1", args: args{
			expr:      "/store/book/*[1]/category",
			jsonBytes: jsonBytes,
		}, want: any("reference"), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:      "/store/book/*[2]/price",
			jsonBytes: jsonBytes,
		}, want: any(float64(12.99)), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DynamicJSONPathFinder{
				gjson:               NewGJSONFinder(),
				oliveagleJSONFinder: NewOliveagleJSONFinder(),
			}
			got, err := d.Find(tt.args.expr, tt.args.jsonBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGJSONFinder_Find(t *testing.T) {
	type args struct {
		expr  string
		bytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{name: "no expression", args: args{
			expr:  "",
			bytes: []byte("{}"),
		}, want: nil, wantErr: true},
		{name: "no jsonBytes", args: args{
			expr:  "data",
			bytes: []byte(""),
		}, want: nil, wantErr: true},
		{name: "expression points to nothing", args: args{
			expr:  "data",
			bytes: []byte(`{"a": "b"}`),
		}, want: nil, wantErr: true},
		{name: "successful fetch data #1", args: args{
			expr:  "expensive",
			bytes: jsonBytes,
		}, want: any(float64(10)), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:  "store.book.0.price",
			bytes: jsonBytes,
		}, want: any(8.95), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:  "store.bicycle",
			bytes: jsonBytes,
		}, want: any(map[string]any{"color": "red", "price": float64(19.95)}), wantErr: false},
		{name: "successful fetch data #1", args: args{
			expr:  "store.book.0.category",
			bytes: jsonBytes,
		}, want: any("reference"), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr:  "store.book.1.price",
			bytes: jsonBytes,
		}, want: any(float64(12.99)), wantErr: false},
		{name: "successful fetch data #3", args: args{
			expr:  "store.book.2.available",
			bytes: jsonBytes,
		}, want: any(true), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr:  "store.bicycle",
			bytes: jsonBytes,
		}, want: any(map[string]any{"color": "red", "price": float64(19.95)}), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			G := GJSONFinder{}
			got, err := G.Find(tt.args.expr, tt.args.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAntchfxJSONQuery_Find(t *testing.T) {
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
		{name: "no expression", args: args{
			expr: "",
			b:    []byte("{}"),
		}, want: nil, wantErr: true},
		{name: "no jsonb", args: args{
			expr: "data",
			b:    []byte(""),
		}, want: nil, wantErr: true},
		{name: "expression points to nothing", args: args{
			expr: "data",
			b:    []byte(`{"a": "b"}`),
		}, want: nil, wantErr: true},
		{name: "successful fetch data #1", args: args{
			expr: "expensive",
			b:    jsonBytes,
		}, want: any(float64(10)), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr: "/store/book/*[1]/price",
			b:    jsonBytes,
		}, want: any(8.95), wantErr: false},
		{name: "successful fetch data #4", args: args{
			expr: "/store/bicycle",
			b:    jsonBytes,
		}, want: any(map[string]any{"color": "red", "price": float64(19.95)}), wantErr: false},
		{name: "successful fetch data #1", args: args{
			expr: "/store/book/*[1]/category",
			b:    jsonBytes,
		}, want: any("reference"), wantErr: false},
		{name: "successful fetch data #2", args: args{
			expr: "/store/book/*[2]/price",
			b:    jsonBytes,
		}, want: any(float64(12.99)), wantErr: false},
		{name: "successful fetch data #3", args: args{
			expr: "/store/book/*[2]/available",
			b:    jsonBytes,
		}, want: any(true), wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AntchfxJSONQueryFinder{}
			got, err := a.Find(tt.args.expr, tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() got = %v, want %v", got, tt.want)
			}
		})
	}
}
