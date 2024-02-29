package models

import (
	"sync"

	"fyp/src/common/ctypes"
	"fyp/src/common/ctypes/state"
)

type ServerState struct {
	mutex          sync.RWMutex
	state          state.State
	updatedChannel <-chan string
}

func NewServerState() (s *ServerState, u <-chan string) {
	updatedChannel := make(chan string, 1024)
	return &ServerState{mutex: sync.RWMutex{}, updatedChannel: updatedChannel, state: state.Empty()}, updatedChannel
}

func (s *ServerState) AddPlayer(name string, player ctypes.Player) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state.Players[name] = player
}

func (s *ServerState) GetPlayers() map[string]ctypes.Player {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.state.Players
}
