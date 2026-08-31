package safe_socket

import "io"

//TODO: Complete with a short-read/short-write tolerant implementation

func SendAll(socket io.Writer, bytes []byte) error {
	sent := 0
	for sent < len(bytes) {
		n, err := socket.Write(bytes[sent:]) //controla el caso de q cierre justo
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
		read += n
		if err != nil {
			if err == io.EOF && read == size {
				return buff, nil
			}
			return nil, err
		}
	}
	return buff, nil

}
