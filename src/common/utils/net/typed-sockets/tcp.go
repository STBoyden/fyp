package typedsockets

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
)

type TCPTypedConnection[T Convertable] struct {
	TypedConnection[T]
}

func NewTCPTypedConnection[T Convertable](conn net.Conn) TCPTypedConnection[T] {
	return TCPTypedConnection[T]{TypedConnection[T]{conn: conn, connectionType: ConnectionTypeTCP}}
}

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

type TCPSocketListener[T Convertable] struct {
	listener *net.TCPListener
}

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

func NewTypedTCPSocketListener[T Convertable](listener *net.TCPListener) *TCPSocketListener[T] {
	return &TCPSocketListener[T]{listener: listener}
}

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

func DialTCP[T Convertable](host, port string) (*TCPTypedConnection[T], error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}

	tc := NewTCPTypedConnection[T](conn)

	return &tc, nil
}
