package main

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"fyp/src/cmd/server/handlers"
	"fyp/src/common/utils/env"
	"fyp/src/common/utils/logging"
	"fyp/src/internal/models"
)

var log = logging.NewServer()

// We use this later in the main function to run the UDP and TCP handlers parallel.
func makeParallel(functions ...func() error) {
	var group sync.WaitGroup
	group.Add(len(functions))

	defer group.Wait()

	for _, function := range functions {
		go func(f func() error) {
			defer group.Done()
			if err := f(); err != nil {
				log.Errorf("Error occurred in handle: %s", err.Error())
			}
		}(function)
	}
}

func main() {
	if _, err := env.LoadEnv(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	var tcpPortStr, udpPortStr string
	serverState, serverStateUpdatedChannel := models.NewServerState()
	gracefulCloseChannel := make(chan interface{})

	if _p, isPresent := os.LookupEnv("TCP_PORT"); isPresent {
		tcpPortStr = _p
	} else {
		tcpPortStr = "8080"
	}
	tcpPort, err := strconv.Atoi(tcpPortStr)
	if err != nil {
		log.Errorf("Could not parse TCP_PORT value, expected a value convertable to an integer: %s", err.Error())
		return
	}

	if _p, isPresent := os.LookupEnv("UDP_PORT"); isPresent {
		udpPortStr = _p
	} else {
		udpPortStr = "8081"
	}
	udpPort, err := strconv.Atoi(udpPortStr)
	if err != nil {
		log.Errorf("Could not parse UDP_PORT value, expected a value convertable to an integer: %s", err.Error())
		return
	}

	errorCorrectionSocket, err := net.ListenTCP(
		"tcp",
		&net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: tcpPort},
	)
	if err != nil {
		log.Errorf("Could not start TCP socket listener: %s", err.Error())
		return
	}

	errorCorrectionHandler := handlers.NewErrorCorrectionHandler(log, serverState, errorCorrectionSocket, tcpPort, gracefulCloseChannel)
	correctionSocketFunc := errorCorrectionHandler.Handle

	// Handle UDP connections to the server.
	addr := &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: udpPort}
	gameSocket, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Errorf("Could not start UDP socket: %s", err.Error())
		return
	}

	gameLogicHandler := handlers.NewGameHandler(log, serverState, gameSocket, addr, udpPort, gracefulCloseChannel)
	gameSocketFunc := gameLogicHandler.Handle

	stateHandler := handlers.NewStateHandler(log, serverState, serverStateUpdatedChannel, gracefulCloseChannel)
	stateFunc := stateHandler.Handle
	handlerFunctions := []func() error{correctionSocketFunc, gameSocketFunc, stateFunc}

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	signalFunc := func() error {
		for s := range signalChannel {
			if s == os.Interrupt || s == syscall.SIGTERM {
				log.Infof("[SIGNAL HANDLER] Received %s signal, gracefully shutting down", s.String())
				for range len(handlerFunctions) {
					gracefulCloseChannel <- nil
				}
				break
			}
		}

		return nil
	}

	makeParallel(append(handlerFunctions, signalFunc)...)
	log.Info("Exited")
}
