package handlers

import (
	"fyp/common/utils/logging"
	"net"
	"strings"
	"sync"
)

type ErrorCorrectionHandler struct {
	Handler

	logger           *logging.Logger
	connections      map[string]net.Conn
	connectionsMutex sync.Mutex
	socket           *net.TCPListener
	port             int
	closeChannel     chan struct{}
}

func NewErrorCorrectionHandler(logger *logging.Logger, socket *net.TCPListener, tcpPort int, gracefulCloseChannel chan struct{}) *ErrorCorrectionHandler {
	return &ErrorCorrectionHandler{
		logger:           logger,
		connections:      map[string]net.Conn{},
		connectionsMutex: sync.Mutex{},
		socket:           socket,
		port:             tcpPort,
		closeChannel:     gracefulCloseChannel,
	}
}

func (ec *ErrorCorrectionHandler) Handle() {
	ec.logger.Infof("Started error correction socket (TCP) on %d\n", ec.port)
	exit := false

	// Async closure to handle the closing of the socket, waits for the gracefulCloseChannel, and exits when it receives anything
	go func() {
		<-ec.closeChannel
		ec.logger.Infof("[TCP] Stopping...")
		ec.socket.Close()
	}()

	go func() {
		for {
			conn, err := ec.socket.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					ec.logger.Warn("[TCP] Closed")
					exit = true
					break
				}

				ec.logger.Errorf("[TCP] Could not receive on TCP socket: %s", err.Error())
				continue
			}

			ec.logger.Infof("[TCP] Connected with %s", conn.RemoteAddr())
			ec.connectionsMutex.Lock()
			ec.connections[conn.RemoteAddr().String()] = conn
			ec.connectionsMutex.Unlock()
		}
	}()

	for {
		if exit {
			break
		}

		for id, conn := range ec.connections {
			if _, err := conn.Write([]byte("ping")); err != nil {
				ec.connectionsMutex.Lock()
				delete(ec.connections, id)
				ec.connectionsMutex.Unlock()
				ec.logger.Infof("[TCP] Disconnected from %s", id)
				continue
			}

			// TODO Handle sending error corrections to client here
		}
	}
}
