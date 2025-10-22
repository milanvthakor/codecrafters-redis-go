package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var errInvalidCmd = errors.New("invalid command")

func handleEchoCmd(cmd []*RespVal) (string, error) {
	if len(cmd) < 2 {
		return "", errInvalidCmd
	}
	return ToSimpleStr(cmd[1].BulkStrs()), nil
}

func handleSetCmd(cmd []*RespVal) (string, error) {
	if len(cmd) < 3 {
		return "", errInvalidCmd
	}

	var exp time.Duration
	if len(cmd) > 3 { // Optional exp arg is present
		if len(cmd) != 5 {
			return "", errInvalidCmd
		}

		flag := strings.ToUpper(cmd[3].BulkStrs())
		dur, err := strconv.ParseInt(cmd[4].BulkStrs(), 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid expiry value")
		}

		switch flag {
		case "EX":
			exp = time.Duration(dur * int64(time.Second))
		case "PX":
			exp = time.Duration(dur * int64(time.Millisecond))
		default:
			return "", fmt.Errorf("invalid expiry flag")
		}
	}

	memCache.Set(cmd[1].BulkStrs(), cmd[2].BulkStrs(), exp)
	return ToSimpleStr("OK"), nil
}

func handleGetCmd(cmd []*RespVal) (string, error) {
	if len(cmd) < 2 {
		return "", errInvalidCmd
	}

	if val, ok := memCache.Get(cmd[1].BulkStrs()); !ok {
		return ToNulls(), nil
	} else {
		return ToBulkStr(val), nil
	}
}

func handleRpushCmd(cmd []*RespVal) (string, error) {
	if len(cmd) < 3 {
		return "", errInvalidCmd
	}

	vals := make([]any, len(cmd)-2)
	for i := 2; i < len(cmd); i++ {
		vals[i-2] = cmd[i].BulkStrs()
	}

	listLen := memCache.Rpush(cmd[1].BulkStrs(), vals...)
	return ToIntegers(listLen), nil
}

func handleLrangeCmd(cmd []*RespVal) (string, error) {
	if len(cmd) < 4 {
		return "", errInvalidCmd
	}

	start, err := strconv.Atoi(cmd[2].BulkStrs())
	if err != nil {
		return "", fmt.Errorf("invalid 'start' index")
	}

	stop, err := strconv.Atoi(cmd[3].BulkStrs())
	if err != nil {
		return "", fmt.Errorf("invalid 'stop' index")
	}

	vals := memCache.Lrange(cmd[1].BulkStrs(), start, stop)
	return ToArray(vals), nil
}

func handleLpushCmd(cmd []*RespVal) (string, error) {
	if len(cmd) < 3 {
		return "", errInvalidCmd
	}

	vals := make([]any, len(cmd)-2)
	for i := 2; i < len(cmd); i++ {
		vals[i-2] = cmd[i].BulkStrs()
	}

	listLen := memCache.Lpush(cmd[1].BulkStrs(), vals...)
	return ToIntegers(listLen), nil
}

func handleLlenCmd(cmd []*RespVal) (string, error) {
	if len(cmd) < 2 {
		return "", errInvalidCmd
	}

	len := memCache.Llen(cmd[1].BulkStrs())
	return ToIntegers(len), nil
}

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
			respStr, err = handleEchoCmd(cmd)

		case "SET":
			respStr, err = handleSetCmd(cmd)

		case "GET":
			respStr, err = handleGetCmd(cmd)

		case "RPUSH":
			respStr, err = handleRpushCmd(cmd)

		case "LRANGE":
			respStr, err = handleLrangeCmd(cmd)

		case "LPUSH":
			respStr, err = handleLpushCmd(cmd)

		case "LLEN":
			respStr, err = handleLlenCmd(cmd)
		}

		// Check the error from the command action, if any
		if err != nil {
			fmt.Println(err)
			return
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
