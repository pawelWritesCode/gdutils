package httpctx

import "testing"

func TestURLValidator_Validate(t *testing.T) {
	type args struct {
		in any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "not string", args: args{in: 1}, wantErr: true},
		{name: "valid URL #1", args: args{in: "https://www.example.com"}, wantErr: false},
		{name: "valid URL #2", args: args{in: "http://www.xxxaaa.com/user.json"}, wantErr: false},
		{name: "invalid URL #1", args: args{in: "example.com/hehe.json"}, wantErr: true},
		{name: "invalid URL #2", args: args{in: "www.google.com"}, wantErr: true},
		{name: "invalid URL #3", args: args{in: "www.example.com/hehe/xd"}, wantErr: true},
		{name: "invalid URL #4", args: args{in: "/usr/bin/local"}, wantErr: true},
		{name: "invalid URL #4", args: args{in: "wsad"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			U := URLValidator{}
			if err := U.Validate(tt.args.in); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
