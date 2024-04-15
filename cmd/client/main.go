package main

import (
	"os"

	"fyp/cmd/client/game"
	"fyp/common/utils/env"
	"fyp/common/utils/logging"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sqweek/dialog"
)

var log = logging.NewClient()

func main() {
	if _, err := env.LoadEnv(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	var tcpPort, udpPort, serverAddress string

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
		dialog.Message(err.Error()).Error()
	}

	err := g.Delete()
	if err != nil {
		log.Fatalf(false, "Error occurred when deleting game: %s", err.Error())
	}

	log.Info("Exited")
}
