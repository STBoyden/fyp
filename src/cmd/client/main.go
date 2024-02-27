package main

import (
	"os"

	"fyp/src/cmd/client/game"
	"fyp/src/common/utils/logging"

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

	g := game.New(serverAddress, tcpPort, udpPort, log)

	ebiten.SetWindowTitle("Final Year Project")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(true)
	ebiten.SetVsyncEnabled(true)

	log.Info("Starting game...")
	if err := ebiten.RunGame(g); err != nil {
		log.Error(err.Error())
	}

	err := g.Delete()
	if err != nil {
		log.Errorf("Error occurred when deleting game: %s", err.Error())
	}

	log.Info("Exited")
}
