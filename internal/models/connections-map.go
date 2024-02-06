package models

import (
	"fyp/common/state"
	typedsockets "fyp/common/utils/net/typed-sockets"
	"sync"
)

type typedConnections interface {
	typedsockets.TCPTypedConnection[state.State] | typedsockets.UDPTypedConnection[state.State]
}

type ConnectionsMap[T typedConnections] struct {
	connections      map[string]*T
	connectionsMutex sync.Mutex
}

func NewConnectionsMap[T typedConnections]() *ConnectionsMap[T] {
	return &ConnectionsMap[T]{
		connections:      map[string]*T{},
		connectionsMutex: sync.Mutex{},
	}
}

func (cm *ConnectionsMap[T]) UpdateConnection(source string, connection *T) {
	// already readied - not needed again
	if cm.connections[source] != nil {
		cm.connectionsMutex.Unlock()
		return
	}

	// connection is now ready
	cm.connections[source] = connection
	cm.connectionsMutex.Unlock()
}

func (cm *ConnectionsMap[T]) DeleteConnection(source string) {
	cm.connectionsMutex.Lock()
	delete(cm.connections, source)
	cm.connectionsMutex.Unlock()
}

// TODO: Implement iter for the map

// func IterConnectionsMap[T typedConnections]() chan struct {
// 	string
// 	typedConnections
// } {
// 	c := make(chan struct {
// 		string
// 		T
// 	})
// }
