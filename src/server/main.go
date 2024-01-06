package main

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"

	logging "github.com/STBoyden/fyp/src/utils"
)

var log = logging.NewServer()

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
	correctionSocketFunc := func() {
		log.Infof("Started error correction socket (TCP) on %d\n", tcpPort)
		go func() {
			<-gracefulCloseChannel
			log.Infof("[TCP] Stopping...")
			errorCorrectionSocket.Close()
		}()

		for {
			conn, err := errorCorrectionSocket.Accept()

			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}

				log.Errorf("[TCP] Could not receive on TCP socket: %s", err.Error())
				continue
			}

			log.Infof("[TCP] Connected with %s\n", conn.RemoteAddr())

			conn.Write([]byte("Hello!"))
		}
	}

	gameSocket, err := net.ListenUDP("udp",
		&net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: udpPort},
	)
	if err != nil {
		log.Errorf("Could not start UDP socket listener: %s", err.Error())
		return
	}
	gameSocketFunc := func() {
		log.Infof("Started game data socket (UDP) on %d\n", udpPort)
		go func() {
			<-gracefulCloseChannel
			log.Infof("[UDP] Stopping...")
			gameSocket.Close()
		}()

		buf := make([]byte, 2048)

		for {
			_, addr, err := gameSocket.ReadFrom(buf)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}

				log.Errorf("[UDP] %s\n", err)
				continue
			}

			log.Infof("[UDP] Received data from %s: %s\n", addr, string(buf))
		}
	}

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
