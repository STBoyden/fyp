package game

import (
	"net"

	logging "github.com/STBoyden/fyp/src/utils"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	initialised bool
	tcpPort     string
	udpPort     string

	logger *logging.Logger

	serverAddress  string
	tcpConn        *net.TCPConn
	tcpIsConnected bool
	udpConn        *net.UDPConn
	udpIsConnected bool
}

func New(serverAddress, tcpPort, udpPort string, logger *logging.Logger) *Game {
	return &Game{
		initialised:   false,
		serverAddress: serverAddress,
		tcpPort:       tcpPort,
		udpPort:       udpPort,
		logger:        logger,
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

		g.udpConn = conn
		g.udpIsConnected = true
	}

	return nil
}

func (g *Game) Update() error {
	if !g.initialised {
		g.init()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Hello, world!")
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return outsideWidth / 2, outsideHeight / 2
}

func (g *Game) Delete() error {
	if g.tcpIsConnected {
		defer g.tcpConn.Close()
	}

	if g.udpIsConnected {
		defer g.udpConn.Close()
	}

	return nil
}
