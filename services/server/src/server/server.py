import socket
import logger
import safe_socket
import threading

from protocol import Protocol
from lottery import Lottery

BATCH_SIZE = 1

class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str, agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.storage_path = storage_path
        self.storage_lock = threading.Lock()
        self.agency_quorum_min = agency_quorum_min
        self.agency_quorum_lock = threading.Lock()
        self._agencies_done = 0
        self.quorum_ready = threading.Event()


    def _handle_client(self, client_socket):
        lottery = Lottery(self.storage_path)
        protocol = Protocol(client_socket)
        agency_id = None
        try:
            while True:
                bets, is_fin = protocol.recv_bets()
                if is_fin: break
                if bets:
                    agency_id = bets[0].agency_id
                with self.storage_lock:
                    lottery.store_bets(bets)

            with self.agency_quorum_lock:
                self._agencies_done += 1
                if self._agencies_done >= self.agency_quorum_min:
                    self.quorum_ready.set()
            self.quorum_ready.wait()

            with self.storage_lock:
                winners = [bet for bet in lottery.load_bets() if lottery.has_won(bet) and bet.agency_id == agency_id]

            for i in range(0, len(winners), BATCH_SIZE):
                protocol.send_bets(winners[i:i + BATCH_SIZE]) #python corta el slice si se pasa
            protocol.send_no_more_bets()
        except Exception as e:
            logger.error("handle-client", logger.LogResult.fail, "err", e)
        finally:
            protocol.close()


    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                thread = threading.Thread(target=self._handle_client, args=(client_socket,))
                thread.start()
