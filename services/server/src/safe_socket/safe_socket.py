import socket

def recv_all(socket: socket.socket, size):
    buffer = b""
    while len(buffer) < size:
        chunk = socket.recv(size - len(buffer))
        if not chunk:
            if not buffer:
                return b""
            raise RuntimeError("Socket connection closed")
        buffer += chunk
    return buffer


def send_all(socket: socket.socket, bytes):
    sent = 0
    while sent < len(bytes):
        sent += socket.send(bytes[sent:])
    return sent
