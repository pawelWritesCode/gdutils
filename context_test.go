package gdutils

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/pawelWritesCode/gdutils/pkg/cache"
	"github.com/pawelWritesCode/gdutils/pkg/debugger"
	"github.com/pawelWritesCode/gdutils/pkg/pathfinder"
	"github.com/pawelWritesCode/gdutils/pkg/schema"
	"github.com/pawelWritesCode/gdutils/pkg/serializer"
	"github.com/pawelWritesCode/gdutils/pkg/template"
)

type newDebugger struct{}

type newCache struct{}

type newClient struct{}

type newTemplateEngine struct{}

type newStringValidator struct{}

type newPathFinder struct{}

type newSerializer struct{}

func (n newSerializer) Deserialize(data []byte, v any) error {
	panic("implement me")
}

func (n newSerializer) Serialize(v any) ([]byte, error) {
	panic("implement me")
}

func (n newPathFinder) Find(expr string, bytes []byte) (any, error) {
	panic("implement me")
}

func (n newStringValidator) Validate(document, schemaPath string) error {
	panic("implement me")
}

func (n newTemplateEngine) Replace(templateValue string, storage map[string]any) (string, error) {
	panic("implement me")
}

func (n newClient) Do(req *http.Request) (*http.Response, error) {
	panic("implement me")
}

func (n newCache) Save(key string, value any) {
	panic("implement me")
}

func (n newCache) GetSaved(key string) (any, error) {
	panic("implement me")
}

func (n newCache) Reset() {
	panic("implement me")
}

func (n newCache) All() map[string]any {
	panic("implement me")
}

func (n newDebugger) Print(info string) {
	panic("implement me")
}

func (n newDebugger) IsOn() bool {
	panic("implement me")
}

func (n newDebugger) TurnOn() {
	panic("implement me")
}

func (n newDebugger) TurnOff() {
	panic("implement me")
}

func (n newDebugger) Reset(isOn bool) {
	panic("implement me")
}

func TestState_ResetState(t *testing.T) {
	s := NewDefaultAPIContext(true, "")
	s.Cache.Save("test", 1)

	s.ResetState(false)

	if s.Debugger.IsOn() != false {
		t.Errorf("IsDebug property did not change")
	}

	if !reflect.DeepEqual(s.Cache.All(), map[string]any{}) {
		t.Errorf("cache did not reset")
	}
}

func TestState_SetDebugger(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefaultDebugger := s.Debugger.(*debugger.DebuggerService)
	if !isDefaultDebugger {
		t.Errorf("default debugger should be *debugger.DebuggerService")
	}

	s.SetDebugger(newDebugger{})

	_, isNewDebugger := s.Debugger.(newDebugger)
	if !isNewDebugger {
		t.Errorf("SetDebugger does not work properly")
	}
}

func TestState_SetCache(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefaultCache := s.Cache.(*cache.ConcurrentCache)
	if !isDefaultCache {
		t.Errorf("default cache should be *cache.ConcurrentCache")
	}

	s.SetCache(newCache{})

	_, isNewCache := s.Cache.(newCache)
	if !isNewCache {
		t.Errorf("SetCache does not work properly")
	}
}

func TestState_SetRequestDoer(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefaultHttpCli := s.RequestDoer.(*http.Client)
	if !isDefaultHttpCli {
		t.Errorf("default request doer is not *http.Client")
	}

	s.SetRequestDoer(newClient{})

	_, isNewClient := s.RequestDoer.(newClient)
	if !isNewClient {
		t.Errorf("SetRequestDoer does not work properly")
	}
}

func TestState_SetTemplateEngine(t *testing.T) {
	s := NewDefaultAPIContext(false, "")
	_, isDefault := s.TemplateEngine.(template.TemplateManager)
	if !isDefault {
		t.Errorf("default TemplateEngine is not template.TemplateManager")
	}

	s.SetTemplateEngine(newTemplateEngine{})

	_, isNewTemplateEngine := s.TemplateEngine.(newTemplateEngine)
	if !isNewTemplateEngine {
		t.Errorf("SetTemplateEngine does not work proplerly")
	}
}

