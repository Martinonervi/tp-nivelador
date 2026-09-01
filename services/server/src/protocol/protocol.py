import struct
import safe_socket
from lottery.bet import Bet

class Protocol:
    def __init__(self, skt):
        self.skt = skt

    def send_bet(self, bet):
        buf = b''
        buf += struct.pack('>B', 0x01)
        buf += struct.pack('>B', bet.agency_id)

        first_name_bytes = bet.first_name.encode('utf-8')
        buf += struct.pack('>H', len(first_name_bytes)) + first_name_bytes

        last_name_bytes = bet.last_name.encode('utf-8')
        buf += struct.pack('>H', len(last_name_bytes)) + last_name_bytes

        buf += struct.pack('>I', bet.document)
        buf += self._parse_birthdate(bet.birthdate)
        buf += struct.pack('>I', bet.number)
        return safe_socket.send_all(self.skt, buf)

    def _parse_birthdate(self, birthdate):
        buf = b''
        year, month, day = map(int, birthdate.split('-'))
        buf += struct.pack('>H', year)
        buf += struct.pack('>B', month)
        buf += struct.pack('>B', day)
        return buf
        
    def recv_bet(self):
        agency_id = struct.unpack('>B', safe_socket.recv_all(self.skt, 1))[0]
        first_name = self._recv_string()
        last_name = self._recv_string()

        #solo quedan cosas fijas, las recibo
        buffer = safe_socket.recv_all(self.skt, 12)
        document = struct.unpack('>I', buffer[:4])[0]
        birthdate = self._recv_birthdate(buffer[4:8])
        number = struct.unpack('>I', buffer[8:])[0]
        return Bet(agency_id, first_name, last_name, document, birthdate, number)

    def _recv_string(self):
        length = struct.unpack('>H', safe_socket.recv_all(self.skt, 2))[0]
        return safe_socket.recv_all(self.skt, length).decode('utf-8')

    def _recv_birthdate(self, buffer):
        year = struct.unpack('>H', buffer[:2])[0]
        month = buffer[2]
        day = buffer[3]
        return f"{year:04d}-{month:02d}-{day:02d}"

    def more_bets(self):
        flag = safe_socket.recv_all(self.skt, 1)
        if not flag:
            return False # tengo que levantar error?
        return flag[0] != 0

    def send_no_more_bets(self):
        safe_socket.send_all(self.skt, b'\x00')

'''
struct.unpack(formato, bytes) 
formato: > big endian 
         B 1 bytes (uint8) 
         H 2 bytes (uint16)
         I 4 bytes (uint32)  
=> siempre devuelve una tupla, podria hacer IH y que me devuelva una tupla de ambos valores
'''