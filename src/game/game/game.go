package game

import (
	"net"
	"strings"

	"github.com/STBoyden/fyp/src/common/utils/logging"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	initialised bool
	tcpPort     string
	udpPort     string

	logger *logging.Logger

	id                  string
	serverAddress       string
	tcpConn             *net.TCPConn
	tcpIsConnected      bool
	udpConn             *net.UDPConn
	udpIsConnected      bool
	udpCloseLoopChannel chan struct{}
	rxUdpSocket         *net.UDPConn

	message string
}

func New(serverAddress, tcpPort, udpPort string, logger *logging.Logger) *Game {
	return &Game{
		initialised:         false,
		serverAddress:       serverAddress,
		tcpPort:             tcpPort,
		udpPort:             udpPort,
		logger:              logger,
		udpCloseLoopChannel: make(chan struct{}),
	}
}

func (g *Game) init() error {
	if !g.tcpIsConnected {
		address := g.serverAddress + ":" + g.tcpPort

		tcpServer, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			g.logger.Errorf("Could not resolve TCP address: %s", err.Error())
			return err
		}

		conn, err := net.DialTCP("tcp", nil, tcpServer)
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

		udpServer, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			g.logger.Errorf("Could not resolve UDP address: %s", err.Error())
			return err
		}

		conn, err := net.DialUDP("udp", nil, udpServer)
		if err != nil {
			g.logger.Errorf("Could not connect to TCP socket: %s", err.Error())
			return err
		}

		g.logger.Infof("Connected to server's UDP socket at %s", address)

		socket, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0})
		if err != nil {
			g.logger.Errorf("Could not start UDP socket: %s", err.Error())
			return err
		}

		localAddress := socket.LocalAddr().String()
		portColonIndex := strings.LastIndex(localAddress, ":")
		portStr := localAddress[portColonIndex+1:]

		_, err = conn.Write([]byte(portStr))
		if err != nil {
			g.logger.Errorf("Could not send port to server: %s", err.Error())
			return err
		}

		g.rxUdpSocket = socket
		g.udpConn = conn
		g.udpIsConnected = true
	}

	return nil
}

func (g *Game) Update() error {
	if !g.initialised {
		g.init()
	}

	go func() {
		<-g.udpCloseLoopChannel
		g.logger.Infof("[UDP-RX] Stopping...")
		g.rxUdpSocket.Close()
	}()

	go func() {
		buffer := make([]byte, 2024)
		for {
			if _, err := g.udpConn.Write([]byte("ready")); err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					g.logger.Warnf("Exiting due to unavailable server: %s", err.Error())
					break
				}

				g.logger.Warnf("Couldn't send ready message to server: %s", err.Error())
				continue
			}

			size, _, err := g.rxUdpSocket.ReadFrom(buffer)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					g.logger.Warn("[UDP-RX] Closed")
					break
				}

				continue
			}

			received := string(buffer[:size])
			g.logger.Infof("[UDP-RX] Received %d bytes from server: %s", size, received)
			g.message = received
		}
	}()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, g.message)
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return outsideWidth / 2, outsideHeight / 2
}

func (g *Game) Delete() error {
	if g.tcpIsConnected {
		g.tcpConn.Close()
	}

	if g.udpIsConnected {
		g.udpConn.Close()
		g.udpCloseLoopChannel <- struct{}{}
	}

	return nil
}
