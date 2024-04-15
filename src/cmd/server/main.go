package main

import (
	"net"
	"os"
	"strconv"
	"sync"

	"fyp/src/cmd/server/handlers"
	"fyp/src/common/utils/env"
	"fyp/src/common/utils/logging"
	"fyp/src/internal/models"
)

var log = logging.NewServer()

// We use this later in the main function to run the UDP and TCP handlers parallel.
func makeParallel(handles ...handlers.Handler) {
	var group sync.WaitGroup
	group.Add(len(handles))

	defer group.Wait()

	for _, handler := range handles {
		go func(f func() error) {
			defer group.Done()
			if err := f(); err != nil {
				log.Errorf("Error occurred in handle: %s", err.Error())
			}
		}(handler.Handle)
	}
}

func main() {
	if _, err := env.LoadEnv(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	var tcpPortStr, udpPortStr string
	serverState, serverStateUpdatedChannel := models.NewServerState()
	gracefulCloseChannel := make(chan any)

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

	tcpSocket, err := net.ListenTCP(
		"tcp",
		&net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: tcpPort},
	)
	if err != nil {
		log.Errorf("Could not start TCP socket listener: %s", err.Error())
		return
	}

	// Handle UDP connections to the server.
	addr := &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: udpPort}
	udpSocket, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Errorf("Could not start UDP socket: %s", err.Error())
		return
	}

	tcpHandler := handlers.NewTCPHandler(log, serverState, tcpSocket, tcpPort, gracefulCloseChannel)
	udpHandler := handlers.NewUDPHandler(log, serverState, udpSocket, addr, udpPort, gracefulCloseChannel)
	stateHandler := handlers.NewStateHandler(log, serverState, serverStateUpdatedChannel, gracefulCloseChannel)
	handles := []handlers.Handler{tcpHandler, udpHandler, stateHandler}

	signaler := handlers.NewSignalHandler(log, &handles, gracefulCloseChannel)

	makeParallel(append(handles, &signaler)...)
	log.Info("Exited")
}
