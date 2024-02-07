package main

import (
	"fyp/src/common/utils/logging"
	"fyp/src/game/net"
	"os"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

var log = logging.NewClient()

func main() {
	tcpPort := ""
	udpPort := ""
	serverAddress := ""

	if _p, isPresent := os.LookupEnv("SERVER_ADDRESS"); isPresent {
		serverAddress = _p
	} else {
		log.Error("SERVER_ADDRESS environment variable not found")
		os.Exit(1)
	}

	if _p, isPresent := os.LookupEnv("SERVER_TCP_PORT"); isPresent {
		tcpPort = _p
	} else {
		log.Error("SERVER_TCP_PORT environment variable not found")
		os.Exit(1)
	}

	if _p, isPresent := os.LookupEnv("SERVER_UDP_PORT"); isPresent {
		udpPort = _p
	} else {
		log.Error("SERVER_UDP_PORT environment variable not found")
		os.Exit(1)
	}

	game := net.New(serverAddress, tcpPort, udpPort, log)

	ebiten.SetWindowSize(1600, 900)
	ebiten.SetWindowTitle("Final Year Project")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(true)

	log.Info("Starting game...")
	if err := ebiten.RunGame(game); err != nil {
		log.Error(err.Error())
	}

	_ = game.Delete()
	log.Info("Exited")
}
