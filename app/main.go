package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// handleConnection handles the single client connection
func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		// Read the resp value
		respVal, err := readRespVal(conn)
		if err == io.EOF {
			return
		} else if err != nil {
			fmt.Println("Failed to read the value: ", err.Error())
			return
		}

		cmd := respVal.ArrElems()
		if len(cmd) < 1 {
			fmt.Println("Invalid command argument")
			return
		}

		var respStr string

		switch strings.ToUpper(cmd[0].BulkStrs()) {
		case "PING":
			respStr = "PONG"

		case "ECHO":
			if len(cmd) < 2 {
				fmt.Println("invalid command")
				return
			}

			respStr = cmd[1].BulkStrs()
		}

		if _, err := conn.Write([]byte("+" + respStr + "\r\n")); err != nil {
			fmt.Println("Error sending the response: ", err.Error())
			return
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
