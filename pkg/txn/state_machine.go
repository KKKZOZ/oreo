package txn

import (
	"errors"
	"fmt"

	"github.com/kkkzoz/oreo/pkg/config"
)

type StateMachine struct {
	state config.State
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		state: config.EMPTY,
	}
}

func (st *StateMachine) SetState(state config.State) error {
	switch state {
	case config.STARTED:
		if st.state != config.EMPTY {
			return errors.New("transaction can't be started as it is not in an empty state")
		}
	case config.COMMITTED:
		if st.state != config.STARTED {
			return errors.New("transaction can only be committed from a started state")
		}
	case config.ABORTED:
		if st.state == config.EMPTY {
			return errors.New("transaction can't be aborted as it hasn't been started yet")
		}
	default:
		return fmt.Errorf("attempted to transition to invalid state %v", state)
	}

	st.state = state
	return nil
}

func (st *StateMachine) CheckState(state config.State) error {
	if st.state != state {
		switch state {
		case config.EMPTY:
			return errors.New("the transaction hasn't been started yet")
		case config.STARTED:
			return errors.New("the transaction is not currently in progress")
		case config.COMMITTED:
			return errors.New("the transaction has not been committed yet")
		case config.ABORTED:
			return errors.New("the transaction hasn't been aborted")
		default:
			return fmt.Errorf("the transaction isn't in the expected state %v", state)
		}
	}
	return nil
}

func (st *StateMachine) GetState() config.State {
	return st.state
}
