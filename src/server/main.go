package main

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/STBoyden/fyp/src/common/utils/logging"
	"github.com/STBoyden/fyp/src/server/handlers"
)

var log = logging.NewServer()

// We use this later in the main function to run the UDP and TCP handlers parallel.
func makeParallel(functions ...func()) {
	var group sync.WaitGroup
	group.Add(len(functions))

	defer group.Wait()

	for _, function := range functions {
		go func(f func()) {
			defer group.Done()
			f()
		}(function)
	}
}

func main() {
	var tcpPortStr string
	var udpPortStr string
	gracefulCloseChannel := make(chan struct{})

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

	errorCorrectionHandler := handlers.NewErrorCorrectionHandler(log, errorCorrectionSocket, tcpPort, gracefulCloseChannel)
	correctionSocketFunc := errorCorrectionHandler.Handle

	// Handle UDP connections to the server.
	addr := &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: udpPort}
	gameSocket, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Errorf("Could not start UDP socket: %s", err.Error())
		return
	}

	gameLogicHandler := handlers.NewGameHandler(log, gameSocket, addr, udpPort, gracefulCloseChannel)
	gameSocketFunc := gameLogicHandler.Handle

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt)
	signalFunc := func() {
		for s := range signalChannel {
			if s == os.Interrupt {
				gracefulCloseChannel <- struct{}{}
				gracefulCloseChannel <- struct{}{}
				log.Infof("[SIGNAL HANDLER] Received interrupt signal, gracefully shutting down")
				break
			}
		}
	}

	makeParallel(gameSocketFunc, correctionSocketFunc, signalFunc)
	log.Info("Exited")
}
