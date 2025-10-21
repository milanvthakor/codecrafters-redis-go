package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

// handleConnection handles the single client connection
func handleConnection(conn net.Conn) {
	for {
		buf := make([]byte, 14) // #bytes for the PING command
		if _, err := conn.Read(buf); err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error reading the connection: ", err.Error())
			os.Exit(1)
		}

		if _, err := conn.Write([]byte("+PONG\r\n")); err != nil {
			fmt.Println("Error sending the response: ", err.Error())
			os.Exit(1)
		}
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}