func TestState_SetSchemaStringValidator(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefault := s.SchemaValidators.StringValidator.(schema.JSONSchemaRawXGValidator)
	if !isDefault {
		t.Errorf("default StringValidator is not schema.JSONSchemaRawXGValidator")
	}

	s.SetSchemaStringValidator(newStringValidator{})

	_, isNewStringValidator := s.SchemaValidators.StringValidator.(newStringValidator)
	if !isNewStringValidator {
		t.Errorf("SetSchemaStringValidator does not work properly")
	}
}

func TestState_SetSchemaReferenceValidator(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefault := s.SchemaValidators.ReferenceValidator.(schema.JSONSchemaReferenceXGValidator)
	if !isDefault {
		t.Errorf("default ReferenceValidator is not schema.JSONSchemaReferenceXGValidator")
	}

	s.SetSchemaReferenceValidator(newStringValidator{})

	_, isNewReferenceValidator := s.SchemaValidators.ReferenceValidator.(newStringValidator)
	if !isNewReferenceValidator {
		t.Errorf("SetSchemaReferenceValidator does not work properly")
	}
}

func TestState_SetJSONPathFinder(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefault := s.PathFinders.JSON.(*pathfinder.DynamicJSONPathFinder)
	if !isDefault {
		t.Errorf("default JSON PathFinder is not pathfinder.DynamicJSONPathFinder")
	}

	s.SetJSONPathFinder(newPathFinder{})

	_, isNewPathFinder := s.PathFinders.JSON.(newPathFinder)
	if !isNewPathFinder {
		t.Errorf("SetJSONPathFinder does not work properly")
	}
}

func TestState_SetYAMLPathFinder(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefault := s.PathFinders.YAML.(pathfinder.GoccyGoYamlFinder)
	if !isDefault {
		t.Errorf("default YAML PathFinder is not pathfinder.GoccyGoYamlFinder")
	}

	s.SetYAMLPathFinder(newPathFinder{})

	_, isNewPathFinder := s.PathFinders.YAML.(newPathFinder)
	if !isNewPathFinder {
		t.Errorf("SetYAMLPathFinder does not work properly")
	}
}

func TestState_SetXMLPathFinder(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefault := s.PathFinders.XML.(pathfinder.AntchfxXMLFinder)
	if !isDefault {
		t.Errorf("default XML PathFinder is not pathfinder.AntchfxXMLFinder")
	}

	s.SetXMLPathFinder(newPathFinder{})

	_, isNewPathFinder := s.PathFinders.XML.(newPathFinder)
	if !isNewPathFinder {
		t.Errorf("SetXMLPathFinder does not work properly")
	}
}

func TestState_SetJSONSerializer(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefault := s.Serializers.JSON.(serializer.JSON)
	if !isDefault {
		t.Errorf("default JSON Formatter is not formatter.JSONFormatter")
	}

	s.SetJSONSerializer(newSerializer{})

	_, isNewFormatter := s.Serializers.JSON.(newSerializer)
	if !isNewFormatter {
		t.Errorf("SetJSONSerializer does not work properly")
	}
}

func TestState_SetYAMLSerializer(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefault := s.Serializers.YAML.(serializer.YAML)
	if !isDefault {
		t.Errorf("default JSON Formatter is not formatter.YAMLFormatter")
	}

	s.SetYAMLSerializer(newSerializer{})

	_, isNewFormatter := s.Serializers.YAML.(newSerializer)
	if !isNewFormatter {
		t.Errorf("SetYAMLSerializer does not work properly")
	}
}

func TestState_SetXMLSerializer(t *testing.T) {
	s := NewDefaultAPIContext(false, "")

	_, isDefault := s.Serializers.XML.(serializer.XML)
	if !isDefault {
		t.Errorf("default XML Formatter is not formatter.XMLFormatter")
	}

	s.SetXMLSerializer(newSerializer{})

	_, isNewFormatter := s.Serializers.XML.(newSerializer)
	if !isNewFormatter {
		t.Errorf("SetXMLSerializer does not work properly")
	}
}
