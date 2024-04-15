package typedsockets

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
)

/*
TCPTypedConnection is a TypedConnection that is suited for TCP connections, and provides
TCP-specific function implementations.
*/
type TCPTypedConnection[T Convertable] struct {
	TypedConnection[T]
}

/*
NewTCPTypedConnection creates a new TCPTypedConnections specialised for T.
*/
func NewTCPTypedConnection[T Convertable](conn net.Conn) TCPTypedConnection[T] {
	return TCPTypedConnection[T]{TypedConnection[T]{conn: conn, connectionType: ConnectionTypeTCP}}
}

/*
ReadFrom reads from the inner connection, attempting to read a T from the connection. On
success, the amount of bytes read is returned and the data parameter is populated with
the read data from the connection. On failure, the amount of bytes read is still returned
but so is an error. The data parameter is left untouched.
*/
func (utc *TCPTypedConnection[T]) ReadFrom(data *T) (int64, error) {
	switch conn := utc.conn.(type) {
	case *net.TCPConn:
		buffer := make([]byte, 4096)
		reader := bytes.NewReader(buffer)

		amountRead, err := conn.ReadFrom(reader)
		if err != nil {
			return amountRead, errors.Join(errors.New("could not receive incoming buffer"), err)
		}

		resizedBuffer := buffer[:amountRead]

		var newData T
		err = newData.Unmarshal(&newData, resizedBuffer)
		if err != nil {
			return amountRead, errors.Join(fmt.Errorf("could not unmarshal incoming buffer into %s", reflect.TypeOf(data)))
		}

		*data = newData

		return amountRead, nil
	default:
		return 0, errors.New("conn is an invalid connection type for this method")
	}
}

/*
DialTCP attempts to connect to a given TCP socket at host:port and creates a new
TCPTypedConnection[T] on success. On failure, an error is returned.
*/
func DialTCP[T Convertable](host, port string) (*TCPTypedConnection[T], error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}

	tc := NewTCPTypedConnection[T](conn)

	return &tc, nil
}

/*
TCPSocketListener is a type-safe wrapper over *net.TCPListener.
*/
type TCPSocketListener[T Convertable] struct {
	listener *net.TCPListener
}

/*
NewTypedTCPSocketListenerFromPort creates a new *TCPSocketListener when given only a
port. This function assumes that the host will be local, and assigns to 0.0.0.0. On success,
the new listener is returned. On failure, an error is returned.
*/
func NewTypedTCPSocketListenerFromPort[T Convertable](port string) (*TCPSocketListener[T], error) {
	iport, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	address := net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: iport}
	listener, err := net.ListenTCP("tcp", &address)
	if err != nil {
		return nil, err
	}

	return &TCPSocketListener[T]{listener: listener}, nil
}

/*
NewTypedTCPSocketListener creates a *TCPSocketListener from a pre-existing
*net.TCPListener.
*/
func NewTypedTCPSocketListener[T Convertable](listener *net.TCPListener) *TCPSocketListener[T] {
	return &TCPSocketListener[T]{listener: listener}
}

/*
Accept starts listening on the inner TCPListner, and creates a *TCPTypedConnection from
the listener. On success, the new *TCPTypedConnection is returned. On failutre, an error
is returned.
*/
func (tsl *TCPSocketListener[T]) Accept() (*TCPTypedConnection[T], error) {
	conn, err := tsl.listener.Accept()
	if err != nil {
		return nil, err
	}

	tc := NewTCPTypedConnection[T](conn)

	return &tc, nil
}

func (tsl *TCPSocketListener[T]) Addr() net.Addr {
	return tsl.listener.Addr()
}

func (tsl *TCPSocketListener[T]) Close() error {
	return tsl.listener.Close()
}
