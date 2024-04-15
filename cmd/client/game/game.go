package game

import (
	"net"
	"strings"

	"fyp/common/ctypes"
	"fyp/common/ctypes/state"
	"fyp/common/ctypes/tiles"
	"fyp/common/utils/logging"
	"fyp/resources"

	typedsockets "fyp/common/utils/net/typed-sockets"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

type Game struct {
	ui                  *ebitenui.UI
	font                font.Face
	spritesheet         ctypes.Spritesheet
	tiles               tiles.Tiles
	currentMap          Map
	localPlayer         ctypes.Player
	localPlayerCanMove  bool
	playerUpdateChannel chan ctypes.Player

	screenWidth  int
	screenHeight int

	audioCtx         *audio.Context
	audioPlayer      *audio.Player
	audioMusicVolume float64

	initialised bool
	tcpPort     string
	udpPort     string

	logger *logging.Logger

	id                  string
	serverAddress       string
	tcpConn             *state.TCPConnection
	tcpIsConnected      bool
	udpConn             *state.UDPConnection
	udpIsConnected      bool
	udpCloseLoopChannel chan any
	rxUDPSocketConn     *state.UDPConnection

	stateChannel       chan state.State
	forceUpdateChannel chan state.State
	serverState        state.State
	clientID           uuid.NullUUID
	clientSlot         int
	players            map[string]ctypes.Player
}

func New(
	serverAddress, tcpPort, udpPort string, logger *logging.Logger,
) *Game {
	return &Game{
		audioCtx:            audio.NewContext(44100),
		audioMusicVolume:    0.25,
		localPlayer:         ctypes.Player{},
		initialised:         false,
		serverAddress:       serverAddress,
		tcpPort:             tcpPort,
		udpPort:             udpPort,
		logger:              logger,
		udpCloseLoopChannel: make(chan any),
		stateChannel:        make(chan state.State),
		forceUpdateChannel:  make(chan state.State),
		clientID:            uuid.NullUUID{Valid: false},
		clientSlot:          0,
	}
}

func (g *Game) init() error {
	defer func() { g.initialised = true }()

	if err := g.initUI(); err != nil {
		return err
	}

	if err := g.spritesheet.Load(); err != nil {
		return err
	}

	g.tiles = tiles.Initialise(&g.spritesheet)
	currentMap, err := LoadMapFromFile("resources/maps/01_start.map")
	if err != nil {
		return err
	}
	g.currentMap = *currentMap

	stream, err := resources.GetMusicBgm()
	if err != nil {
		return err
	}

	loop := audio.NewInfiniteLoop(stream, stream.Length())
	player, err := g.audioCtx.NewPlayer(loop)
	if err != nil {
		return err
	}
	g.audioPlayer = player
	g.audioPlayer.SetVolume(g.audioMusicVolume)

	if !g.tcpIsConnected {
		address := g.serverAddress + ":" + g.tcpPort

		_, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			g.logger.Fatalf(false, "Could not resolve TCP address: %s", err.Error())
			return err
		}

		conn, err := typedsockets.DialTCP[state.State](g.serverAddress, g.tcpPort)
		if err != nil {
			g.logger.Fatalf(false, "Could not connect to TCP socket: %s", err.Error())
			return err
		}

		localAddressString := conn.LocalAddr().String()
		portColonIndex := strings.LastIndex(localAddressString, ":")
		g.id = localAddressString[portColonIndex+1:]

		g.logger.Infof("Connected to server at %s", address)

		g.tcpConn = conn
		g.tcpIsConnected = true
	}

	if !g.udpIsConnected {
		address := g.serverAddress + ":" + g.udpPort

		_, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			g.logger.Fatalf(false, "Could not resolve UDP address: %s", err.Error())
			return err
		}

		conn, err := typedsockets.DialUDP[state.State](g.serverAddress, g.udpPort)
		if err != nil {
			g.logger.Fatalf(false, "Could not connect to TCP socket: %s", err.Error())
			return err
		}

		g.logger.Infof("Connected to server's UDP socket at %s", address)

		socket, err := typedsockets.NewTypedUDPSocketListener[state.State]("0")
		if err != nil {
			g.logger.Fatalf(false, "Could not start UDP socket: %s", err.Error())
			return err
		}

		socketConn, err := socket.Conn()
		if err != nil {
			g.logger.Fatalf(false, "Could not get connection from socket: %s", err.Error())
			return err
		}

		localAddress := socketConn.LocalAddr().String()
		portColonIndex := strings.LastIndex(localAddress, ":")
		portStr := localAddress[portColonIndex+1:]

		s := state.WithClientUDPPort(portStr)

		bytesWritten, err := conn.Write(s)
		if err != nil {
			g.logger.Errorf("[UDP] Could not send port to server: %s", err.Error())
			return err
		}

		g.logger.Debugf("[UDP NET-INIT] Wrote %d bytes to server at %s: %s", bytesWritten, conn.RemoteAddr().String(), s)
		g.logger.Debugf("[UDP NET-INIT] Local addr: %s", conn.LocalAddr().String())

		type chanFields struct {
			State *state.State
			Error error
		}

		initStateChan := make(chan chanFields)

		go func() {
			for {
				var initState state.State
				_, _, err = socketConn.ReadFrom(&initState)
				if err != nil {
					if strings.Contains(err.Error(), "use of closed network connection") {
						continue
					}

					g.logger.Errorf("[UDP] Could not get initial state from server: %s", err.Error())
					initStateChan <- chanFields{Error: err}
					continue
				}

				x, y := currentMap.GetSpawnPoint()
				initState.Client.InitialPosition.X = x
				initState.Client.InitialPosition.Y = y

				initStateChan <- chanFields{State: &initState}
				break
			}
		}()

		res := <-initStateChan
		if err = res.Error; err != nil {
			g.logger.Fatalf(false, "[UDP] Could not get initial state from server: %s", err.Error())
			return err
		}

		g.rxUDPSocketConn = socketConn
		g.udpConn = conn
		g.udpIsConnected = true
		g.clientID = res.State.Client.ID
		g.clientSlot = res.State.Client.Slot

		player, err := ctypes.NewPlayer(res.State.Client.Colour, &g.spritesheet, res.State.Client.InitialPosition)
		if err != nil {
			g.logger.Fatalf(false, "[UDP] Could not create player from initial state from server: %s\n\nState received from server: %s", err.Error(), res.State)
			return err
		}

		g.localPlayer = *player
		g.playerUpdateChannel = make(chan ctypes.Player)
	}

	receivedState := state.Empty()

	go func(c <-chan ctypes.Player) {
		for {
			if player, ok := <-c; ok {
				if _, err := g.udpConn.Write(state.WithClientReady(g.clientID.UUID, player)); err != nil {
					if strings.Contains(err.Error(), "connection refused") {
						g.logger.Warnf("Exiting due to unavailable server: %s", err.Error())
						break
					}

					g.logger.Warnf("Couldn't send ready message to server: %s", err.Error())
					continue
				}
			} else {
				continue
			}

			size, _, err := g.rxUDPSocketConn.ReadFrom(&receivedState)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					g.logger.Warn("[UDP-RX] Closed")
					break
				}

				continue
			}

			g.logger.Tracef("[UDP-RX] Received %d bytes from server: %s", size, receivedState)

			if receivedState.Server.PriorityUpdate {
				g.logger.Trace("[UDP] Game received priority update: force updating")
				g.forceUpdateChannel <- receivedState
				continue
			} else if (receivedState.Server.UpdateID - g.serverState.Server.UpdateID) > 500 {
				g.logger.Trace("[UDP] Game is lagging by more than 500 updates: force updating")
				g.forceUpdateChannel <- receivedState
				continue
			}

			g.stateChannel <- receivedState
		}
	}(g.playerUpdateChannel)

	go func() {
		<-g.udpCloseLoopChannel
		g.logger.Infof("[UDP-RX] Stopping...")

		defer g.udpConn.Close()
		defer g.rxUDPSocketConn.Close()

		if _, err := g.udpConn.Write(state.WithClientDisconnecting(g.clientID, g.localPlayer.PlayerSpriteIndex.String())); err != nil {
			g.logger.Warnf("[UDP-TX] Could not warn server of disconnection: %s", err)
		}
	}()

	return nil
}

