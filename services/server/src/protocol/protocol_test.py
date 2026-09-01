import socket
import struct
import unittest
from protocol import Protocol
from src_frozen.lottery.bet import Bet


class TestProtocol(unittest.TestCase):
    def test_recv_bet(self):
        client, server = socket.socketpair()

        buf = b''
        buf += struct.pack('>B', 1)  # agency_id
        buf += struct.pack('>H', len("Santiago Lionel")) + b"Santiago Lionel"  # first_name
        buf += struct.pack('>H', len("Lorca")) + b"Lorca"  # last_name
        buf += struct.pack('>I', 30904465)  # document
        buf += struct.pack('>H', 1999) + bytes([3, 17])  # birthdate
        buf += struct.pack('>I', 7574)  # number

        client.sendall(buf)
        client.close()

        p = Protocol(server)
        received = p.recv_bet()
        server.close()

        self.assertEqual(received, Bet(1, "Santiago Lionel", "Lorca", 30904465, "1999-03-17", 7574))


if __name__ == '__main__':
    unittest.main()