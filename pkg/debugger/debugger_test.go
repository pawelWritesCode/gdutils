package debugger

import "testing"

func TestDebuggerService_IsOn(t *testing.T) {
	d := New(false)
	if d.IsOn() {
		t.Errorf("IsOn should be false")
	}
}

func TestDebuggerService_TurnOn(t *testing.T) {
	d := New(false)

	if d.IsOn() {
		t.Errorf("IsOn should be false")
	}

	d.TurnOn()

	if !d.IsOn() {
		t.Errorf("IsOn should be true")
	}
}

func TestDebuggerService_TurnOff(t *testing.T) {
	d := New(false)

	if d.IsOn() {
		t.Errorf("IsOn should be false")
	}

	d.TurnOn()

	if !d.IsOn() {
		t.Errorf("IsOn should be true")
	}

	d.TurnOff()

	if d.IsOn() {
		t.Errorf("IsOn should be false again")
	}
}

func TestDebuggerService_Reset(t *testing.T) {
	d := New(false)

	if d.IsOn() {
		t.Errorf("IsOn should be false")
	}

	d.TurnOn()

	if !d.IsOn() {
		t.Errorf("IsOn should be true")
	}

	d.Reset(false)
	if d.IsOn() {
		t.Errorf("IsOn should be false again")
	}
}
