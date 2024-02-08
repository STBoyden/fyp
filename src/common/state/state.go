package state

import (
	"encoding/json"
	"fmt"

	typedsockets "fyp/src/common/utils/net/typed-sockets"
)

type serverPing string

const ServerPing serverPing = "ping"

type State struct {
	ClientUDPPort string `json:"client_udp_port,omitempty"`
	ClientReady   bool   `json:"client_is_ready,omitempty"`

	ServerPing    serverPing `json:"ping_message,omitempty"`
	ServerMessage string     `json:"server_message,omitempty"`
}

// Check that `State` corrrectly implements `typedsockets.Convertable` and `fmt.Stringer`.
var (
	_ typedsockets.Convertable = State{}
	_ fmt.Stringer             = State{}
)

func Empty() State {
	return State{}
}

// Turn the current `State` and marshal it into JSON.
func (s State) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

// Overwrites the current `State` with the JSON `data`.
func (s State) Unmarshal(data []byte) (bool, error) {
	ptr := &s
	err := json.Unmarshal(data, ptr)

	return ptr == nil, err
}

func (s State) String() string {
	data, err := s.Marshal()
	if err != nil {
		panic(err)
	}

	return string(data)
}
