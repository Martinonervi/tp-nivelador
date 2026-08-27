package safe_socket

import "io"

//TODO: Complete with a short-read/short-write tolerant implementation

func SendAll(socket io.Writer, bytes []byte) error {
	sent := 0
	for sent < len(bytes) {
		n, err := socket.Write(bytes[sent:])
		if err != nil {
			return err
		}
		sent += n
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	read := 0
	for read < size {
		n, err := socket.Read(buff[read:])
		if err != nil {
			return nil, err //devolver buffer? ver docu
		}
		read += n
	}
	return buff, nil

}
