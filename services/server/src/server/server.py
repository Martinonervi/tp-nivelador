import socket
import logger
import safe_socket

from protocol import Protocol
from lottery import Lottery

BATCH_SIZE = 1

class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.storage_path = storage_path

    def _handle_client(self, client_socket):
        lottery = Lottery(self.storage_path)
        protocol = Protocol(client_socket)
        try:
            while True:
                bets, is_fin = protocol.recv_bets()
                if is_fin: break
                lottery.store_bets(bets)

            winners = [bet for bet in lottery.load_bets() if lottery.has_won(bet)]

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

                self._handle_client(client_socket)
