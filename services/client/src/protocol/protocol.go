package protocol

import (
	"net"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const ECHO_CLIENT_BUFFER_SIZE = 1024

type Protocol struct {
	skt net.Conn
}

func NewProtocol(socket net.Conn) (*Protocol, error) {
	return &Protocol{skt: socket}, nil
}

func (p *Protocol) Send(line string) error {
	message := make([]byte, ECHO_CLIENT_BUFFER_SIZE)
	copy(message, []byte(line))
	if err := safe_socket.SendAll(p.skt, message); err != nil {
		logger.Error("send-message", logger.Fail)
		return err
	}
	return nil
}

func (p *Protocol) Recv() (string, error) {
	responseBuffer, err := safe_socket.RecvAll(p.skt, ECHO_CLIENT_BUFFER_SIZE)
	if err != nil {
		logger.Error("recv-response", logger.Fail)
		return string(responseBuffer), err
	}

	var trimmed string = strings.TrimRight(string(responseBuffer), "\x00")
	return trimmed, nil
}

func (p *Protocol) Close() error {
	err := p.skt.Close()
	if err != nil {
		return err
	}
	return nil
}
