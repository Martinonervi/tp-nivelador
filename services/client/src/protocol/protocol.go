package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const (
	MSG_BATCH byte = 0x00
	MSG_FIN   byte = 0x01
	MSG_ACK   byte = 0x02
	MSG_HELLO byte = 0x03

	HEADER_SIZE      = 3
	MAX_PAYLOAD_SIZE = 65535
)

// PROTOCOLO
// => Header(3B) + payload(variable)
// Header:
//		msgType 1B
//		lenPayload 2B
//
// Payload de MSG_HELLO:
//		agencyId 1B
//
// Payload de MSG_BATCH:
//	lenBets 2B
//  N bets:
//		lenStrFirstName 2B
//		firstName (variable)
//		lenStrLastName 2B
//		lastName (variable)
//		Document 4B
//		birthday 4B
//		betNumber 4B

type Protocol struct {
	skt net.Conn
}

func NewProtocol(socket net.Conn) *Protocol {
	return &Protocol{skt: socket}
}

func (p *Protocol) Close() error {
	return p.skt.Close()
}

// SEND

func (p *Protocol) SendAck() error {
	return p.sendFrame(MSG_ACK, nil)
}

func (p *Protocol) SendHello(agencyId int) error {
	return p.sendFrame(MSG_HELLO, []byte{byte(agencyId)})
}

func (p *Protocol) SendNoMoreBets() error {
	return p.sendFrame(MSG_FIN, nil)
}

func (p *Protocol) SendBets(listOfBets []lottery.Bet) error {
	var payload []byte
	payload = binary.BigEndian.AppendUint16(payload, uint16(len(listOfBets)))

	for _, bet := range listOfBets {
		buffer, err := serializeBet(bet)
		if err != nil {
			return err
		}
		payload = append(payload, buffer...)
	}

	if err := p.sendFrame(MSG_BATCH, payload); err != nil {
		return err
	}
	return p.recvAck()
}

func serializeBet(bet lottery.Bet) ([]byte, error) {
	var buffer []byte
	buffer = appendString(buffer, bet.FirstName)
	buffer = appendString(buffer, bet.LastName)
	buffer = appendDocument(buffer, bet.Document)
	buffer, err := appendBirthdate(buffer, bet.Birthdate)
	if err != nil {
		return buffer, err
	}
	buffer = appendBetNumber(buffer, bet.Number)
	return buffer, nil
}

func (p *Protocol) sendFrame(msgType byte, payload []byte) error {
	if len(payload) > MAX_PAYLOAD_SIZE {
		return fmt.Errorf("payload too large: %d bytes", len(payload))
	}
	frame := []byte{msgType}
	frame = binary.BigEndian.AppendUint16(frame, uint16(len(payload)))
	frame = append(frame, payload...)
	return safe_socket.SendAll(p.skt, frame)
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
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(betNumber))
	buffer = append(buffer, buf...)
	return buffer
}

func (p *Protocol) recvAck() error {
	msgType, _, err := p.recvFrame()
	if err != nil {
		return err
	}
	if msgType != MSG_ACK {
		return fmt.Errorf("expected ACK, got msgType %d", msgType)
	}
	return nil
}

// RECV

func (p *Protocol) RecvWinners() ([]lottery.Bet, bool, error) {
	msgType, payload, err := p.recvFrame()
	if err != nil {
		return nil, false, err
	}

	switch msgType {
	case MSG_FIN:
		return nil, false, nil
	case MSG_BATCH:
		winners, err := deserializeBets(payload)
		if err != nil {
			return nil, false, err
		}
		return winners, true, nil
	default:
		return nil, false, fmt.Errorf("unknown msgType: %d", msgType)
	}
}

func (p *Protocol) recvFrame() (byte, []byte, error) {
	header, err := safe_socket.RecvAll(p.skt, HEADER_SIZE)
	if err != nil {
		return 0, nil, err
	}
	payloadLen := int(binary.BigEndian.Uint16(header[1:3]))
	if payloadLen == 0 {
		return header[0], nil, nil
	}
	payload, err := safe_socket.RecvAll(p.skt, payloadLen)
	if err != nil {
		return 0, nil, err
	}
	return header[0], payload, nil
}

var errShortPayload = errors.New("payload shorter than expected")

func deserializeBets(payload []byte) ([]lottery.Bet, error) {
	if len(payload) < 2 {
		return nil, errShortPayload
	}

	countOfBets := int(binary.BigEndian.Uint16(payload[0:2]))
	offset := 2

	bets := make([]lottery.Bet, 0, countOfBets)
	for range countOfBets {
		bet, next, err := deserializeBet(payload, offset)
		if err != nil {
			return nil, err
		}
		bets = append(bets, bet)
		offset = next
	}

	if offset != len(payload) {
		return nil, fmt.Errorf("%d bytes left unparsed", len(payload)-offset)
	}

	return bets, nil
}

func deserializeBet(payload []byte, offset int) (lottery.Bet, int, error) {
	bet := lottery.Bet{}

	firstName, offset, err := deserializeString(payload, offset)
	if err != nil {
		return bet, offset, err
	}
	bet.FirstName = firstName
	lastName, offset, err := deserializeString(payload, offset)
	if err != nil {
		return bet, offset, err
	}
	bet.LastName = lastName

	if offset+12 > len(payload) {
		return bet, offset, errShortPayload
	}
	buffer := payload[offset : offset+12]
	bet.Document = int(binary.BigEndian.Uint32(buffer[:4]))
	bet.Birthdate = deserializeBirthdate(buffer[4:8])
	bet.Number = int(binary.BigEndian.Uint32(buffer[8:]))
	offset += 12

	return bet, offset, nil
}

func deserializeString(payload []byte, offset int) (string, int, error) {
	if offset+2 > len(payload) {
		return "", offset, errShortPayload
	}
	length := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
	offset += 2
	if offset+length > len(payload) {
		return "", offset, errShortPayload
	}
	return string(payload[offset : offset+length]), offset + length, nil
}

func deserializeBirthdate(buffer []byte) string {
	year := binary.BigEndian.Uint16(buffer[:2])
	month := buffer[2]
	day := buffer[3]
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}
