import safe_socket
from lottery.bet import Bet

MSG_BATCH = 0x00
MSG_FIN = 0x01
MSG_ACK = 0x02
MSG_HELLO = 0x03

HEADER_SIZE = 3
MAX_PAYLOAD_SIZE = 65535

SHORT_PAYLOAD = "payload shorter than expected"

class Protocol:
    def __init__(self, skt):
        self.skt = skt

    def close(self):
        return self.skt.close()

    # SEND
    def send_bets(self, bets):
        payload = len(bets).to_bytes(2, "big")
        for bet in bets:
            payload += self._serialize_bet(bet)
        self._send_frame(MSG_BATCH, payload)

    def _send_frame(self, msg_type, payload=b""):
        if len(payload) > MAX_PAYLOAD_SIZE:
            raise ValueError(f"payload too large: {len(payload)} bytes")
        frame = msg_type.to_bytes(1, "big")
        frame += len(payload).to_bytes(2, "big")
        frame += payload
        safe_socket.send_all(self.skt, frame)

    def _serialize_bet(self, bet):
        buf = self._serialize_string(bet.first_name)
        buf += self._serialize_string(bet.last_name)
        buf += bet.document.to_bytes(4, "big")
        buf += self._serialize_birthdate(bet.birthdate)
        buf += bet.number.to_bytes(4, "big")
        return buf

    def _serialize_string(self, s):
        encoded = s.encode("utf-8")
        return len(encoded).to_bytes(2, "big") + encoded

    def _serialize_birthdate(self, birthdate):
        year, month, day = map(int, birthdate.split("-"))
        return year.to_bytes(2, "big") + month.to_bytes(1, "big") + day.to_bytes(1, "big")
    
    def recv_ack(self):
        msg_type, _ = self._recv_frame()
        if msg_type != MSG_ACK:
            raise ValueError(f"expected ACK, got msg_type {msg_type}")

    def send_ack(self):
        self._send_frame(MSG_ACK)

    def send_no_more_bets(self):
        self._send_frame(MSG_FIN)


    # RECV

    def recv_hello(self):
        msg_type, payload = self._recv_frame()
        if msg_type != MSG_HELLO:
            raise ValueError(f"expected MSG_HELLO, got msgType {msg_type}")
        if len(payload) != 1:
            raise ValueError(SHORT_PAYLOAD)
        return payload[0]

    def recv_bets(self, agency_id): # devuelve (bets, is_fin)
        msg_type, payload = self._recv_frame()
        if msg_type == MSG_FIN:
            return [], True
        elif msg_type == MSG_BATCH:
            return self._deserialize_bets(payload, agency_id), False
        else:
            raise ValueError(f"unknown msgType: {msg_type}")

    def _recv_frame(self):
        header = safe_socket.recv_all(self.skt, HEADER_SIZE)
        if not header:
            raise ConnectionError("closed socket")
        payload_len = int.from_bytes(header[1:3], "big")
        payload = safe_socket.recv_all(self.skt, payload_len) if payload_len else b""
        return header[0], payload

    def _deserialize_bets(self, payload, agency_id):
        if len(payload) < 2:
            raise ValueError(SHORT_PAYLOAD)

        count_of_bets = int.from_bytes(payload[0:2], "big")
        offset = 2

        bets = []
        for _ in range(count_of_bets):
            bet, offset = self._deserialize_bet(payload, offset, agency_id)
            bets.append(bet)

        if offset != len(payload):
            raise ValueError(f"{len(payload) - offset} bytes left unparsed")
        return bets

    def _deserialize_bet(self, payload, offset, agency_id):
        first_name, offset = self._deserialize_string(payload, offset)
        last_name, offset = self._deserialize_string(payload, offset)

        if offset + 12 > len(payload):
            raise ValueError(SHORT_PAYLOAD)
        document = int.from_bytes(payload[offset:offset + 4], "big")
        birthdate = self._deserialize_birthdate(payload[offset + 4:offset + 8])
        number = int.from_bytes(payload[offset + 8:offset + 12], "big")
        offset += 12

        return Bet(agency_id, first_name, last_name, document, birthdate, number), offset

    def _deserialize_string(self, payload, offset):
        if offset + 2 > len(payload):
            raise ValueError(SHORT_PAYLOAD)
        length = int.from_bytes(payload[offset:offset + 2], "big")
        offset += 2
        if offset + length > len(payload):
            raise ValueError(SHORT_PAYLOAD)
        return payload[offset:offset + length].decode("utf-8"), offset + length

    def _deserialize_birthdate(self, buffer):
        year = int.from_bytes(buffer[:2], "big")
        return f"{year:04d}-{buffer[2]:02d}-{buffer[3]:02d}"