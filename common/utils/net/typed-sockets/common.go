package typedsockets

import (
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"time"
)

type Convertable interface {
	// Takes the current state of the implementing struct, and marshals it into a
	// format chosen by the implementor.
	Marshal() (data []byte, err error)

	// Unmarshals the given data passed as parameter into the implementor's type.
	// The bool return is used to indicate whether the pointer to the implementor
	// type is nil (false) or not nil (true).
	Unmarshal(data []byte) (isNil bool, err error)
}

type connectionType int

const (
	CONNECTION_TYPE_TCP connectionType = iota
	CONNECTION_TYPE_UDP
)

func (ct connectionType) String() string {
	switch ct {
	case CONNECTION_TYPE_TCP:
		return "tcp"
	case CONNECTION_TYPE_UDP:
		return "udp"
	default:
		return "unknown"
	}
}

// Ensure that ConnectionType implements Stringer correctly
var _ fmt.Stringer = CONNECTION_TYPE_TCP

type TypedConnection[T Convertable] struct {
	conn           net.Conn
	connectionType connectionType
}

func NewTypedConnection[T Convertable](conn net.Conn, connectionType connectionType) TypedConnection[T] {
	return TypedConnection[T]{conn: conn, connectionType: connectionType}
}

func (tc *TypedConnection[T]) ConnectionType() connectionType {
	return tc.connectionType
}

// Reads from the connection and attempts to read an entire buffer into `T`.
func (tc *TypedConnection[T]) Read(data *T) (bytesRead int, err error) {
	if data == nil {
		return 0, errors.New("data pointer was nil")
	}

	buffer := make([]byte, 0, 4096)
	chunk := make([]byte, 256)

	for {
		amount, err := tc.conn.Read(chunk)
		if err != nil {
			if err != io.EOF {
				return amount, err
			}

			break
		}

		buffer = append(buffer, chunk[:amount]...)
	}

	isNil, err := (*data).Unmarshal(buffer)
	if err != nil {
		return 0, errors.Join(errors.New("unmarshal of data returned an error"), err)
	}

	if isNil {
		return 0, fmt.Errorf("unmarshal of data returned a nil pointer of %s", reflect.TypeOf(data))
	}

	return len(buffer), nil
}

func (tc *TypedConnection[T]) Write(data T) (bytesWritten int, err error) {
	buffer, err := data.Marshal()
	if err != nil {
		return 0, errors.Join(errors.New("could not marshal data to write"), err)
	}

	return tc.conn.Write(buffer)
}

func (tc *TypedConnection[T]) Close() error {
	return tc.conn.Close()
}

func (tc *TypedConnection[T]) LocalAddr() net.Addr {
	return tc.conn.LocalAddr()
}

func (tc *TypedConnection[T]) RemoteAddr() net.Addr {
	return tc.conn.RemoteAddr()
}

func (tc *TypedConnection[T]) SetDeadline(t time.Time) error {
	return tc.conn.SetDeadline(t)
}

func (tc *TypedConnection[T]) SetReadDeadline(t time.Time) error {
	return tc.conn.SetReadDeadline(t)
}

func (tc *TypedConnection[T]) SetWriteDeadline(t time.Time) error {
	return tc.conn.SetWriteDeadline(t)
}
