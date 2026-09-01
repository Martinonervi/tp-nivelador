import socket
import logger
import safe_socket

from protocol import Protocol
from lottery import Lottery

_ECHO_SERVER_MESSAGE_SIZE = 1024


class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port

    def _handle_client(self, client_socket):
        protocol = Protocol(client_socket)
        bets = []
        while True:
            flag = protocol.more_bets()
            if not flag: break
            bet = protocol.recv_bet()
            bets.append(bet)

        lottery = Lottery(None) # deberia pasarle un path?
        #lottery.store_bets(bets)

        winners = [bet for bet in bets if lottery.has_won(bet)]
        #winners = [bet for bet in lottery.load_bets() if lottery.has_won(bet)]
        for winner in winners:
            protocol.send_bet(winner)
        protocol.send_no_more_bets()


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


'''
    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0
        try:
            logger.info(action, logger.LogResult.in_progress)
            while True:
                client_message = safe_socket.recv_all(
                    client_socket, _ECHO_SERVER_MESSAGE_SIZE
                )
                if not client_message:
                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        message_amount,
                    )
                    return
                message_amount += 1
                safe_socket.send_all(client_socket, client_message)
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail, "messages-amount", message_amount
            )
            raise e
        '''