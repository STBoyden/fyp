package typedsockets

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
)

type UDPTypedConnection[T Convertable] struct {
	TypedConnection[T]
}

func NewUDPTypedConnection[T Convertable](conn net.Conn) UDPTypedConnection[T] {
	return UDPTypedConnection[T]{TypedConnection[T]{conn: conn, connectionType: ConnectionTypeUDP}}
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
		buffer := make([]byte, 4096)

		amountRead, addr, err := conn.ReadFrom(buffer)
		println(string(buffer))
		if err != nil {
			return amountRead, addr, errors.Join(errors.New("could not receive incoming buffer"), err)
		}
		if amountRead <= 0 {
			return 0, nil, errors.New("nothing read")
		}

		resizedBuffer := buffer[:amountRead]
		isNil, err := (*data).Unmarshal(resizedBuffer)
		if err != nil {
			return amountRead, addr, errors.Join(fmt.Errorf("could not unmarshal incoming buffer into %s: %s", reflect.TypeOf(data), err.Error()))
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
	connection       UDPTypedConnection[T]
	startedListening bool
}

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

func DialUDP[T Convertable](host, port string) (*UDPTypedConnection[T], error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}

	tc := NewUDPTypedConnection[T](conn)

	return &tc, nil
}
