import unittest

from protocol import Protocol
from lottery.bet import Bet


class TestProtocol(unittest.TestCase):
    def test_serialize_deserialize_bets(self):
        sent = [
            Bet(1, "Martino", "Nervi", 12345678, "1999-03-01", 7574),
            Bet(1, "Cirilo", "Pato", 91011113, "2004-05-10", 1033),
        ]
        protocol = Protocol(None)

        payload = len(sent).to_bytes(2, "big")
        for bet in sent:
            payload += protocol._serialize_bet(bet)

        received = protocol._deserialize_bets(payload, 1)

        self.assertEqual(received, sent)


if __name__ == "__main__":
    unittest.main()