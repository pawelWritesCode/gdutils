// Package debugger holds definition of Debugger.
package debugger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/hokaccha/go-prettyjson"
	"github.com/pawelWritesCode/df"
)

// Debugger represents debugger.
type Debugger interface {
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

// DebuggerService is utility tool for debugging
type DebuggerService struct {

	// actualState tells whether debugger is on/off, true = on, false = off.
	actualState bool

	// isColored determines whether output will be colored
	isColored bool

	// limit is the maximum number of bytes to be printed.
	limit uint16

	// writer is place where output will be written.
	writer io.Writer
}

func New(isOn, isColored bool, bytesLimit uint16, writer io.Writer) *DebuggerService {
	return &DebuggerService{actualState: isOn, limit: bytesLimit, writer: writer}
}

// NewDefault returns *DebuggerService writing to stdOut with output limit up to 3072 bytes
func NewDefault(isOn bool) *DebuggerService {
	return &DebuggerService{actualState: isOn, limit: 3072, writer: os.Stdout, isColored: true}
}

// IsOn tells whether debugging mode is activated.
func (d *DebuggerService) IsOn() bool {
	return d.actualState
}

// TurnOn turns on debugging mode.
func (d *DebuggerService) TurnOn() {
	d.actualState = true
}

// TurnOff turns off debugging mode.
func (d *DebuggerService) TurnOff() {
	d.actualState = false
}

// Reset resets debugging mode to init state.
func (d *DebuggerService) Reset(isOn bool) {
	d.actualState = isOn
}

// Print prints provided info.
func (d *DebuggerService) Print(info string) {
	_, _ = fmt.Fprintln(d.writer, d.prepareMessage(info))
}

// prepareMessage makes few modifications to the message.
func (d *DebuggerService) prepareMessage(info string) string {
	var output = []byte(info)

	if df.IsJSON([]byte(info)) {
		var rm json.RawMessage
		_ = json.Unmarshal([]byte(info), &rm)

		if d.isColored {
			output, _ = prettyjson.Marshal(rm)
		} else {
			output, _ = json.MarshalIndent(rm, "", "\t")
		}
	}

	var bytesToPrint uint16
	if len(output) <= int(d.limit) {
		bytesToPrint = uint16(len(output))
	} else {
		bytesToPrint = d.limit
	}

	return string(output[:bytesToPrint])
}
