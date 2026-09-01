package protocol

import (
	"net"
	"testing"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
)

func TestSendRecvBet(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	sent := lottery.Bet{
		FirstName: "Santiago",
		LastName:  "Lorca",
		Document:  30904465,
		Birthdate: "1999-03-17",
		Number:    7574,
	}

	go func() {
		p := &Protocol{skt: client}
		if err := p.SendBet(sent); err != nil {
			t.Errorf("SendBet error: %v", err)
		}
	}()

	p := &Protocol{skt: server}
	received, err := p.RecvBet()
	if err != nil {
		t.Fatalf("RecvBet error: %v", err)
	}

	if received.FirstName != sent.FirstName {
		t.Errorf("FirstName: got %q, want %q", received.FirstName, sent.FirstName)
	}
	if received.LastName != sent.LastName {
		t.Errorf("LastName: got %q, want %q", received.LastName, sent.LastName)
	}
	if received.Document != sent.Document {
		t.Errorf("Document: got %d, want %d", received.Document, sent.Document)
	}
	if received.Birthdate != sent.Birthdate {
		t.Errorf("Birthdate: got %q, want %q", received.Birthdate, sent.Birthdate)
	}
	if received.Number != sent.Number {
		t.Errorf("Number: got %d, want %d", received.Number, sent.Number)
	}
}
