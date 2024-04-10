package models

import (
	"sync"

	"fyp/src/common/ctypes/state"
)

type typedConnections interface {
	state.TCPConnection | state.UDPConnection
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

func (cm *ConnectionsMap[T]) GetConnection(source string) *T {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	return cm.connections[source]
}

type ConnectionsMapIterType[T typedConnections] struct {
	ID   string
	Conn T
}

/*
Iter allows for doing a `for-range` loop over the connections map inside the
ConnectionsMap struct. The loop holds a lock for the entirety of the contents
of the map, thus more entries will increase the amount of time this function
keeps the lock.
*/
func (cm *ConnectionsMap[T]) Iter() <-chan ConnectionsMapIterType[T] {
	iterChannel := make(chan ConnectionsMapIterType[T])

	go func() {
		cm.mutex.RLock()
		defer cm.mutex.RUnlock()

		for key, value := range cm.connections {
			value := *value
			iterChannel <- ConnectionsMapIterType[T]{ID: key, Conn: value}
		}

		close(iterChannel)
	}()

	return iterChannel
}
