package mathutils

import "testing"

func Test_RandomInt(t *testing.T) {
	type args struct {
		from int
		to   int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "zero to one", args: args{
			from: 1,
			to:   0,
		}, want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomInt(tt.args.from, tt.args.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RandomInt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
