package handlers

import (
	"net"
	"strings"

	"fyp/src/common/ctypes/state"
	"fyp/src/common/utils/logging"
	"fyp/src/internal/models"

	typedsockets "fyp/src/common/utils/net/typed-sockets"
)

type ErrorCorrectionHandler struct {
	Handler

	serverState    *models.ServerState
	logger         *logging.Logger
	connectionsMap *models.ConnectionsMap[state.TCPConnection]
	socket         *state.TCPSocketListener
	port           int
	closeChannel   <-chan interface{}
}

func NewErrorCorrectionHandler(logger *logging.Logger, serverState *models.ServerState, socket *net.TCPListener, tcpPort int, gracefulCloseChannel <-chan interface{}) *ErrorCorrectionHandler {
	return &ErrorCorrectionHandler{
		logger:         logger,
		serverState:    serverState,
		connectionsMap: models.NewConnectionsMap[state.TCPConnection](),
		socket:         typedsockets.NewTypedTCPSocketListener[state.State](socket),
		port:           tcpPort,
		closeChannel:   gracefulCloseChannel,
	}
}

func (ec *ErrorCorrectionHandler) Handle() error {
	ec.logger.Infof("Started error correction socket (TCP) on %d\n", ec.port)
	exitChan := make(chan bool)

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
					exitChan <- true
					break
				}

				ec.logger.Errorf("[TCP] Could not receive on TCP socket: %s", err.Error())
				continue
			}

			ec.logger.Infof("[TCP] Connected with %s", conn.RemoteAddr())
			ec.connectionsMap.UpdateConnection(conn.RemoteAddr().String(), conn)
		}
	}()

	for !<-exitChan {
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
