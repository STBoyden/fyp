package net

import (
	"fyp/src/common/state"
	"fyp/src/common/utils/logging"
	typedsockets "fyp/src/common/utils/net/typed-sockets"
	"net"
	"strings"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Net struct {
	initialised bool
	tcpPort     string
	udpPort     string

	logger *logging.Logger

	id                  string
	serverAddress       string
	tcpConn             *typedsockets.TCPTypedConnection[state.State]
	tcpIsConnected      bool
	udpConn             *typedsockets.UDPTypedConnection[state.State]
	udpIsConnected      bool
	udpCloseLoopChannel chan struct{}
	rxUDPSocketConn     *typedsockets.UDPTypedConnection[state.State]

	message string
}

func New(
	serverAddress string, tcpPort string, udpPort string, logger *logging.Logger,
) *Net {
	return &Net{
		initialised:         false,
		serverAddress:       serverAddress,
		tcpPort:             tcpPort,
		udpPort:             udpPort,
		logger:              logger,
		udpCloseLoopChannel: make(chan struct{}),
	}
}

func (g *Net) init() error {
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

		s := state.State{ClientUDPPort: portStr}

		bytesWritten, err := conn.Write(s)
		if err != nil {
			g.logger.Errorf("Could not send port to server: %s", err.Error())
			return err
		}

		g.logger.Debugf("[UDP NET-INIT] Wrote %d bytes to server at %s: %s", bytesWritten, conn.RemoteAddr().String(), s)

		g.rxUDPSocketConn = socketConn
		g.udpConn = conn
		g.udpIsConnected = true
	}

	return nil
}

func (g *Net) Update() error {
	if !g.initialised {
		err := g.init()
		if err != nil {
			return err
		}
	}

	go func() {
		<-g.udpCloseLoopChannel
		g.logger.Infof("[UDP-RX] Stopping...")
		g.rxUDPSocketConn.Close()
	}()

	go func() {
		receivedState := state.Empty()

		for {
			if _, err := g.udpConn.Write(state.State{ClientReady: true}); err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					g.logger.Warnf("Exiting due to unavailable server: %s", err.Error())
					break
				}

				g.logger.Warnf("Couldn't send ready message to server: %s", err.Error())
				continue
			}

			size, err := g.rxUDPSocketConn.Read(&receivedState)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					g.logger.Warn("[UDP-RX] Closed")
					break
				}

				continue
			}

			g.logger.Infof("[UDP-RX] Received %d bytes from server: %s", size, receivedState)
			g.message = receivedState.String()
		}
	}()

	return nil
}

func (g *Net) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, g.message)
}

func (g *Net) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth / 2, outsideHeight / 2
}

func (g *Net) Delete() error {
	if g.tcpIsConnected {
		g.tcpConn.Close()
	}

	if g.udpIsConnected {
		g.udpConn.Close()
		g.udpCloseLoopChannel <- struct{}{}
	}

	return nil
}
