package typedsockets

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type Convertable interface {
	fmt.Stringer

	// Takes the current state of the implementing struct, and marshals it into a
	// format chosen by the implementor.
	Marshal() (data []byte, err error)

	// Unmarshals the given data passed as parameter into the implementor's type.
	Unmarshal(v any, data []byte) error
}

type ConnectionType int

const (
	ConnectionTypeTCP ConnectionType = iota
	ConnectionTypeUDP
)

func (ct ConnectionType) String() string {
	switch ct {
	case ConnectionTypeTCP:
		return "tcp"
	case ConnectionTypeUDP:
		return "udp"
	default:
		return "unknown"
	}
}

// Ensure that ConnectionType implements Stringer correctly.
var _ fmt.Stringer = ConnectionTypeTCP

type TypedConnection[T Convertable] struct {
	conn           net.Conn
	connectionType ConnectionType
}

func NewTypedConnection[T Convertable](conn net.Conn, connectionType ConnectionType) TypedConnection[T] {
	return TypedConnection[T]{conn: conn, connectionType: connectionType}
}

func (tc *TypedConnection[T]) ConnectionType() ConnectionType {
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
			if errors.Is(err, io.EOF) {
				return amount, err
			}

			break
		}

		buffer = append(buffer, chunk[:amount]...)
	}

	var newData T
	err = newData.Unmarshal(&newData, buffer)
	if err != nil {
		return 0, errors.Join(errors.New("unmarshal of data returned an error"), err)
	}

	*data = newData

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
