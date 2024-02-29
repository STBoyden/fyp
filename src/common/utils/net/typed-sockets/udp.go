package typedsockets

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
)

/*
UDPTypedConnection is a TypedConnection that is suited for UDP connections, and provides
UDP-specific function implementations.
*/
type UDPTypedConnection[T Convertable] struct {
	TypedConnection[T]
}

/*
NewUDPTypedConnection creates a new UDPTypedConnections specialised for T.
*/
func NewUDPTypedConnection[T Convertable](conn net.Conn) UDPTypedConnection[T] {
	return UDPTypedConnection[T]{TypedConnection[T]{conn: conn, connectionType: ConnectionTypeUDP}}
}

/*
WriteTo writes to the inner connection. It attempts to write the given data T. On
success, it will return the amount of bytes written. On failure, it will return an error.
*/
func (utc *UDPTypedConnection[T]) WriteTo(data T, addr net.Addr) (int, error) {
	switch conn := utc.conn.(type) {
	case *net.UDPConn:
		buffer, err := data.Marshal()
		if err != nil {
			return 0, err
		}

		return conn.WriteTo(buffer, addr)
	default:
		return 0, errors.New("conn is an invalid connection type for this method")
	}
}

/*
ReadFrom reads from the inner connection, attempting to read a T from the connection. On
success, the amount of bytes read is returned and the data parameter is populated with
the read data from the connection. On failure, the amount of bytes read is still returned
but so is an error. The data parameter is left untouched.
*/
func (utc *UDPTypedConnection[T]) ReadFrom(data *T) (int, net.Addr, error) {
	switch conn := utc.conn.(type) {
	case *net.UDPConn:
		buffer := make([]byte, 4096)

		amountRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			return amountRead, addr, errors.Join(errors.New("could not receive incoming buffer"), err)
		}
		if amountRead <= 0 {
			return 0, nil, errors.New("nothing read")
		}

		resizedBuffer := buffer[:amountRead]

		var newData T
		err = newData.Unmarshal(&newData, resizedBuffer)
		if err != nil {
			return amountRead, addr, errors.Join(fmt.Errorf("could not unmarshal incoming buffer into %s: %s", reflect.TypeOf(data), err.Error()))
		}

		*data = newData

		return amountRead, addr, nil
	default:
		return 0, nil, errors.New("conn is an invalid connection type for this method")
	}
}

/*
DialUDP attempts to connect to a given UDP socket at host:port and creates a new
UDPTypedConnection[T] on success. On failure, an error is returned.
*/
func DialUDP[T Convertable](host, port string) (*UDPTypedConnection[T], error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}

	tc := NewUDPTypedConnection[T](conn)

	return &tc, nil
}

/*
UDPSocketListener describes a type-safe UDP socket listener.
*/
type UDPSocketListener[T Convertable] struct {
	connection       UDPTypedConnection[T]
	startedListening bool
}

/*
NewTypedUDPSocketListener creates a new *UDPSocketListener when given only aport. This
function assumes that the host will be local, and assigns to 0.0.0.0. On success, the new
listener is returned. On failure, an error is returned.
*/
func NewTypedUDPSocketListener[T Convertable](port string) (*UDPSocketListener[T], error) {
	iport, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	address := net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: iport}
	conn, err := net.ListenUDP("udp", &address)
	if err != nil {
		return nil, err
	}

	return &UDPSocketListener[T]{
			connection:       NewUDPTypedConnection[T](conn),
			startedListening: true,
		},
		nil
}

/*
Conn returns the inner type-safe connection of the listener.
*/
func (usl *UDPSocketListener[T]) Conn() (*UDPTypedConnection[T], error) {
	if !usl.startedListening {
		return nil, errors.New("this socket hasn't started listening")
	}

	return &usl.connection, nil
}

// func (this *UDPSocketListener[T]) Accept() (*UDPTypedConnection[T], error) {
// 	if !this.startedListening {
// 		return nil, errors.New("this socket isn't listening anywhere")
// 	}

// 	conn, err := this.listener.Accept()
// 	println("got here ACCEPT")
// 	if err != nil {
// 		return nil, err
// 	}

// 	tc := NewUDPTypedConnection[T](conn)

// 	return &tc, nil
// }

// func (this *UDPSocketListener[T]) Addr() net.Addr {
// 	return this.listener.Addr()
// }

// func (this *UDPSocketListener[T]) Close() error {
// 	return this.listener.Close()
// }
