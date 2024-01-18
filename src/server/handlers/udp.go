package handlers

import (
	"errors"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/STBoyden/fyp/src/common/utils/logging"

	"github.com/phakornkiong/go-pattern-match/pattern"
)

type GameHandler struct {
	Handler

	logger           *logging.Logger
	connections      map[string]net.Conn
	connectionsMutex sync.Mutex
	socket           *net.UDPConn
	connInfo         netip.AddrPort
	closeChannel     chan struct{}
}

func NewGameHandler(logger *logging.Logger, socket *net.UDPConn, udpHost *net.UDPAddr, udpPort int, gracefulCloseChannel chan struct{}) *GameHandler {
	return &GameHandler{
		logger:           logger,
		connections:      map[string]net.Conn{},
		connectionsMutex: sync.Mutex{},
		socket:           socket,
		connInfo:         netip.AddrPortFrom(udpHost.AddrPort().Addr(), uint16(udpPort)),
		closeChannel:     gracefulCloseChannel,
	}
}

func (gh *GameHandler) Handle() {
	gh.logger.Infof("Started game data socket (UDP) on %d\n", gh.connInfo.Port())

	// Async closure to handle the closing of the socket, waits for the gracefulCloseChannel, and exits when it receives anything
	go func() {
		<-gh.closeChannel
		gh.logger.Infof("[UDP] Stopping...")
		gh.socket.Close()
	}()

	buffer := make([]byte, 2048)

	var port int
	var clientIp net.IP
	var clientAddress net.UDPAddr

	for {
		size, addr, err := gh.socket.ReadFrom(buffer)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				gh.logger.Warn("[UDP] Closed")

				break
			}

			gh.logger.Errorf("[UDP] %s\n", err)
			continue
		}

		if size == 0 {
			continue
		}

		s := string(buffer[:size])

		match := func(input string) error {
			return pattern.NewMatcher[error](input).
				WithPattern(
					pattern.String().StartsWith("ready"),
					func() error {
						clientConn, err := net.DialUDP("udp", nil, &clientAddress)
						if err != nil {
							gh.logger.Errorf("[UDP] Could not connect to client at '%s'", &clientAddress)
							return err
						}
						gh.connectionsMutex.Lock()

						source := clientConn.RemoteAddr().String()

						// already readied - not needed again
						if gh.connections[source] != nil {
							gh.connectionsMutex.Unlock()
							return nil
						}

						// connection is now ready
						gh.connections[source] = clientConn
						gh.connectionsMutex.Unlock()

						gh.logger.Infof("[UDP] Connected to client's UDP socket at %s", source)

						//TODO Logic with sending the game updates and logic here.
						gh.logger.Infof("[UDP] Sending 'hello' message to client at %s", source)
						size, err = clientConn.Write([]byte("Hello from server!"))
						if err != nil {
							gh.logger.Errorf("[UDP] Couldn't send to client: %s", err.Error())
							return err
						}

						return nil
					}).
				WithPattern(
					pattern.String().Regex(regexp.MustCompile(`^\d+$`)),
					func() error {
						port, err = strconv.Atoi(input)
						if err != nil {
							gh.logger.Errorf("[UDP] Could not parse '%s' as a valid port number: %s", input, err.Error())
							return err
						}

						portColonIndex := strings.LastIndex(addr.String(), ":")
						clientIp = net.ParseIP(addr.String()[:portColonIndex-1])
						clientAddress = net.UDPAddr{IP: clientIp, Port: port}

						return nil
					}).
				Otherwise(
					func() error {
						message := "this *SHOULD* be unreachable. If you see this, please send full logs to stboyden@duck.com and a description of what you were doing when this message appeared"

						gh.logger.Error(message)
						return errors.New(message)
					})
		}

		err = match(s)
		if err != nil {
			continue
		}
	}
}
