package protocol

import (
	"testing"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
)

func TestSerializeDeserializeBets(t *testing.T) {
	sent := []lottery.Bet{
		{AgencyId: 1, FirstName: "Martino", LastName: "Nervi", Document: 12345678, Birthdate: "1999-03-01", Number: 7574},
		{AgencyId: 1, FirstName: "Cirilo", LastName: "Pato", Document: 91011113, Birthdate: "2004-05-10", Number: 9325},
	}

	payload := []byte{byte(len(sent))}
	for _, bet := range sent {
		buffer, err := serializeBet(bet)
		if err != nil {
			t.Fatalf("serializeBet error: %v", err)
		}
		payload = append(payload, buffer...)
	}

	received, err := deserializeBets(payload)
	if err != nil {
		t.Fatalf("deserializeBets error: %v", err)
	}
	if len(received) != len(sent) {
		t.Fatalf("esperaba %d apuestas, obtuve %d", len(sent), len(received))
	}
	for i := range sent {
		if received[i] != sent[i] {
			t.Errorf("bet[%d]:\n  got  %+v\n  want %+v", i, received[i], sent[i])
		}
	}
}
