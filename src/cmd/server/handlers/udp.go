package handlers

import (
	"net"
	"net/netip"
	"strings"

	"fyp/src/common/ctypes/state"
	"fyp/src/common/utils/logging"
	"fyp/src/internal/models"

	typedsockets "fyp/src/common/utils/net/typed-sockets"

	"github.com/google/uuid"
)

type GameHandler struct {
	logger          *logging.Logger
	serverState     *models.ServerState
	connectionsMap  *models.ConnectionsMap[typedsockets.UDPTypedConnection[state.State]]
	connectionSlots map[uuid.UUID]int
	connectedAmount int
	socket          typedsockets.UDPTypedConnection[state.State]
	connInfo        netip.AddrPort
	closeChannel    <-chan interface{}
	exitChannel     chan bool
}

var _ Handler = &GameHandler{}

func NewGameHandler(logger *logging.Logger, serverState *models.ServerState, socket *net.UDPConn, udpHost *net.UDPAddr, udpPort int, gracefulCloseChannel <-chan interface{}) *GameHandler {
	return &GameHandler{
		logger:          logger,
		serverState:     serverState,
		connectionsMap:  models.NewConnectionsMap[typedsockets.UDPTypedConnection[state.State]](),
		socket:          typedsockets.NewUDPTypedConnection[state.State](socket),
		connInfo:        netip.AddrPortFrom(udpHost.AddrPort().Addr(), uint16(udpPort)),
		closeChannel:    gracefulCloseChannel,
		exitChannel:     make(chan bool),
		connectionSlots: make(map[uuid.UUID]int),
	}
}

func (gh *GameHandler) Handle() error {
	gh.logger.Infof("Started game data socket (UDP) on %d\n", gh.connInfo.Port())

	// Async closure to handle the closing of the socket, waits for the gracefulCloseChannel, and exits when it receives anything
	go func() {
		<-gh.closeChannel
		gh.logger.Infof("[UDP] Stopping...")
		gh.socket.Close()
		gh.exitChannel <- true
	}()

	clientState := state.Empty()
	var clientIP string
	var clientPort string
	oldIDs := make(map[uuid.UUID]struct{})

outer:
	for {
		select {
		case doClose := <-gh.exitChannel:
			if doClose {
				break outer
			}
		default:
		}

		size, addr, err := gh.socket.ReadFrom(&clientState)

		if addr == nil {
			continue
		}

		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			} else if strings.Contains(err.Error(), "nothing read") {
				continue
			}

			gh.logger.Errorf("[UDP] %s\n", err)
			continue
		}

		if size == 0 {
			continue
		}

		if clientState.ClientUDPPort != "" {
			portColonIndex := strings.LastIndex(addr.String(), ":")
			clientIP = addr.String()[:portColonIndex]
			clientPort = clientState.ClientUDPPort
			continue
		}

		var id uuid.UUID
		var slot int

		if _, ok := oldIDs[clientState.ClientID.UUID]; !ok {
			gh.logger.Debugf("[UDP] Initial connection with client at %s: %s", addr, clientState)

			id, err = uuid.NewRandom()
			if err != nil {
				gh.logger.Errorf("[UDP] Could not pre-generate UUID: %s", err)
			}
			oldIDs[id] = struct{}{}
		} else {
			id = clientState.ClientID.UUID
		}

		if _, ok := gh.connectionSlots[id]; !ok && gh.connectedAmount <= 4 {
			gh.connectionSlots[id] = gh.connectedAmount
			slot = gh.connectedAmount

			gh.connectedAmount++
		}

		// gh.logger.Debugf("[UDP] Received state from client at %s: %s", addr, clientState)

		if gh.connectionsMap.ContainsConnection(id.String()) {
			// gh.logger.Debugf("[UDP] Got player data from %s (id: %s)", addr.String(), id.String())
			for name, player := range clientState.Players {
				gh.serverState.AddPlayer(name, player)
			}
		}

		if clientState.ClientReady {
			clientConn, err := typedsockets.DialUDP[state.State](clientIP, clientPort)
			if err != nil {
				gh.logger.Errorf("[UDP] Could not connect to client at '%s:%s'", clientIP, clientPort)
				return err
			}

			gh.connectionsMap.UpdateConnection(id.String(), clientConn)

			gh.logger.Infof("[UDP] Connected to client's UDP socket at %s:%s", clientIP, clientPort)

			// TODO Logic with sending the game updates and logic here.
			gh.logger.Infof("[UDP] Sending 'hello' message and new ID to client at %s:%s", clientIP, clientPort)
			_, err = clientConn.Write(state.State{ServerMessage: "Hello, world!", ClientSlot: slot, ClientID: uuid.NullUUID{UUID: id, Valid: true}})
			if err != nil {
				gh.logger.Errorf("[UDP] Couldn't send to client: %s", err.Error())
				return err
			}

			// return nil
		}

		if clientState.ClientDisconnecting {
			gh.connectedAmount--
			delete(gh.connectionSlots, id)
			gh.connectionsMap.DeleteConnection(id.String())

			gh.logger.Infof("[UDP] Disconnected from client with id: %s", id.String())
		}
	}

	gh.logger.Warn("[UDP] Closed")
	return nil
}
