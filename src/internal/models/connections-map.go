package models

import (
	"sync"

	"fyp/src/common/state"

	typedsockets "fyp/src/common/utils/net/typed-sockets"
)

type typedConnections interface {
	typedsockets.TCPTypedConnection[state.State] | typedsockets.UDPTypedConnection[state.State]
}

type ConnectionsMap[T typedConnections] struct {
	connections map[string]*T
	mutex       sync.RWMutex
}

func NewConnectionsMap[T typedConnections]() *ConnectionsMap[T] {
	return &ConnectionsMap[T]{
		connections: map[string]*T{},
		mutex:       sync.RWMutex{},
	}
}

func (cm *ConnectionsMap[T]) UpdateConnection(source string, connection *T) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// already readied - not needed again
	if cm.connections[source] != nil {
		return
	}

	// connection is now ready
	cm.connections[source] = connection
}

func (cm *ConnectionsMap[T]) DeleteConnection(source string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.connections, source)
}

func (cm *ConnectionsMap[T]) ContainsConnection(source string) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.connections[source] != nil
}

type ConnectionsMapIterType[T typedConnections] struct {
	ID   string
	Conn T
}

func (cm *ConnectionsMap[T]) Iter() chan ConnectionsMapIterType[T] {
	iterChannel := make(chan ConnectionsMapIterType[T])

	go func() {
		cm.mutex.Lock()
		defer cm.mutex.Unlock()

		for key, value := range cm.connections {
			value := *value
			iterChannel <- ConnectionsMapIterType[T]{ID: key, Conn: value}
		}

		close(iterChannel)
	}()

	return iterChannel
}
