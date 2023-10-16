package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
)

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
		log.Fatal(":: SERVER :: [FATAL]: Could not parse TCP_PORT value, expected a value convertable to an integer: " + err.Error())
	}

	if _p, isPresent := os.LookupEnv("UDP_PORT"); isPresent {
		udpPortStr = _p
	} else {
		udpPortStr = "8081"
	}
	udpPort, err := strconv.Atoi(udpPortStr)
	if err != nil {
		log.Fatal(":: SERVER :: [FATAL]: Could not parse UDP_PORT value, expected a value convertable to an integer: " + err.Error())
	}

	errorCorrectionSocket, err := net.ListenTCP(
		"tcp",
		&net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: tcpPort},
	)
	if err != nil {
		log.Fatal(":: SERVER :: [FATAL]: Could not start TCP socket listener: " + err.Error())
	}
	correctionSocketFunc := func() {
		log.Printf(":: SERVER :: [INFO]: Started error correction socket (TCP) on %d\n", tcpPort)
		go func() {
			<-gracefulCloseChannel
			log.Println(":: SERVER - TCP :: [INFO]: Stopping...")
			errorCorrectionSocket.Close()
		}()

		for {
			conn, err := errorCorrectionSocket.Accept()

			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}

				log.Println(":: SERVER - TCP :: [ERR]: Could not receive on TCP socket: " + err.Error())
				continue
			}

			log.Printf(":: SERVER - TCP :: [INFO]: Connected with %s\n", conn.RemoteAddr())

			conn.Write([]byte("Hello!"))
		}
	}

	gameSocket, err := net.ListenUDP("udp",
		&net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: udpPort},
	)
	if err != nil {
		log.Fatal(":: SERVER :: [FATAL]: Could not start UDP socket listener: " + err.Error())
	}
	gameSocketFunc := func() {
		log.Printf(":: SERVER :: [INFO]: Started game data socket (UDP) on %d\n", udpPort)
		go func() {
			<-gracefulCloseChannel
			log.Println(":: SERVER - UDP :: [INFO]: Stopping...")
			gameSocket.Close()
		}()

		buf := make([]byte, 2048)

		for {
			_, addr, err := gameSocket.ReadFrom(buf)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}

				log.Printf(":: SERVER - UDP :: [ERR]: %s\n", err)
				continue
			}

			log.Printf(":: SERVER - UDP :: [INFO]: Received data from %s: %s\n", addr, string(buf))
		}
	}

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt)
	signalFunc := func() {
		for s := range signalChannel {
			if s == os.Interrupt {
				gracefulCloseChannel <- struct{}{}
				gracefulCloseChannel <- struct{}{}
				log.Println(":: SERVER - SIGNAL HANDLER :: [INFO]: Received interrupt signal, gracefully shutting down")
				break
			}
		}
	}

	makeParallel(gameSocketFunc, correctionSocketFunc, signalFunc)

}
