package protocol

import (
	"encoding/binary"
	"net"
	"strings"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
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

func (p *Protocol) SendBet(bet lottery.Bet) error {
	var buffer []byte // podria ser mas optimo si le asigno la length
	buffer = appendString(buffer, bet.FirstName)
	buffer = appendString(buffer, bet.LastName)
	buffer = appendDocument(buffer, bet.Document)
	buffer, _ = appendBirthdate(buffer, bet.Birthdate)
	buffer = appendBetNumber(buffer, bet.Number)
	return safe_socket.SendAll(p.skt, buffer)
}

func appendString(buffer []byte, s string) []byte {
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(s)))
	buffer = append(buffer, lenBuf...)
	return append(buffer, []byte(s)...)
}

func appendDocument(buffer []byte, document int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(document))
	return append(buffer, buf...)
}

func appendBirthdate(buffer []byte, birthdate string) ([]byte, error) {
	date, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf, uint16(date.Year()))
	buf[2] = byte(date.Month())
	buf[3] = byte(date.Day())
	buffer = append(buffer, buf...)
	return buffer, nil
}

func appendBetNumber(buffer []byte, betNumber int) []byte {
	buf := make([]byte, 4) //podria usar 2
	binary.BigEndian.PutUint32(buf, uint32(betNumber))
	buffer = append(buffer, buf...)
	return buffer
}
