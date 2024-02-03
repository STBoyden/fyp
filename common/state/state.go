package state

import (
	"encoding/json"
	"fmt"
	typedsockets "fyp/common/utils/net/typed-sockets"
)

type serverPing string

const SERVER_PING serverPing = "ping"

type State struct {
	ClientUdpPort string `json:"client_udp_port,omitempty"`
	ClientReady   bool   `json:"client_is_ready,omitempty"`

	ServerPing serverPing `json:"ping_message,omitempty"`
}

// Check that `State` corrrectly implements `typedsockets.Convertable`
var _ typedsockets.Convertable = State{}
var _ fmt.Stringer = State{}

func EmptyState() State {
	return State{}
}

// Turn the current `State` and marshal it into JSON
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
