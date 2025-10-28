package main

import (
	"io"
	"net"
	"time"
)

// readUntilCRLF reads from the connection until the CRLF appears
func readUntilCRLF(conn net.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	var data string
	for {
		b := make([]byte, 1)
		_, err := conn.Read(b)
		if err == io.EOF {
			return data, err
		}
		if err != nil {
			return "", err
		}

		data += string(b)
		if len(data) >= 2 && data[len(data)-2] == '\r' && data[len(data)-1] == '\n' {
			return data[:len(data)-2], nil
		}
	}
}
