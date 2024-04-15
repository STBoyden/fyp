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

type UDPHandler struct {
	logger          *logging.Logger
	serverState     *models.ServerState
	connectionsMap  *models.ConnectionsMap[state.UDPConnection]
	connectionSlots map[uuid.UUID]int
	connectedAmount int
	socket          state.UDPConnection
	connInfo        netip.AddrPort
	closeChannel    <-chan any
	exitChannel     chan bool
	updateID        atomic.Uint64
	updates         map[uint64]state.State
}

var _ Handler = &UDPHandler{}

func NewUDPHandler(logger *logging.Logger, serverState *models.ServerState, socket *net.UDPConn, udpHost *net.UDPAddr, udpPort int, gracefulCloseChannel <-chan any) *UDPHandler {
	return &UDPHandler{
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

func (uh *UDPHandler) handleDisconnection(id, name string) {
	if !uh.serverState.ContainsPlayer(name) {
		return
	}

	uh.connectionsMap.DeleteConnection(id)
	uh.serverState.RemovePlayer(name)

	players := uh.serverState.GetPlayers()

	for entry := range uh.connectionsMap.Iter() {
		_, err := entry.Conn.Write(state.WithUpdatedPlayers(int(uh.updateID.Load()), players))
		if err != nil {
			uh.logger.Errorf("[UDP: handleDisconnection] Could not update %s's version of currently connected players: %s", name, err.Error())
			continue
		}

		if len(uh.serverState.GetPlayers()) < 2 {
			_, err := entry.Conn.Write(state.WithServerMakingPlayerUnableToMove())
			if err != nil {
				uh.logger.Errorf("[UDP: handleDisconnection] Could not make player unmovable: %s", err.Error())
			}
		}
	}
}

func (uh *UDPHandler) handleConnection(clientID string, clientState state.State) {
	if uh.serverState.ContainsPlayer(clientState.Client.Player.Name) {
		uh.serverState.UpdatePlayer(clientState.Client.Player.Name, clientState.Client.Player.Inner)
	} else {
		uh.serverState.AddPlayer(clientState.Client.Player.Name, clientState.Client.Player.Inner)
	}

	for entry := range uh.connectionsMap.Iter() {
		if clientID == entry.ID {
			continue
		}

		otherPlayers := uh.serverState.FilterPlayers(func(key string, _ ctypes.Player) bool {
			return key != entry.ID // we want only ids that aren't the current entry's ID
		})

		_, err := entry.Conn.Write(state.WithUpdatedPlayers(int(uh.updateID.Load()), otherPlayers))
		if err != nil {
			uh.logger.Errorf("[UDP: handleConnection] Could not send other player state: %s", err.Error())
		}

		if len(uh.serverState.GetPlayers()) >= 2 {
			_, err := entry.Conn.Write(state.WithServerMakingPlayerAbleToMove())
			if err != nil {
				uh.logger.Errorf("[UDP: handleConnection] Could not make player movable: %s", err.Error())
			}
		}
	}
}

func (uh *UDPHandler) Handle() error {
	uh.logger.Infof("Started game data socket (UDP) on %d\n", uh.connInfo.Port())

	// Async closure to handle the closing of the socket, waits for the gracefulCloseChannel, and exits when it receives anything
	go func() {
		<-uh.closeChannel
		uh.logger.Infof("[UDP] Stopping...")
		uh.socket.Close()
		uh.exitChannel <- true
	}()

	clientState := state.Empty()
	var clientIP string
	var clientPort string
	connectedIDs := make(map[uuid.UUID]int)

outer:
	for {

		select {
		case doClose := <-uh.exitChannel:
			if doClose {
				break outer
			}
		default:
		}

		size, addr, err := uh.socket.ReadFrom(&clientState)

		if addr == nil {
			continue
		}

		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			} else if strings.Contains(err.Error(), "nothing read") {
				continue
			}

			uh.logger.Errorf("[UDP] %s\n", err)
			continue
		}

		if size == 0 {
			continue
		}

		uh.updateID.Add(1)

		if clientState.Message == state.Messages.FROM_CLIENT {
			clientData := clientState.Client

			switch clientState.Submessage {
			case state.Submessages.CLIENT_SENDING_UDP_PORT:
				portColonIndex := strings.LastIndex(addr.String(), ":")
				clientIP = addr.String()[:portColonIndex]
				clientPort = clientData.UDPPort

				uh.logger.Debugf("[UDP] Initial connection with client at %s: %s", addr, clientState)
				id, err := uuid.NewRandom()
				if err != nil {
					uh.logger.Errorf("[UDP] Could not pre-generate UUID: %s", err)
				}
				connectedIDs[id] = uh.connectedAmount

				clientConn, err := typedsockets.DialUDP[state.State](clientIP, clientPort)
				if err != nil {
					uh.logger.Errorf("[UDP] Could not connect to client at '%s:%s'", clientIP, clientPort)
					return err
				}

				uh.connectionsMap.UpdateConnection(id.String(), clientConn)
				uh.logger.Infof("[UDP] Connected to client's UDP socket at %s:%s. Client ID: %s", clientIP, clientPort, id)

				_, err = clientConn.Write(state.WithNewClientConnection(id, connectedIDs[id]))
				if err != nil {
					uh.logger.Errorf("[UDP] Couldn't send to client: %s", err.Error())
					return err
				}
				uh.logger.Infof("[UDP] Sent initial data to client at %s:%s", clientIP, clientPort)

				continue
			case state.Submessages.CLIENT_SENDING_LOCAL_DATA:
				uh.logger.Tracef("[UDP] Receiving client local data from: %s", clientData.ID.UUID.String())
				fallthrough
			case state.Submessages.CLIENT_READY:
				id := clientData.ID.UUID.String()

				if uh.connectionsMap.ContainsConnection(id) {
					if _, ok := uh.connectionSlots[clientData.ID.UUID]; !ok && uh.connectedAmount <= 4 {
						uh.connectionSlots[clientData.ID.UUID] = uh.connectedAmount
						connectedIDs[clientData.ID.UUID] = uh.connectedAmount

						uh.connectedAmount++
					}

					uh.logger.Tracef("[UDP] Handling connection for %s", id)
					uh.handleConnection(id, clientState)
				} else {
					uh.logger.Errorf("[UDP] Client with id '%s' not found", clientState.Client.ID.UUID)
					continue
				}
			case state.Submessages.CLIENT_REQUESTING_UPDATE_ID:
				id := clientData.ID.UUID.String()
				requestedUpdateID := clientData.UpdateID

				if !uh.connectionsMap.ContainsConnection(id) {
					uh.logger.Errorf("[UDP] Client with id '%s' not found", id)
				}

				prevState, ok := uh.updates[requestedUpdateID]

				if !ok {
					uh.logger.Errorf("[UDP] Requested update with id '%d' not found", requestedUpdateID)
				}

				prevState.SetAsResending()

				conn := uh.connectionsMap.GetConnection(id)
				_, err := conn.Write(prevState)
				if err != nil {
					uh.logger.Errorf("[UDP] Could not resend update with id '%d': %s", requestedUpdateID, err)
				}
			case state.Submessages.CLIENT_DISCONNECTING:
				id := clientData.ID.UUID.String()
				name := clientData.Player.Name

				uh.connectedAmount--
				delete(uh.connectionSlots, clientData.ID.UUID)

				uh.handleDisconnection(id, name)

				uh.logger.Infof("[UDP] Disconnected from client with id: %s", id)
			}
		}

		uh.updates[uh.updateID.Load()] = uh.serverState.Copy()
	}

	uh.logger.Warn("[UDP] Closed")
	return nil
}
