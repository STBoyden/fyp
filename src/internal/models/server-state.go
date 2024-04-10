package models

import (
	"sync"

	"fyp/src/common/ctypes"
	"fyp/src/common/ctypes/state"
)

type ServerState struct {
	mutex          sync.RWMutex
	state          state.State
	updatedChannel chan<- string
}

func NewServerState() (s *ServerState, u <-chan string) {
	updatedChannel := make(chan string, 1024)
	return &ServerState{
		mutex:          sync.RWMutex{},
		updatedChannel: updatedChannel,
		state:          state.Empty(),
	}, updatedChannel
}

func (s *ServerState) AddPlayer(name string, player ctypes.Player) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state.Server.Players[name] = player
	s.updatedChannel <- "added player"
}

func (s *ServerState) RemovePlayer(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.state.Server.Players, name)
	s.updatedChannel <- "removed player"
}

func (s *ServerState) ContainsPlayer(name string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, ok := s.state.Server.Players[name]

	return ok
}

func (s *ServerState) UpdatePlayer(name string, data ctypes.Player) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.state.Server.Players[name]; ok {
		s.state.Server.Players[name] = data
		s.updatedChannel <- "updated player"
	}
}

func (s *ServerState) FilterPlayers(filter func(key string, player ctypes.Player) bool) map[string]ctypes.Player {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	filtered := make(map[string]ctypes.Player)
	for key, player := range s.state.Server.Players {
		if !filter(key, player) {
			continue
		}

		filtered[key] = player
	}

	return filtered
}

func (s *ServerState) GetPlayers() map[string]ctypes.Player {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.state.Server.Players
}

func (s *ServerState) String() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.state.String()
}

func (s *ServerState) Copy() state.State {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.state
}