func (g *Game) initUI() error {
	rootContainer := widget.NewContainer()
	eui := &ebitenui.UI{Container: rootContainer}

	fontFace, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return err
	}

	g.font = truetype.NewFace(fontFace, &truetype.Options{})
	g.ui = eui

	return nil
}

func (g *Game) Update() error {
	if !g.initialised {
		err := g.init()
		if err != nil {
			return err
		}
	}

	if !g.audioPlayer.IsPlaying() {
		g.audioPlayer.Play()
	}

	g.ui.Container.RemoveChildren()

	g.ui.Update()
	if g.localPlayerCanMove {
		g.localPlayer.Update()
	}
	g.tiles.StepAnimateTiles()

	colliding, tile := g.currentMap.IsColliding(int(g.localPlayer.Position.X), int(g.localPlayer.Position.Y))
	if colliding {
		tile := *tile

		switch tile {
		case tiles.Typeses.GROUND_UL_TILE:
		case tiles.Typeses.GROUND_UM_TILE:
		case tiles.Typeses.GROUND_UR_TILE:
		case tiles.Typeses.GROUND_ML_TILE:
		case tiles.Typeses.GROUND_MM_TILE:
		case tiles.Typeses.GROUND_MR_TILE:
			break
		default:
			g.localPlayer.TickPhysics()
		}
	} else {
		g.localPlayer.TickPhysics()
	}

	select {
	case g.playerUpdateChannel <- g.localPlayer:
	default:
	}

	select {
	case s := <-g.forceUpdateChannel:
		g.logger.Trace("Force updated")
		g.serverState = s
	case s := <-g.stateChannel:
		g.logger.Info("Updated from server")
		g.serverState = s
	default:
	}

	if g.serverState.Message == state.Messages.FROM_SERVER {
		switch g.serverState.Submessage {
		case state.Submessages.SERVER_FIRST_CLIENT_CONNECTION_INFORMATION:
			g.clientID = g.serverState.Client.ID
		case state.Submessages.SERVER_UPDATING_PLAYERS:
			g.players = make(map[string]ctypes.Player)

			for name, player := range g.serverState.Server.Players {
				if name == g.localPlayer.PlayerSpriteIndex.String() {
					continue
				}

				g.players[name] = player
			}
		case state.Submessages.SERVER_THIS_CLIENT_CAN_MOVE:
			g.localPlayerCanMove = true
		case state.Submessages.SERVER_THIS_CLIENT_CANNOT_MOVE:
			g.localPlayerCanMove = false

			x, y := g.currentMap.GetSpawnPoint()
			g.localPlayer.Position = ctypes.NewPosition(x, y)
		case state.Submessages.SUBMESSAGE_NONE:
			// do nothing
		default:
			g.logger.Warnf("Unknown or unhandled state submessage: %s", g.serverState.Submessage.String())
		}
	}

	id := g.clientID
	player := g.localPlayer

	go updateServer(id, player, g.udpConn, g.logger)

	return nil
}

func updateServer(clientID uuid.NullUUID, player ctypes.Player, conn *state.UDPConnection, logger *logging.Logger) {
	_, err := conn.Write(state.WithUpdatedPlayerState(clientID, player))
	if err != nil {
		logger.Errorf("error updating server's version of this player via UDP: %s", err.Error())
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.currentMap.Draw(screen, &g.tiles)
	g.localPlayer.Draw(screen)

	if !g.localPlayerCanMove {
		ebitenutil.DebugPrintAt(screen, "Waiting for players...", g.screenWidth/2, g.screenHeight/2)
	}

	for _, player := range g.players {
		player.InitFrames(&g.spritesheet)
		player.RemoteUpdatePosition()
		player.Draw(screen)
	}
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	g.screenWidth = 640
	g.screenHeight = 480

	return g.screenWidth, g.screenHeight
}

func (g *Game) Delete() error {
	if g.tcpIsConnected {
		g.tcpConn.Close()
	}

	if g.udpIsConnected {
		g.udpCloseLoopChannel <- nil
	}

	g.audioPlayer.Close()

	return nil
}
