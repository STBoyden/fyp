package typedsockets

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"reflect"
)

type TCPTypedConnection[T Convertable] struct {
	TypedConnection[T]
}

func newTCPTypedConnection[T Convertable](conn net.Conn) TCPTypedConnection[T] {
	return TCPTypedConnection[T]{TypedConnection[T]{conn: conn}}
}

func (utc *TCPTypedConnection[T]) ReadFrom(data *T) (int64, error) {
	switch conn := utc.conn.(type) {
	case *net.TCPConn:
		buffer := make([]byte, 0, 4096)
		reader := bytes.NewReader(buffer)

		amountRead, err := conn.ReadFrom(reader)
		if err != nil {
			return amountRead, errors.Join(errors.New("could not receive incoming buffer"), err)
		}

		isNil, err := (*data).Unmarshal(buffer)
		if err != nil {
			return amountRead, errors.Join(fmt.Errorf("could not unmarshal incoming buffer into %s", reflect.TypeOf(data)))
		}

		if isNil {
			return amountRead, fmt.Errorf("result of unmarshal was a nil value for type %s", reflect.TypeOf(data))
		}

		return amountRead, nil
	default:
		return 0, errors.New("conn is an invalid connection type for this method")
	}
}

type TCPSocketListener[T Convertable] struct {
	listener net.Listener
}

func NewTypedTCPSocketListenerFromPort[T Convertable](port string) (*TCPSocketListener[T], error) {
	address := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", address)

	if err != nil {
		return nil, err
	}

	return &TCPSocketListener[T]{listener: listener}, nil
}

func NewTypedTCPSocketListener[T Convertable](listener *net.TCPListener) *TCPSocketListener[T] {
	return &TCPSocketListener[T]{listener: listener}
}

func (this *TCPSocketListener[T]) Accept() (*TCPTypedConnection[T], error) {
	conn, err := this.listener.Accept()
	if err != nil {
		return nil, err
	}

	tc := newTCPTypedConnection[T](conn)

	return &tc, nil
}

func (this *TCPSocketListener[T]) Addr() net.Addr {
	return this.listener.Addr()
}

func (this *TCPSocketListener[T]) Close() error {
	return this.listener.Close()
}

func DialTCP[T Convertable](host string, port string) (*TCPTypedConnection[T], error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}

	tc := newTCPTypedConnection[T](conn)

	return &tc, nil
}
