import os
import sys

import logger
import server

SERVER_HOST = os.environ["SERVER_HOST"]
SERVER_PORT = int(os.environ["SERVER_PORT"])
SERVER_STORAGE = os.environ["SERVER_STORAGE"]


def main():
    logger.init()
    s = server.Server(SERVER_HOST, SERVER_PORT, SERVER_STORAGE)
    try:
        s.run()
    except Exception as e:
        logger.error("server-run", logger.LogResult.fail, "err", e)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
