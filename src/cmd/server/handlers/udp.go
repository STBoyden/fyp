package handlers

import (
	"net"
	"net/netip"
	"strings"

	"fyp/src/common/state"
	"fyp/src/common/utils/logging"
	"fyp/src/internal/models"

	typedsockets "fyp/src/common/utils/net/typed-sockets"
)

type GameHandler struct {
	logger         *logging.Logger
	connectionsMap *models.ConnectionsMap[typedsockets.UDPTypedConnection[state.State]]
	socket         typedsockets.UDPTypedConnection[state.State]
	connInfo       netip.AddrPort
	closeChannel   chan interface{}
}

var _ Handler = GameHandler{}

func NewGameHandler(logger *logging.Logger, socket *net.UDPConn, udpHost *net.UDPAddr, udpPort int, gracefulCloseChannel chan interface{}) *GameHandler {
	return &GameHandler{
		logger:         logger,
		connectionsMap: models.NewConnectionsMap[typedsockets.UDPTypedConnection[state.State]](),
		socket:         typedsockets.NewUDPTypedConnection[state.State](socket),
		connInfo:       netip.AddrPortFrom(udpHost.AddrPort().Addr(), uint16(udpPort)),
		closeChannel:   gracefulCloseChannel,
	}
}

func (gh GameHandler) Handle() error {
	gh.logger.Infof("Started game data socket (UDP) on %d\n", gh.connInfo.Port())

	// Async closure to handle the closing of the socket, waits for the gracefulCloseChannel, and exits when it receives anything
	go func() {
		<-gh.closeChannel
		gh.logger.Infof("[UDP] Stopping...")
		gh.socket.Close()
	}()

	clientState := state.Empty()
	var clientIP string
	var clientPort string
	oldAddrs := make(map[string]struct{}) // hack to get a set of `net.Addr`s

	for {
		size, addr, err := gh.socket.ReadFrom(&clientState)

		if addr == nil {
			continue
		}

		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				gh.logger.Warn("[UDP] Closed")
				break
			} else if strings.Contains(err.Error(), "nothing read") {
				continue
			}

			gh.logger.Errorf("[UDP] %s\n", err)
			continue
		}

		// check if the address is in the set of the previously connected addresses, and log if it isn't
		if _, ok := oldAddrs[addr.String()]; !ok {
			gh.logger.Debugf("[UDP] Initial connection with %s", addr)
			oldAddrs[addr.String()] = struct{}{}
		}

		if size == 0 || clientState == state.Empty() {
			continue
		}

		gh.logger.Debugf("[UDP] Received state from client at %s: %s", addr, clientState)

		if clientState.ClientUDPPort != "" {
			portColonIndex := strings.LastIndex(addr.String(), ":")
			clientIP = addr.String()[:portColonIndex]
			clientPort = clientState.ClientUDPPort
		}

		if clientState.ClientReady {
			clientConn, err := typedsockets.DialUDP[state.State](clientIP, clientPort)
			if err != nil {
				gh.logger.Errorf("[UDP] Could not connect to client at '%s:%s'", clientIP, clientPort)
				return err
			}

			source := clientConn.RemoteAddr().String()
			gh.connectionsMap.UpdateConnection(source, clientConn)

			gh.logger.Infof("[UDP] Connected to client's UDP socket at %s", source)

			// TODO Logic with sending the game updates and logic here.
			gh.logger.Infof("[UDP] Sending 'hello' message to client at %s", source)
			_, err = clientConn.Write(state.State{ServerMessage: "Hello, world!"})
			if err != nil {
				gh.logger.Errorf("[UDP] Couldn't send to client: %s", err.Error())
				return err
			}

			return nil
		}
	}

	return nil
}
