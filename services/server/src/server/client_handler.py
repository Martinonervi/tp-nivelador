import signal
from threading import BrokenBarrierError

import logger

from protocol import Protocol
from lottery import Lottery

WINNERS_BATCH_SIZE = 32

class ClientHandler:
    def __init__(self, client_socket, storage_path, storage_lock, quorum):
        self.protocol = Protocol(client_socket)
        self.storage_path = storage_path
        self.storage_lock = storage_lock
        self.quorum = quorum

    def run(self):
        action = "handle-client"
        signal.signal(signal.SIGTERM, self._shutdown)
        lottery = Lottery(self.storage_path)
        try:
            agency_id = self.protocol.recv_hello()
            while True:
                bets, is_fin = self.protocol.recv_bets(agency_id)
                if is_fin:
                    break
                with self.storage_lock:
                    lottery.store_bets(bets)
                self.protocol.send_ack()

            self.quorum.wait()

            with self.storage_lock:
                winners = [bet for bet in lottery.load_bets()
                           if lottery.has_won(bet) and bet.agency_id == agency_id]

            for i in range(0, len(winners), WINNERS_BATCH_SIZE):
                self.protocol.send_bets(winners[i:i + WINNERS_BATCH_SIZE])
                self.protocol.recv_ack()
            self.protocol.send_no_more_bets()

        except BrokenBarrierError:
            logger.info(action, logger.LogResult.success, "shutdown", True)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, "err", e)
        finally:
            self.protocol.close()

    def _shutdown(self, signum, frame):
        self.protocol.close()