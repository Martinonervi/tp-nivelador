import socket

# TODO: Complete with a short-read/short-write tolerant implementation


def recv_all(socket: socket.socket, size):
    buffer = b""
    while len(buffer) != size:
        chunk = socket.recv(size - len(buffer))
        if not chunk:
            return b""
        buffer += chunk
    return buffer


def send_all(socket: socket.socket, bytes):
    sent = 0
    while sent != len(bytes):
        n = socket.send(bytes[sent:])
        sent += n
    return
