package protocol

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

const (
	MSG_BATCH byte = 0x00
	MSG_FIN   byte = 0x01
	MSG_ACK   byte = 0x02
	MSG_HELLO byte = 0x03

	MSG_TYPE_SIZE    = 1
	// LEN_PAYLOAD_SIZE solo lo usa el HEADER_SIZE, en GO el tamaño real lo fija AppendUint16,
	// pero la dejo por simetria con el protocolo de python.
	LEN_PAYLOAD_SIZE = 2
	HEADER_SIZE      = MSG_TYPE_SIZE + LEN_PAYLOAD_SIZE
	AGENCY_ID_SIZE   = 1

	LEN_BETS_SIZE = 2
	LEN_STR_SIZE  = 2

	DOCUMENT_SIZE  = 4
	BIRTHDATE_SIZE = 4
	NUMBER_SIZE    = 4

	MAX_PAYLOAD_SIZE = 65535
)
