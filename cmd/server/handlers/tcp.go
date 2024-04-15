package handlers

import (
	"net"
	"strings"

	"fyp/common/ctypes/state"
	"fyp/common/utils/logging"
	"fyp/internal/models"

	typedsockets "fyp/common/utils/net/typed-sockets"
)

type TCPHandler struct {
	Handler

	serverState    *models.ServerState
	logger         *logging.Logger
	connectionsMap *models.ConnectionsMap[state.TCPConnection]
	socket         *state.TCPSocketListener
	port           int
	closeChannel   <-chan any
}

func NewTCPHandler(logger *logging.Logger, serverState *models.ServerState, socket *net.TCPListener, tcpPort int, gracefulCloseChannel <-chan any) *TCPHandler {
	return &TCPHandler{
		logger:         logger,
		serverState:    serverState,
		connectionsMap: models.NewConnectionsMap[state.TCPConnection](),
		socket:         typedsockets.NewTypedTCPSocketListener[state.State](socket),
		port:           tcpPort,
		closeChannel:   gracefulCloseChannel,
	}
}

func (th *TCPHandler) Handle() error {
	th.logger.Infof("Started error correction socket (TCP) on %d\n", th.port)
	exitChan := make(chan bool)

	// Async closure to handle the closing of the socket, waits for the gracefulCloseChannel, and exits when it receives anything
	go func() {
		<-th.closeChannel
		th.logger.Infof("[TCP] Stopping...")
		th.socket.Close()
	}()

	go func() {
		for {
			conn, err := th.socket.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					th.logger.Warn("[TCP] Closed")
					exitChan <- true
					break
				}

				th.logger.Errorf("[TCP] Could not receive on TCP socket: %s", err.Error())
				continue
			}

			th.logger.Infof("[TCP] Connected with %s", conn.RemoteAddr())
			th.connectionsMap.UpdateConnection(conn.RemoteAddr().String(), conn)
		}
	}()

	for !<-exitChan {
		for pair := range th.connectionsMap.Iter() {
			conn := pair.Conn
			id := pair.ID

			if _, err := conn.Write(state.WithServerPing()); err == nil {
				continue
			}

			th.connectionsMap.DeleteConnection(id)
			th.logger.Infof("[TCP] Disconnected from %s", id)

			// TODO Handle sending error corrections to client here
		}
	}

	return nil
}
