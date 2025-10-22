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

		// Check if the command is present or not
		cmd := respVal.ArrElems()
		if len(cmd) < 1 {
			fmt.Println("Invalid command argument")
			return
		}

		// Perform the action as per the command
		var respStr string
		switch strings.ToUpper(cmd[0].BulkStrs()) {
		case "PING":
			respStr = ToSimpleStr("PONG")

		case "ECHO":
			if len(cmd) < 2 {
				fmt.Println("invalid command")
				return
			}

			respStr = ToSimpleStr(cmd[1].BulkStrs())

		case "SET":
			if len(cmd) < 3 {
				fmt.Println("invalid command")
				return
			}

			memCache.Set(cmd[1].BulkStrs(), cmd[2].BulkStrs())
			respStr = ToSimpleStr("OK")

		case "GET":
			if len(cmd) < 2 {
				fmt.Println("invalid command")
				return
			}

			if val, ok := memCache.Get(cmd[1].BulkStrs()); !ok {
				respStr = ToNulls()
			} else {
				respStr = ToBulkStr(val)
			}
		}

		// Return the response
		if _, err := conn.Write([]byte(respStr)); err != nil {
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
