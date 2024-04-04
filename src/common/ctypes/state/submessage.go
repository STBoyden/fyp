package state

type submessage int

//revive:disable:var-naming
//go:generate goenums submessage.go
const (
	submessage_none submessage = iota
	client_sending_udp_port
	client_ready
	client_sending_local_data
	client_disconnecting
	client_has_finished_level

	server_ping
	server_first_client_connection_information
	server_updating_players
	server_this_client_can_move
	server_this_client_cannot_move
	server_players_have_finished
)
