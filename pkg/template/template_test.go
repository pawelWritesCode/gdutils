package template

import (
	"fmt"
	"testing"
	"time"
)

func TestTemplateManager_Replace(t *testing.T) {
	tm, err := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Feb 4, 2014 at 6:05pm (PST)")
	if err != nil {
		t.Errorf("invalid time parsing")
	}
	fmt.Println(tm)

	type args struct {
		templateValue string
		storage       map[string]any
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
			storage:       map[string]any{},
		}, want: "abc", wantErr: false},
		{name: "simple string without template value #2", args: args{
			templateValue: "abc",
			storage:       map[string]any{"abc": "def"},
		}, want: "abc", wantErr: false},
		{name: "simple string with missing template value in storage", args: args{
			templateValue: "abc {{.NAME}}",
			storage:       map[string]any{},
		}, want: "", wantErr: true},
		{name: "simple string with template value #2", args: args{
			templateValue: "abc {{.NAME}}",
			storage:       map[string]any{"NAME": "xyz"},
		}, want: "abc xyz", wantErr: false},
		{name: "missing at least one template value in storage #1", args: args{
			templateValue: "abc {{.NAME}}",
			storage:       map[string]any{"OTHER": "xyz"},
		}, want: "", wantErr: true},
		{name: "missing at least one template value in storage #2", args: args{
			templateValue: "abc {{.NAME}} {{.LAST_NAME}}",
			storage:       map[string]any{"NAME": "xyz"},
		}, want: "", wantErr: true},
		{name: "multi template string #1", args: args{
			templateValue: "abc {{.NAME}} {{.LAST_NAME}}",
			storage:       map[string]any{"NAME": "xyz", "LAST_NAME": "xxx"},
		}, want: "abc xyz xxx", wantErr: false},
		{name: "multi template string #1", args: args{
			templateValue: "{{.NAME}} abc {{.LAST_NAME}}",
			storage:       map[string]any{"NAME": true, "LAST_NAME": 111},
		}, want: "true abc 111", wantErr: false},
		{name: "key with dot #1", args: args{
			templateValue: "{{.RANDOM_USER.TOKEN}}",
			storage:       map[string]any{"RANDOM_USER": map[string]string{"TOKEN": "a.b.c"}},
		}, want: "a.b.c", wantErr: false},
		{name: "key with dot #2", args: args{
			templateValue: "{{.RANDOM_USER.TOKEN}}",
			storage: map[string]any{"RANDOM_USER": struct {
				TOKEN string
			}{TOKEN: "a.b.c"}},
		}, want: "a.b.c", wantErr: false},
		{name: "key with dot #3", args: args{
			templateValue: "{{.RANDOM_USER.Token}}",
			storage: map[string]any{"RANDOM_USER": struct {
				Token string
			}{Token: "a.b.c"}},
		}, want: "a.b.c", wantErr: false},
		{name: "use function on template #1", args: args{
			templateValue: `{{.TIME.Format "2006-01-02T15:04:05Z07:00"}}`,
			storage:       map[string]any{"TIME": tm},
		}, want: "2014-02-04T18:05:00Z", wantErr: false},
		{name: "use function on template #2", args: args{
			templateValue: `{{.TIME.Format "Jan 2, 2006 at 3:04pm (MST)"}}`,
			storage:       map[string]any{"TIME": tm},
		}, want: "Feb 4, 2014 at 6:05pm (PST)", wantErr: false},
		{name: "use function on template #3", args: args{
			templateValue: `{{.TIME.Format "2006-01-02"}}`,
			storage:       map[string]any{"TIME": tm},
		}, want: "2014-02-04", wantErr: false},
		{name: "use function on template #4", args: args{
			templateValue: `{{.TIME}}`,
			storage:       map[string]any{"TIME": tm},
		}, want: "2014-02-04 18:05:00 +0000 PST", wantErr: false},
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
