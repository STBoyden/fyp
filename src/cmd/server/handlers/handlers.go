package handlers

type Handler interface {
	Handle() error
}
