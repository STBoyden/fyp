package handlers

import (
	"fyp/src/common/utils/logging"
	"fyp/src/internal/models"
)

type StateHandler struct {
	logger        *logging.Logger
	serverState   *models.ServerState
	updateChannel <-chan string
	closeChannel  <-chan any
	exitChannel   chan bool
}

func NewStateHandler(logger *logging.Logger, serverState *models.ServerState, updateChannel <-chan string, gracefulCloseChannel <-chan any) *StateHandler {
	return &StateHandler{
		logger:        logger,
		serverState:   serverState,
		updateChannel: updateChannel,
		closeChannel:  gracefulCloseChannel,
		exitChannel:   make(chan bool),
	}
}

func (sh *StateHandler) Handle() error {
	go func() {
		<-sh.closeChannel
		sh.logger.Info("[STATE-HANDLER] Stopping...")
		sh.exitChannel <- true
	}()

	for !<-sh.exitChannel {
		select {
		case message := <-sh.updateChannel:
			sh.logger.Infof("[STATE-HANDLER] State updated: %s", message)
		default:
		}
	}

	sh.logger.Info("[STATE-HANDLER] Stopped.")
	return nil
}
