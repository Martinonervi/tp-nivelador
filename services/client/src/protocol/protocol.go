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
	payload := make([]byte, AGENCY_ID_SIZE)
	payload[0] = byte(agencyId)
	return p.sendFrame(MSG_HELLO, payload)
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
	lenBuf := make([]byte, LEN_STR_SIZE)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(s)))
	buffer = append(buffer, lenBuf...)
	return append(buffer, []byte(s)...)
}

func appendDocument(buffer []byte, document int) []byte {
	buf := make([]byte, DOCUMENT_SIZE)
	binary.BigEndian.PutUint32(buf, uint32(document))
	return append(buffer, buf...)
}

func appendBirthdate(buffer []byte, birthdate string) ([]byte, error) {
	date, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, BIRTHDATE_SIZE)
	binary.BigEndian.PutUint16(buf, uint16(date.Year()))
	buf[2] = byte(date.Month())
	buf[3] = byte(date.Day())
	buffer = append(buffer, buf...)
	return buffer, nil
}

func appendBetNumber(buffer []byte, betNumber int) []byte {
	buf := make([]byte, NUMBER_SIZE)
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
	payloadLen := int(binary.BigEndian.Uint16(header[MSG_TYPE_SIZE:HEADER_SIZE]))
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
	if len(payload) < LEN_BETS_SIZE {
		return nil, errShortPayload
	}

	countOfBets := int(binary.BigEndian.Uint16(payload[0:LEN_BETS_SIZE]))
	offset := LEN_BETS_SIZE

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

	tailSize := DOCUMENT_SIZE + BIRTHDATE_SIZE + NUMBER_SIZE
	if offset+tailSize > len(payload) {
		return bet, offset, errShortPayload
	}
	buffer := payload[offset : offset+tailSize]
	bet.Document = int(binary.BigEndian.Uint32(buffer[:DOCUMENT_SIZE]))
	bet.Birthdate = deserializeBirthdate(buffer[DOCUMENT_SIZE : DOCUMENT_SIZE+BIRTHDATE_SIZE])
	bet.Number = int(binary.BigEndian.Uint32(buffer[DOCUMENT_SIZE+BIRTHDATE_SIZE:]))
	offset += tailSize

	return bet, offset, nil
}

func deserializeString(payload []byte, offset int) (string, int, error) {
	if offset+LEN_STR_SIZE > len(payload) {
		return "", offset, errShortPayload
	}
	length := int(binary.BigEndian.Uint16(payload[offset : offset+LEN_STR_SIZE]))
	offset += LEN_STR_SIZE
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
