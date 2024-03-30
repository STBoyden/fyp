// Code generated by goenums. DO NOT EDIT.
// This file was generated by github.com/zarldev/goenums/cmd/goenums
// using the command:
// goenums filename.go

package state

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type Submessage struct {
	submessage
}

type submessageContainer struct {
	SUBMESSAGE_NONE                            Submessage
	CLIENT_SENDING_UDP_PORT                    Submessage
	CLIENT_READY                               Submessage
	CLIENT_SENDING_LOCAL_DATA                  Submessage
	CLIENT_DISCONNECTING                       Submessage
	SERVER_PING                                Submessage
	SERVER_FIRST_CLIENT_CONNECTION_INFORMATION Submessage
	SERVER_UPDATING_PLAYERS                    Submessage
}

var Submessages = submessageContainer{
	SUBMESSAGE_NONE: Submessage{
		submessage: submessage_none,
	},
	CLIENT_SENDING_UDP_PORT: Submessage{
		submessage: client_sending_udp_port,
	},
	CLIENT_READY: Submessage{
		submessage: client_ready,
	},
	CLIENT_SENDING_LOCAL_DATA: Submessage{
		submessage: client_sending_local_data,
	},
	CLIENT_DISCONNECTING: Submessage{
		submessage: client_disconnecting,
	},
	SERVER_PING: Submessage{
		submessage: server_ping,
	},
	SERVER_FIRST_CLIENT_CONNECTION_INFORMATION: Submessage{
		submessage: server_first_client_connection_information,
	},
	SERVER_UPDATING_PLAYERS: Submessage{
		submessage: server_updating_players,
	},
}

func (c submessageContainer) All() []Submessage {
	return []Submessage{
		c.SUBMESSAGE_NONE,
		c.CLIENT_SENDING_UDP_PORT,
		c.CLIENT_READY,
		c.CLIENT_SENDING_LOCAL_DATA,
		c.CLIENT_DISCONNECTING,
		c.SERVER_PING,
		c.SERVER_FIRST_CLIENT_CONNECTION_INFORMATION,
		c.SERVER_UPDATING_PLAYERS,
	}
}

var invalidSubmessage = Submessage{}

func ParseSubmessage(a any) Submessage {
	switch v := a.(type) {
	case Submessage:
		return v
	case string:
		return stringToSubmessage(v)
	case fmt.Stringer:
		return stringToSubmessage(v.String())
	case int:
		return intToSubmessage(v)
	case int64:
		return intToSubmessage(int(v))
	case int32:
		return intToSubmessage(int(v))
	}
	return invalidSubmessage
}

func stringToSubmessage(s string) Submessage {
	lwr := strings.ToLower(s)
	switch lwr {
	case "submessage_none":
		return Submessages.SUBMESSAGE_NONE
	case "client_sending_udp_port":
		return Submessages.CLIENT_SENDING_UDP_PORT
	case "client_ready":
		return Submessages.CLIENT_READY
	case "client_sending_local_data":
		return Submessages.CLIENT_SENDING_LOCAL_DATA
	case "client_disconnecting":
		return Submessages.CLIENT_DISCONNECTING
	case "server_ping":
		return Submessages.SERVER_PING
	case "server_first_client_connection_information":
		return Submessages.SERVER_FIRST_CLIENT_CONNECTION_INFORMATION
	case "server_updating_players":
		return Submessages.SERVER_UPDATING_PLAYERS
	}
	return invalidSubmessage
}

func intToSubmessage(i int) Submessage {
	if i < 0 || i >= len(Submessages.All()) {
		return invalidSubmessage
	}
	return Submessages.All()[i]
}

func ExhaustiveSubmessages(f func(Submessage)) {
	for _, p := range Submessages.All() {
		f(p)
	}
}

var validSubmessages = map[Submessage]bool{
	Submessages.SUBMESSAGE_NONE:                            true,
	Submessages.CLIENT_SENDING_UDP_PORT:                    true,
	Submessages.CLIENT_READY:                               true,
	Submessages.CLIENT_SENDING_LOCAL_DATA:                  true,
	Submessages.CLIENT_DISCONNECTING:                       true,
	Submessages.SERVER_PING:                                true,
	Submessages.SERVER_FIRST_CLIENT_CONNECTION_INFORMATION: true,
	Submessages.SERVER_UPDATING_PLAYERS:                    true,
}

func (p Submessage) IsValid() bool {
	return validSubmessages[p]
}

func (p Submessage) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *Submessage) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, `"`), ` `)
	*p = ParseSubmessage(string(b))
	return nil
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the goenums command to generate them again.
	// Does not identify newly added constant values unless order changes
	var x [1]struct{}
	_ = x[submessage_none-0]
	_ = x[client_sending_udp_port-1]
	_ = x[client_ready-2]
	_ = x[client_sending_local_data-3]
	_ = x[client_disconnecting-4]
	_ = x[server_ping-5]
	_ = x[server_first_client_connection_information-6]
	_ = x[server_updating_players-7]
}

const _submessage_name = "submessage_noneclient_sending_udp_portclient_readyclient_sending_local_dataclient_disconnectingserver_pingserver_first_client_connection_informationserver_updating_players"

var _submessage_index = [...]uint16{0, 15, 38, 50, 75, 95, 106, 148, 171}

func (i submessage) String() string {
	if i < 0 || i >= submessage(len(_submessage_index)-1) {
		return "submessage(" + (strconv.FormatInt(int64(i), 10) + ")")
	}
	return _submessage_name[_submessage_index[i]:_submessage_index[i+1]]
}
