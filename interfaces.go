package gdutils

import (
	"net/http"

	"github.com/pawelWritesCode/gdutils/pkg/osutils"
	"github.com/pawelWritesCode/gdutils/pkg/types"
)

// cacheable represents ability to store/retrieve arbitrary values.
type cacheable interface {
	// Save preserve provided value under given key.
	Save(key string, value any)

	// GetSaved retrieve value from given key.
	GetSaved(key string) (any, error)

	// Reset turns cache into init state - clears all entries.
	Reset()

	// All returns all cache entries.
	All() map[string]any
}

// debuggable defines desired debugger behaviour.
type debuggable interface {
	// Print prints provided info.
	Print(info string)

	// IsOn tells whether debugging mode is activated.
	IsOn() bool

	// TurnOn turns on debugging mode.
	TurnOn()

	// TurnOff turns off debugging mode.
	TurnOff()

	// Reset resets debugging mode to init state.
	Reset(isOn bool)
}

// serializable describes ability to serialize and deserialize data
type serializable interface {
	// Deserialize deserializes data on v
	Deserialize(data []byte, v any) error

	// Serialize serializes v
	Serialize(v any) ([]byte, error)
}

// requestDoer describes ability to make HTTP(s) requests.
type requestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// fileRecognizer describes entity that has ability to find file reference in input
type fileRecognizer interface {
	// Recognize recognizes file reference in provided input
	Recognize(input string) (osutils.FileReference, bool)
}

// pathFinder describes ability to obtain node(s) from data in fixed data format
type pathFinder interface {
	// Find obtains data from bytes according to given expression
	Find(expr string, bytes []byte) (any, error)
}

// templateEngine is entity that has ability to work with templates.
type templateEngine interface {
	// Replace replaces template values using provided storage.
	Replace(templateValue string, storage map[string]any) (string, error)
}

// typeMapper represents entity that has ability to map data's type into corresponding types.DataType of given format.
type typeMapper interface {
	// Map maps data type.
	Map(data any) types.DataType
}
