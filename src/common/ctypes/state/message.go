package state

type message int

//revive:disable:var-naming
//go:generate goenums message.go
const (
	message_none message = iota
	hello
	from_client
	from_server
)
