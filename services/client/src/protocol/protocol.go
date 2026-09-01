package protocol

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

type Protocol struct {
	skt net.Conn
}

func NewProtocol(socket net.Conn) (*Protocol, error) {
	return &Protocol{skt: socket}, nil
}

func (p *Protocol) Close() error {
	return p.skt.Close()
}

// SEND

func (p *Protocol) SendNoMoreBets() error {
	return safe_socket.SendAll(p.skt, []byte{0x00})
}

func (p *Protocol) SendBet(bet lottery.Bet) error {
	var buffer []byte // podria ser mas optimo si le asigno la length
	buffer = append(buffer, 0x01)
	buffer = append(buffer, byte(bet.AgencyId))
	buffer = appendString(buffer, bet.FirstName)
	buffer = appendString(buffer, bet.LastName)
	buffer = appendDocument(buffer, bet.Document)
	var err error
	buffer, err = appendBirthdate(buffer, bet.Birthdate)
	if err != nil {
		return err
	}
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

// RECV

func (p *Protocol) RecvBet() (lottery.Bet, error) {
	bet := lottery.Bet{}
	aux, err := safe_socket.RecvAll(p.skt, 1)
	if err != nil {
		return bet, err
	}
	bet.AgencyId = int(aux[0])
	bet.FirstName, err = p.recvString()
	if err != nil {
		return bet, err
	}
	bet.LastName, err = p.recvString()
	if err != nil {
		return bet, err
	}
	buffer, err := safe_socket.RecvAll(p.skt, 12)
	if err != nil {
		return bet, err
	}
	bet.Document = int(binary.BigEndian.Uint32(buffer[:4]))
	bet.Birthdate = parseBirthdate(buffer[4:8])
	bet.Number = int(binary.BigEndian.Uint32(buffer[8:]))
	return bet, nil
}

func (p *Protocol) recvString() (string, error) {
	lenBuf, err := safe_socket.RecvAll(p.skt, 2)
	if err != nil {
		return "", err
	}
	length := int(binary.BigEndian.Uint16(lenBuf))
	data, err := safe_socket.RecvAll(p.skt, length)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseBirthdate(buffer []byte) string {
	year := binary.BigEndian.Uint16(buffer[:2])
	month := buffer[2]
	day := buffer[3]
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func (p *Protocol) MoreBets() (bool, error) {
	buf, err := safe_socket.RecvAll(p.skt, 1)
	if err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}
