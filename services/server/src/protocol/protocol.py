import safe_socket
from lottery.bet import Bet

class Protocol:
    def __init__(self, skt):
        self.skt = skt

    def send_bet(self, bet):
        buf = b''
        buf += (0x01).to_bytes(1, byteorder='big')
        buf += bet.agency_id.to_bytes(1, byteorder='big')

        first_name_bytes = bet.first_name.encode('utf-8')
        buf += len(first_name_bytes).to_bytes(2, byteorder='big') + first_name_bytes

        last_name_bytes = bet.last_name.encode('utf-8')
        buf += len(last_name_bytes).to_bytes(2, byteorder='big') + last_name_bytes

        buf += bet.document.to_bytes(4, byteorder='big')
        buf += self._pack_birthdate(bet.birthdate)
        buf += bet.number.to_bytes(4, byteorder='big')
        return safe_socket.send_all(self.skt, buf)

    def _pack_birthdate(self, birthdate):
        buf = b''
        year, month, day = map(int, birthdate.split('-'))
        buf += year.to_bytes(2, byteorder='big')
        buf += month.to_bytes(1, byteorder='big')
        buf += day.to_bytes(1, byteorder='big')
        return buf

    def recv_bet(self):
        agency_id = int.from_bytes(safe_socket.recv_all(self.skt, 1), byteorder='big')
        first_name = self._recv_string()
        last_name = self._recv_string()

        buffer = safe_socket.recv_all(self.skt, 12)
        document = int.from_bytes(buffer[:4], byteorder='big')
        birthdate = self._unpack_birthdate(buffer[4:8])
        number = int.from_bytes(buffer[8:], byteorder='big')
        return Bet(agency_id, first_name, last_name, document, birthdate, number)

    def _recv_string(self):
        length = int.from_bytes(safe_socket.recv_all(self.skt, 2), byteorder='big')
        return safe_socket.recv_all(self.skt, length).decode('utf-8')

    def _unpack_birthdate(self, buffer):
        year = int.from_bytes(buffer[:2], byteorder='big')
        month = buffer[2]
        day = buffer[3]
        return f"{year:04d}-{month:02d}-{day:02d}"

    def more_bets(self):
        flag = safe_socket.recv_all(self.skt, 1)
        if not flag:
            raise RuntimeError("connection closed")
        return flag[0] != 0

    def send_no_more_bets(self):
        safe_socket.send_all(self.skt, b'\x00')

    def close(self):
        return self.skt.close()
