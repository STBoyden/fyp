package handlers

import (
	"os"
	"os/signal"
	"syscall"

	"fyp/src/common/utils/logging"
)

type SignalHandler struct {
	log                  *logging.Logger
	handles              *[]Handler
	signalChannel        <-chan os.Signal
	gracefulCloseChannel chan<- interface{}
}

func NewSignalHandler(logger *logging.Logger, handlers *[]Handler, gracefulCloseChannel chan<- interface{}) SignalHandler {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	return SignalHandler{
		log:                  logger,
		handles:              handlers,
		signalChannel:        signalChannel,
		gracefulCloseChannel: gracefulCloseChannel,
	}
}

func (sh *SignalHandler) Handle() error {
	for s := range sh.signalChannel {
		if s == os.Interrupt || s == syscall.SIGTERM {
			sh.log.Infof("[SIGNAL HANDLER] Received %s signal, gracefully shutting down", s.String())
			for range len(*sh.handles) {
				sh.gracefulCloseChannel <- nil
			}
			break
		}
	}

	return nil
}
