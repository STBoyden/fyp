package state

import (
	"encoding/json"
	"errors"
	"fmt"

	"fyp/src/common/ctypes"
	typedsockets "fyp/src/common/utils/net/typed-sockets"

	"github.com/google/uuid"
)

type serverPing string

const (
	ServerPing serverPing = "ping"
)

var UnknownClientID = uuid.NullUUID{Valid: false}

type State struct {
	ClientUDPPort       string        `json:"client_udp_port,omitempty"`
	ClientReady         bool          `json:"client_is_ready,omitempty"`
	ClientID            uuid.NullUUID `json:"client_id,omitempty"`
	ClientSlot          int           `json:"client_slot,omitempty"`
	ClientDisconnecting bool          `json:"client_is_disconnecting,omitempty"`

	ServerPing    serverPing               `json:"ping_message,omitempty"`
	ServerMessage string                   `json:"server_message,omitempty"`
	Players       map[string]ctypes.Player `json:"players,omitempty"`
}

// Check that `State` corrrectly implements `typedsockets.Convertable` and `fmt.Stringer`.
var (
	_ typedsockets.Convertable = State{}
	_ fmt.Stringer             = State{}
)

func Empty() State {
	return State{Players: make(map[string]ctypes.Player)}
}

func (s *State) IsEmpty() bool {
	return (s.ClientUDPPort == "" &&
		!s.ClientReady &&
		s.ServerPing == "" &&
		s.ServerMessage == "" &&
		s.ClientID == UnknownClientID &&
		s.Players == nil)
}

// Turn the current `State` and marshal it into JSON.
func (s State) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

// Overwrites the current `State` with the JSON `data`.
func (State) Unmarshal(v any, data []byte) error {
	if v, ok := v.(*State); ok {
		err := json.Unmarshal(data, &v)

		return err
	}

	if _, ok := v.(State); ok {
		return errors.New("this unmarshal function requires *State, got State instead")
	}

	return errors.New("this unmarshal function is not for this type")
}

func (s State) String() string {
	data, err := s.Marshal()
	if err != nil {
		panic(err)
	}

	return string(data)
}
