import socket

# TODO: Complete with a short-read/short-write tolerant implementation


def recv_all(socket: socket.socket, size):
    buffer = b""
    while len(buffer) < size:
        chunk = socket.recv(size - len(buffer))
        if not chunk: # si recibo 0 es senal de que cerro correctame
            if not buffer: # si me quede a la mitad = error
                return b""
            raise RuntimeError("Socket connection closed")
        buffer += chunk
    return buffer


def send_all(socket: socket.socket, bytes):
    sent = 0
    while sent < len(bytes):
        n = socket.send(bytes[sent:])
        sent += n
        if sent == 0:
            raise RuntimeError("Socket connection closed")
    return
