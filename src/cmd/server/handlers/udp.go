package handlers

import (
	"net"
	"net/netip"
	"strings"
	"sync/atomic"

	"fyp/src/common/ctypes"
	"fyp/src/common/ctypes/state"
	"fyp/src/common/utils/logging"
	"fyp/src/internal/models"

	typedsockets "fyp/src/common/utils/net/typed-sockets"

	"github.com/google/uuid"
)

type GameHandler struct {
	logger          *logging.Logger
	serverState     *models.ServerState
	connectionsMap  *models.ConnectionsMap[state.UDPConnection]
	connectionSlots map[uuid.UUID]int
	connectedAmount int
	socket          state.UDPConnection
	connInfo        netip.AddrPort
	closeChannel    <-chan interface{}
	exitChannel     chan bool
	updateID        atomic.Uint64
	updates         map[uint64]state.State
}

var _ Handler = &GameHandler{}

func NewGameHandler(logger *logging.Logger, serverState *models.ServerState, socket *net.UDPConn, udpHost *net.UDPAddr, udpPort int, gracefulCloseChannel <-chan interface{}) *GameHandler {
	return &GameHandler{
		logger:          logger,
		serverState:     serverState,
		connectionsMap:  models.NewConnectionsMap[state.UDPConnection](),
		socket:          typedsockets.NewUDPTypedConnection[state.State](socket),
		connInfo:        netip.AddrPortFrom(udpHost.AddrPort().Addr(), uint16(udpPort)),
		closeChannel:    gracefulCloseChannel,
		exitChannel:     make(chan bool),
		connectionSlots: make(map[uuid.UUID]int),
		updates:         make(map[uint64]state.State),
	}
}

func (gh *GameHandler) handleDisconnection(id, name string) {
	if !gh.serverState.ContainsPlayer(name) {
		return
	}

	gh.connectionsMap.DeleteConnection(id)
	gh.serverState.RemovePlayer(name)

	players := gh.serverState.GetPlayers()

	for entry := range gh.connectionsMap.Iter() {
		_, err := entry.Conn.Write(state.WithUpdatedPlayers(int(gh.updateID.Load()), players))
		if err != nil {
			gh.logger.Errorf("[UDP: handleDisconnection] Could not update %s's version of currently connected players: %s", name, err.Error())
			continue
		}

		if len(gh.serverState.GetPlayers()) < 2 {
			_, err := entry.Conn.Write(state.WithServerMakingPlayerUnableToMove())
			if err != nil {
				gh.logger.Errorf("[UDP: handleDisconnection] Could not make player unmovable: %s", err.Error())
			}
		}
	}
}

func (gh *GameHandler) handleConnection(clientID string, clientState state.State) {
	if gh.serverState.ContainsPlayer(clientState.Client.Player.Name) {
		gh.serverState.UpdatePlayer(clientState.Client.Player.Name, clientState.Client.Player.Inner)
	} else {
		gh.serverState.AddPlayer(clientState.Client.Player.Name, clientState.Client.Player.Inner)
	}

	for entry := range gh.connectionsMap.Iter() {
		if clientID == entry.ID {
			continue
		}

		otherPlayers := gh.serverState.FilterPlayers(func(key string, _ ctypes.Player) bool {
			return key != entry.ID // we want only ids that aren't the current entry's ID
		})

		_, err := entry.Conn.Write(state.WithUpdatedPlayers(int(gh.updateID.Load()), otherPlayers))
		if err != nil {
			gh.logger.Errorf("[UDP: handleConnection] Could not send other player state: %s", err.Error())
		}

		if len(gh.serverState.GetPlayers()) >= 2 {
			_, err := entry.Conn.Write(state.WithServerMakingPlayerAbleToMove())
			if err != nil {
				gh.logger.Errorf("[UDP: handleConnection] Could not make player movable: %s", err.Error())
			}
		}
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
	connectedIDs := make(map[uuid.UUID]int)

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

		gh.updateID.Add(1)

		if clientState.Message == state.Messages.FROM_CLIENT {
			clientData := clientState.Client

			switch clientState.Submessage {
			case state.Submessages.CLIENT_SENDING_UDP_PORT:
				portColonIndex := strings.LastIndex(addr.String(), ":")
				clientIP = addr.String()[:portColonIndex]
				clientPort = clientData.UDPPort

				gh.logger.Debugf("[UDP] Initial connection with client at %s: %s", addr, clientState)
				id, err := uuid.NewRandom()
				if err != nil {
					gh.logger.Errorf("[UDP] Could not pre-generate UUID: %s", err)
				}
				connectedIDs[id] = gh.connectedAmount

				clientConn, err := typedsockets.DialUDP[state.State](clientIP, clientPort)
				if err != nil {
					gh.logger.Errorf("[UDP] Could not connect to client at '%s:%s'", clientIP, clientPort)
					return err
				}

				gh.connectionsMap.UpdateConnection(id.String(), clientConn)
				gh.logger.Infof("[UDP] Connected to client's UDP socket at %s:%s. Client ID: %s", clientIP, clientPort, id)

				_, err = clientConn.Write(state.WithNewClientConnection(id, connectedIDs[id]))
				if err != nil {
					gh.logger.Errorf("[UDP] Couldn't send to client: %s", err.Error())
					return err
				}
				gh.logger.Infof("[UDP] Sent initial data to client at %s:%s", clientIP, clientPort)

				continue

			case state.Submessages.CLIENT_READY:
				id := clientData.ID.UUID.String()

				if gh.connectionsMap.ContainsConnection(id) {
					if _, ok := gh.connectionSlots[clientData.ID.UUID]; !ok && gh.connectedAmount <= 4 {
						gh.connectionSlots[clientData.ID.UUID] = gh.connectedAmount
						connectedIDs[clientData.ID.UUID] = gh.connectedAmount

						gh.connectedAmount++
					}

					gh.handleConnection(id, clientState)
				} else {
					gh.logger.Errorf("[UDP] Client with id '%s' not found", clientState.Client.ID.UUID)
					continue
				}

			case state.Submessages.CLIENT_REQUESTING_UPDATE_ID:
				id := clientData.ID.UUID.String()
				requestedUpdateID := clientData.UpdateID

				if !gh.connectionsMap.ContainsConnection(id) {
					gh.logger.Errorf("[UDP] Client with id '%s' not found", id)
				}

				prevState, ok := gh.updates[requestedUpdateID]

				if !ok {
					gh.logger.Errorf("[UDP] Requested update with id '%d' not found", requestedUpdateID)
				}

				prevState.SetAsResending()

				conn := gh.connectionsMap.GetConnection(id)
				_, err := conn.Write(prevState)
				if err != nil {
					gh.logger.Errorf("[UDP] Could not resend update with id '%d': %s", requestedUpdateID, err)
				}
			case state.Submessages.CLIENT_DISCONNECTING:
				id := clientData.ID.UUID.String()
				name := clientData.Player.Name

				gh.connectedAmount--
				delete(gh.connectionSlots, clientData.ID.UUID)

				gh.handleDisconnection(id, name)

				gh.logger.Infof("[UDP] Disconnected from client with id: %s", id)
			}
		}

		gh.updates[gh.updateID.Load()] = gh.serverState.Copy()
	}

	gh.logger.Warn("[UDP] Closed")
	return nil
}
