import multiprocessing as mp
import socket
import logger

from protocol import Protocol
from lottery import Lottery

from .client_handler import ClientHandler
import signal

JOIN_TIMEOUT_SECONDS = 3

class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str, agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.storage_path = storage_path
        self.storage_lock = mp.Lock()
        self.quorum = mp.Barrier(agency_quorum_min)
        self.children = []
        self.server_socket = None
        self.running = True

    def run(self):
        signal.signal(signal.SIGTERM, self._shutdown)
        action = "accept-connection"
        self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

        try:
            self.server_socket.bind((self.server_host, self.server_port))
            self.server_socket.listen()
            while self.running:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = self.server_socket.accept()
                except Exception as e:
                    if self.running:
                        logger.error(action, logger.LogResult.fail, "err", e)
                        raise
                    break
                logger.info(action, logger.LogResult.success)

                handler = ClientHandler(client_socket, self.storage_path, self.storage_lock, self.quorum)
                process = mp.Process(target=handler.run)

                process.start()
                client_socket.close()
                self.children.append(process)

        finally:
            self._close()


    def _shutdown(self, signum, frame):
        self.running = False
        if self.server_socket is not None:
            self.server_socket.close()
        self.quorum.abort()

    def _close(self):
        action = "shutdown"
        logger.info(action, logger.LogResult.in_progress)
        for process in self.children:
            process.terminate()
        for process in self.children:
            process.join(timeout=JOIN_TIMEOUT_SECONDS)
        if self.server_socket is not None:
            self.server_socket.close()
        logger.info(action, logger.LogResult.success, "children", len(self.children))
