package txn

import (
	"testing"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
)

func TestNewStateMachine(t *testing.T) {
	sm := NewStateMachine()

	if sm.GetState() != config.EMPTY {
		t.Errorf("Expected to start with empty state, received: %#v", sm.GetState())
	}
}

func TestSetInvalidState(t *testing.T) {
	sm := NewStateMachine()

	errorExpected := sm.SetState(123)
	if errorExpected == nil {
		t.Errorf("Expected error when setting to an invalid state, received no error")
	}
}

func TestSetState(t *testing.T) {
	sm := NewStateMachine()

	err := sm.SetState(config.STARTED)
	if err != nil {
		t.Errorf("Expected no error when setting state from EMPTY to STARTED, received: %s", err.Error())
	}

	err = sm.SetState(config.COMMITTED)
	if err != nil {
		t.Errorf("Expected no error when setting state from STARTED to COMMITTED, received: %s", err.Error())
	}
}

func TestSetStateWithErroneousTransitions(t *testing.T) {
	sm := NewStateMachine()

	err := sm.SetState(config.COMMITTED)
	if err == nil {
		t.Errorf("Expected error when setting state from EMPTY to COMMITTED, received no error")
	}

	sm.state = config.STARTED

	err = sm.SetState(config.STARTED)
	if err == nil {
		t.Errorf("Expected error when setting state from STARTED to STARTED, received no error")
	}
}

func TestCheckState(t *testing.T) {
	sm := NewStateMachine()

	err := sm.CheckState(config.EMPTY)
	if err != nil {
		t.Errorf("Expected no error when checking state EMPTY, received: %s", err.Error())
	}

	err = sm.CheckState(config.STARTED)
	if err == nil {
		t.Errorf("Expected error when checking state STARTED, received no error")
	}
}

func TestGetState(t *testing.T) {
	sm := NewStateMachine()
	sm.state = config.STARTED

	received := sm.GetState()
	if received != config.STARTED {
		t.Errorf("Expected to get state STARTED, received: %#v", received)
	}
}
