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

var UnknownClientID = uuid.NullUUID{Valid: false}

type playerFields struct {
	Name  string        `json:"name,omitempty"`
	Inner ctypes.Player `json:"inner,omitempty"`
}

type clientFields struct {
	UDPPort         string              `json:"udp_port,omitempty"`
	ID              uuid.NullUUID       `json:"id,omitempty"`
	Slot            int                 `json:"slot,omitempty"`
	InitialPosition ctypes.Position     `json:"initial_position,omitempty"`
	Colour          ctypes.PlayerColour `json:"player_colour"`
	Player          playerFields        `json:"player,omitempty"`
}

type serverFields struct {
	Players  map[string]ctypes.Player `json:"players,omitempty"`
	UpdateID int                      `json:"update_id,omitempty"`
}

/*
state.State is a common object that is shared between both the client and the server over
the TCP and UDP connections. state.State implements typedscockets.Convertable in the case
that the specific (de)serialisation format is changed in the future without changing the
state.State API.
*/
type State struct {
	Message    Message      `json:"message"`
	Submessage Submessage   `json:"sub_message"`
	Client     clientFields `json:"client,omitempty"`
	Server     serverFields `json:"server,omitempty"`
}

func WithUpdatedPlayers(serverUpdateID int, playersMap map[string]ctypes.Player) State {
	return State{
		Message:    Messages.FROM_SERVER,
		Submessage: Submessages.SERVER_UPDATING_PLAYERS,
		Server: serverFields{
			UpdateID: serverUpdateID,
			Players:  playersMap,
		},
	}
}

/*
WithUpdatedPlayerState returns a state.State that contains the player data from
ctypes.Player so that it can be used to update the server's version of this client.
*/
func WithUpdatedPlayerState(clientID uuid.NullUUID, playerState ctypes.Player) State {
	clientPlayer := playerFields{
		Name:  playerState.PlayerSpriteIndex.String(),
		Inner: playerState,
	}

	return State{
		Message:    Messages.FROM_CLIENT,
		Submessage: Submessages.CLIENT_SENDING_LOCAL_DATA,
		Client: clientFields{
			ID:     clientID,
			Player: clientPlayer,
		},
	}
}

func WithClientUDPPort(clientUDPPort string) State {
	return State{
		Message:    Messages.FROM_CLIENT,
		Submessage: Submessages.CLIENT_SENDING_UDP_PORT,
		Client: clientFields{
			UDPPort: clientUDPPort,
		},
	}
}

func WithClientDisconnecting(clientID uuid.NullUUID, playerName string) State {
	return State{
		Message:    Messages.FROM_CLIENT,
		Submessage: Submessages.CLIENT_DISCONNECTING,
		Client: clientFields{
			ID:     clientID,
			Player: playerFields{Name: playerName},
		},
	}
}

func WithNewClientConnection(clientID uuid.UUID, slot int) State {
	return State{
		Message:    Messages.FROM_SERVER,
		Submessage: Submessages.SERVER_FIRST_CLIENT_CONNECTION_INFORMATION,
		Client: clientFields{
			Slot:            slot,
			ID:              uuid.NullUUID{UUID: clientID, Valid: true},
			InitialPosition: ctypes.NewPosition(100, 100),
			Colour:          ctypes.PlayerColourFromInt(slot),
		},
	}
}

func WithClientReady(clientID uuid.UUID, player ctypes.Player) State {
	return State{
		Message:    Messages.FROM_CLIENT,
		Submessage: Submessages.CLIENT_READY,
		Client: clientFields{
			ID: uuid.NullUUID{UUID: clientID, Valid: true},
			Player: playerFields{
				Name:  player.PlayerSpriteIndex.String(),
				Inner: player,
			},
		},
	}
}

func WithServerPing() State {
	return State{
		Message:    Messages.FROM_SERVER,
		Submessage: Submessages.SERVER_PING,
	}
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
	return State{Server: serverFields{Players: make(map[string]ctypes.Player)}}
}

func (s *State) IsEmpty() bool {
	return (s.Submessage == Submessages.SUBMESSAGE_NONE &&
		s.Message == Messages.MESSAGE_NONE)
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
