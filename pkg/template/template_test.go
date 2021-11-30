package template

import "testing"

func TestTemplateManager_Replace(t *testing.T) {
	type args struct {
		templateValue string
		storage       map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "error when nil storage", args: args{
			templateValue: "{{.NAME}}",
			storage:       nil,
		}, want: "", wantErr: true},
		{name: "simple string without template value #1", args: args{
			templateValue: "abc",
			storage:       map[string]interface{}{},
		}, want: "abc", wantErr: false},
		{name: "simple string without template value #2", args: args{
			templateValue: "abc",
			storage:       map[string]interface{}{"abc": "def"},
		}, want: "abc", wantErr: false},
		{name: "simple string with missing template value in storage", args: args{
			templateValue: "abc {{.NAME}}",
			storage:       map[string]interface{}{},
		}, want: "", wantErr: true},
		{name: "simple string with template value #2", args: args{
			templateValue: "abc {{.NAME}}",
			storage:       map[string]interface{}{"NAME": "xyz"},
		}, want: "abc xyz", wantErr: false},
		{name: "missing at least one template value in storage #1", args: args{
			templateValue: "abc {{.NAME}}",
			storage:       map[string]interface{}{"OTHER": "xyz"},
		}, want: "", wantErr: true},
		{name: "missing at least one template value in storage #2", args: args{
			templateValue: "abc {{.NAME}} {{.LAST_NAME}}",
			storage:       map[string]interface{}{"NAME": "xyz"},
		}, want: "", wantErr: true},
		{name: "multi template string #1", args: args{
			templateValue: "abc {{.NAME}} {{.LAST_NAME}}",
			storage:       map[string]interface{}{"NAME": "xyz", "LAST_NAME": "xxx"},
		}, want: "abc xyz xxx", wantErr: false},
		{name: "multi template string #1", args: args{
			templateValue: "{{.NAME}} abc {{.LAST_NAME}}",
			storage:       map[string]interface{}{"NAME": true, "LAST_NAME": 111},
		}, want: "true abc 111", wantErr: false},
		{name: "key with dot #1", args: args{
			templateValue: "{{.RANDOM_USER.TOKEN}}",
			storage:       map[string]interface{}{"RANDOM_USER": map[string]string{"TOKEN": "a.b.c"}},
		}, want: "a.b.c", wantErr: false},
		{name: "key with dot #2", args: args{
			templateValue: "{{.RANDOM_USER.TOKEN}}",
			storage: map[string]interface{}{"RANDOM_USER": struct {
				TOKEN string
			}{TOKEN: "a.b.c"}},
		}, want: "a.b.c", wantErr: false},
		{name: "key with dot #3", args: args{
			templateValue: "{{.RANDOM_USER.Token}}",
			storage: map[string]interface{}{"RANDOM_USER": struct {
				Token string
			}{Token: "a.b.c"}},
		}, want: "a.b.c", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := TemplateManager{}
			got, err := tm.Replace(tt.args.templateValue, tt.args.storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("Replace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Replace() got = %v, want %v", got, tt.want)
			}
		})
	}
}
