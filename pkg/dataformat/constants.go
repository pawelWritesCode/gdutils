package dataformat

const (
	// FormatJSON describes JSON data format.
	FormatJSON DataFormat = "JSON"

	// FormatPlainText describes plan text data format.
	FormatPlainText DataFormat = "plain text"
)

// DataFormat describes format of data.
type DataFormat string
