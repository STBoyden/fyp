package main

import (
	"log"
	"net"
	"os"
	"strconv"
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
		defer errorCorrectionSocket.Close()

		for {
			conn, err := errorCorrectionSocket.Accept()

			if err != nil {
				log.Fatal(":: SERVER - TCP:: [FATAL]: Could not receive on TCP socket: " + err.Error())
			}

			log.Printf(":: SERVER - TCP:: [INFO]: Connected with %s\n", conn.RemoteAddr())

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
		defer gameSocket.Close()

		buf := make([]byte, 2048)

		for {
			_, addr, err := gameSocket.ReadFrom(buf)
			if err != nil {
				log.Printf("::SERVER - UDP::  [WARN]: %s\n", err)
			}

			log.Printf(":: SERVER - UDP:: [INFO]: Received data from %s: %s\n", addr, string(buf))
		}
	}

	makeParallel(gameSocketFunc, correctionSocketFunc)
}
