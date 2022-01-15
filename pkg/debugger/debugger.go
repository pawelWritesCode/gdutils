// Package debugger holds definition of Debugger.
package debugger

import "fmt"

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
}

func New(isOn bool) *DebuggerService {
	return &DebuggerService{actualState: isOn}
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
	fmt.Printf("%s: %s\n", "debug", info)
}
