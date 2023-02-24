package taproot

import (
	"sync"
)

type serverState int8

const (
	SERVER_STATE_UNKNOWN      serverState = 0
	SERVER_STATE_INITIALIZING serverState = 1
	SERVER_STATE_RUNNING      serverState = 2
	SERVER_STATE_CLOSING      serverState = 3
	SERVER_STATE_CLOSED       serverState = 4
)

type serverStateManager struct {
	sync.Mutex
	currentState serverState
}

func (s *serverStateManager) setState(state serverState) {
	s.Lock()
	s.currentState = state
	s.Unlock()
}
