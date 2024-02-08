package handlers

import (
	"net"
	"strings"

	"fyp/src/common/state"
	"fyp/src/common/utils/logging"
	typedsockets "fyp/src/common/utils/net/typed-sockets"
	"fyp/src/internal/models"
)

type ErrorCorrectionHandler struct {
	Handler

	logger         *logging.Logger
	connectionsMap *models.ConnectionsMap[typedsockets.TCPTypedConnection[state.State]]
	socket         *typedsockets.TCPSocketListener[state.State]
	port           int
	closeChannel   chan struct{}
}

func NewErrorCorrectionHandler(logger *logging.Logger, socket *net.TCPListener, tcpPort int, gracefulCloseChannel chan struct{}) *ErrorCorrectionHandler {
	return &ErrorCorrectionHandler{
		logger:         logger,
		connectionsMap: models.NewConnectionsMap[typedsockets.TCPTypedConnection[state.State]](),
		socket:         typedsockets.NewTypedTCPSocketListener[state.State](socket),
		port:           tcpPort,
		closeChannel:   gracefulCloseChannel,
	}
}

func (ec ErrorCorrectionHandler) Handle() error {
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
			ec.connectionsMap.UpdateConnection(conn.RemoteAddr().String(), conn)
		}
	}()

	for {
		if exit {
			break
		}

		for pair := range ec.connectionsMap.Iter() {
			conn := pair.Conn
			id := pair.ID

			if _, err := conn.Write(state.State{ServerPing: state.ServerPing}); err == nil {
				continue
			}

			ec.connectionsMap.DeleteConnection(id)
			ec.logger.Infof("[TCP] Disconnected from %s", id)

			// TODO Handle sending error corrections to client here
		}
	}

	return nil
}
