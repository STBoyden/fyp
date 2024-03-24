package game

import (
	"fmt"
	"net"
	"strings"

	"fyp/resources"
	"fyp/src/common/ctypes"
	"fyp/src/common/ctypes/state"
	"fyp/src/common/utils/logging"

	typedsockets "fyp/src/common/utils/net/typed-sockets"

	ebitenui "github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

type Game struct {
	ui          *ebitenui.UI
	font        font.Face
	spritesheet ctypes.Spritesheet
	localPlayer ctypes.Player

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
	udpCloseLoopChannel chan interface{}
	rxUDPSocketConn     *state.UDPConnection

	stateChannel chan state.State
	state        state.State
	clientID     uuid.NullUUID
	clientSlot   int
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
		udpCloseLoopChannel: make(chan interface{}),
		stateChannel:        make(chan state.State),
		clientID:            uuid.NullUUID{Valid: false},
		clientSlot:          0,
	}
}

func (g *Game) init() error {
	defer func() { g.initialised = true }()

	if !g.tcpIsConnected {
		address := g.serverAddress + ":" + g.tcpPort

		_, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			g.logger.Errorf("Could not resolve TCP address: %s", err.Error())
			return err
		}

		conn, err := typedsockets.DialTCP[state.State](g.serverAddress, g.tcpPort)
		if err != nil {
			g.logger.Errorf("Could not connect to TCP socket: %s", err.Error())
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
			g.logger.Errorf("Could not resolve UDP address: %s", err.Error())
			return err
		}

		conn, err := typedsockets.DialUDP[state.State](g.serverAddress, g.udpPort)
		if err != nil {
			g.logger.Errorf("Could not connect to TCP socket: %s", err.Error())
			return err
		}

		g.logger.Infof("Connected to server's UDP socket at %s", address)

		socket, err := typedsockets.NewTypedUDPSocketListener[state.State]("0")
		if err != nil {
			g.logger.Errorf("Could not start UDP socket: %s", err.Error())
			return err
		}

		socketConn, err := socket.Conn()
		if err != nil {
			g.logger.Errorf("Could not get connection from socket: %s", err.Error())
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
					g.logger.Errorf("[UDP] Could not get ID from server: %s", err.Error())
					initStateChan <- chanFields{Error: err}
					continue
				}

				initStateChan <- chanFields{State: &initState}
			}
		}()

		res := <-initStateChan
		if err = res.Error; err != nil {
			g.logger.Errorf("[UDP] Could not get initial state from server: %s", err.Error())
			return err
		}

		g.rxUDPSocketConn = socketConn
		g.udpConn = conn
		g.udpIsConnected = true
		g.clientID = res.State.Client.ID
		g.clientSlot = res.State.Client.Slot
	}

	receivedState := state.Empty()

	go func() {
		for {
			if _, err := g.udpConn.Write(state.WithClientReady(g.clientID.UUID)); err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					g.logger.Warnf("Exiting due to unavailable server: %s", err.Error())
					break
				}

				g.logger.Warnf("Couldn't send ready message to server: %s", err.Error())
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
			g.stateChannel <- receivedState
		}
	}()

	if err := g.initUI(); err != nil {
		return err
	}

	if err := g.spritesheet.Load(); err != nil {
		return err
	}

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

	g.ui.Update()
	g.localPlayer.Update()

	go func() {
		<-g.udpCloseLoopChannel
		g.logger.Infof("[UDP-RX] Stopping...")

		defer g.udpConn.Close()
		defer g.rxUDPSocketConn.Close()

		if _, err := g.udpConn.Write(state.WithClientDisconnecting(g.clientID)); err != nil {
			g.logger.Warnf("[UDP-TX] Could not warn server of disconnection: %s", err)
		}
	}()

	select {
	case s := <-g.stateChannel:
		g.logger.Info("Updated from server")
		g.state = s
	default:
	}

	if g.state.Message == state.Messages.FROM_SERVER {
		switch g.state.Submessage {
		case state.Submessages.SERVER_FIRST_CLIENT_CONNECTION_INFORMATION:
			g.clientID = g.state.Client.ID
		default:
			g.logger.Warnf("Unknown or unhandled state submessage: %s", g.state.Submessage.String())
		}
	}

	return nil
}

func (g Game) UpdateServer() {
	_, err := g.udpConn.Write(state.WithUpdatedPlayerState(g.clientID, g.localPlayer))
	if err != nil {
		g.logger.Errorf("error updating server's version of this player via UDP: %s", err.Error())
	}
}

func (g *Game) DebugDrawPlayerSprites(image *ebiten.Image) {
	for playerIndex := ctypes.PlayerMinColour; playerIndex <= ctypes.PlayerMaxColour; playerIndex++ {
		playerSprites, err := g.spritesheet.GetPlayer(playerIndex)
		if err != nil {
			g.logger.Error(err.Error())
		}

		for index, sprite := range playerSprites {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(index)*ctypes.SpriteSizeF, 100+((ctypes.SpriteSizeF+1.0)*float64(playerIndex)))
			op.GeoM.Scale(2, 2)
			image.DrawImage(sprite, op)
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	go g.UpdateServer()

	ebitenutil.DebugPrint(screen, fmt.Sprintf("%.0f FPS", ebiten.ActualFPS()))
	ebitenutil.DebugPrintAt(screen, g.state.Message.String(), screen.Bounds().Dx()/2, screen.Bounds().Dy()/2)
	g.DebugDrawPlayerSprites(screen)

	g.localPlayer.Draw(screen)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return 640, 480
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
