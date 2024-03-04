/*
state provides the State object that is used by both cmd/client and cmd/server.
*/
package state

/*
TODO: Finish documentation.
*/

import (
	"encoding/json"
	"errors"
	"fmt"

	"fyp/src/common/ctypes"
	typedsockets "fyp/src/common/utils/net/typed-sockets"

	"github.com/google/uuid"
)

const (
	serverPing = "ping"
)

var UnknownClientID = uuid.NullUUID{Valid: false}

/*
state.State is a common object that is shared between both the client and the server over
the TCP and UDP connections. state.State implements typedscockets.Convertable in the case
that the specific (de)serialisation format is changed in the future without changing the
state.State API.
*/
type State struct {
	ClientUDPPort       string        `json:"client_udp_port,omitempty"`
	ClientReady         bool          `json:"client_is_ready,omitempty"`
	ClientID            uuid.NullUUID `json:"client_id,omitempty"`
	ClientSlot          int           `json:"client_slot,omitempty"`
	ClientDisconnecting bool          `json:"client_is_disconnecting,omitempty"`

	ServerPing    string                   `json:"ping_message,omitempty"`
	ServerMessage string                   `json:"server_message,omitempty"`
	Players       map[string]ctypes.Player `json:"players,omitempty"`
}

/*
WithUpdatedPlayerState returns a state.State that contains the player data from
ctypes.Player so that it can be used to update the server's version of this client.
*/
func WithUpdatedPlayerState(clientID uuid.NullUUID, playerState ctypes.Player) State {
	players := make(map[string]ctypes.Player)
	players[playerState.PlayerSpriteIndex.String()] = playerState

	return State{ClientID: clientID, Players: players}
}

func WithClientUDPPort(clientUDPPort string) State {
	return State{ClientUDPPort: clientUDPPort}
}

func WithClientDisconnecting(clientID uuid.NullUUID) State {
	return State{ClientID: clientID, ClientDisconnecting: true}
}

func WithNewClientConnection(newClientID uuid.UUID, clientSlot int) State {
	return State{
		ServerMessage: "Hello, world!",
		ClientSlot:    clientSlot,
		ClientID:      uuid.NullUUID{UUID: newClientID, Valid: true},
	}
}

func WithClientReady() State {
	return State{ClientReady: true}
}

func WithServerPing() State {
	return State{ServerPing: serverPing}
}

// Check that `State` corrrectly implements `typedsockets.Convertable` and `fmt.Stringer`.
var (
	_ typedsockets.Convertable = State{}
	_ fmt.Stringer             = State{}
)

/*
Type aliases for typedsockets types to reduce code noise.
*/
type (
	UDPConnection     = typedsockets.UDPTypedConnection[State]
	TCPConnection     = typedsockets.TCPTypedConnection[State]
	TCPSocketListener = typedsockets.TCPSocketListener[State]
)

/*
Empty returns an empty State with a non-nil Players field.
*/
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

// Marshal is a wrapper over json.Marshal to implement typedsockets.Convertable.
func (s State) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

/*
Unmarshal is a wrapper over json.Unmarshal that checks the type of v to only be of
type *State.
*/
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
