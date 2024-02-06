package typedsockets

import (
	"errors"
	"fmt"
	"net"
	"reflect"
)

type UDPTypedConnection[T Convertable] struct {
	TypedConnection[T]
}

func NewUDPTypedConnection[T Convertable](conn net.Conn) UDPTypedConnection[T] {
	return UDPTypedConnection[T]{TypedConnection[T]{conn: conn, connectionType: CONNECTION_TYPE_UDP}}
}

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

func (utc *UDPTypedConnection[T]) ReadFrom(data *T) (int, net.Addr, error) {
	switch conn := utc.conn.(type) {
	case *net.UDPConn:
		buffer := make([]byte, 0, 4096)
		amountRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			return amountRead, addr, errors.Join(errors.New("could not receive incoming buffer"), err)
		}

		isNil, err := (*data).Unmarshal(buffer)
		if err != nil {
			return amountRead, addr, errors.Join(fmt.Errorf("could not unmarshal incoming buffer into %s", reflect.TypeOf(data)))
		}

		if isNil {
			return amountRead, addr, fmt.Errorf("result of unmarshal was a nil value for type %s", reflect.TypeOf(data))
		}

		return amountRead, addr, nil
	default:
		return 0, nil, errors.New("conn is an invalid connection type for this method")
	}
}

type UDPSocketListener[T Convertable] struct {
	listener net.Listener
}

func NewTypedUDPSocketListener[T Convertable](port string) (*UDPSocketListener[T], error) {
	address := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &UDPSocketListener[T]{listener: listener}, nil
}

func (this *UDPSocketListener[T]) Accept() (*UDPTypedConnection[T], error) {
	conn, err := this.listener.Accept()
	println("got here ACCEPT")
	if err != nil {
		return nil, err
	}

	tc := NewUDPTypedConnection[T](conn)

	return &tc, nil
}

func (this *UDPSocketListener[T]) Addr() net.Addr {
	return this.listener.Addr()
}

func (this *UDPSocketListener[T]) Close() error {
	return this.listener.Close()
}

func DialUDP[T Convertable](host string, port string) (*UDPTypedConnection[T], error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}

	tc := NewUDPTypedConnection[T](conn)

	return &tc, nil
}
