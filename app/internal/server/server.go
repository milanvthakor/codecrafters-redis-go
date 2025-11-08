package server

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/internal/cmd"
	"github.com/codecrafters-io/redis-starter-go/app/internal/errors"
	"github.com/codecrafters-io/redis-starter-go/app/internal/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

// Server represents the Redis server.
type Server struct {
	store    *storage.Mem
	registry *cmd.Registry
}

// NewServer creates a new Redis server instance.
func NewServer() *Server {
	return &Server{
		store:    storage.NewMem(),
		registry: cmd.NewRegistry(),
	}
}

// HandleConnection handles a single client connection.
func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	// Return the response
	returnResp := func(respStr string) {
		if _, err := conn.Write([]byte(respStr)); err != nil {
			fmt.Println("Error sending the response: ", err.Error())
			return
		}
	}

	var (
		isMultiCmdExecuted bool
		// transactions holds the commands issued after the "MULTI" command
		transactions [][]*protocol.RespVal
	)

	for {
		// Read the resp value
		respVal, err := protocol.ReadRespVal(conn)
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

		cmdName := strings.ToUpper(cmd[0].BulkStrs())
		if isMultiCmdExecuted && cmdName != "EXEC" && cmdName != "DISCARD" {
			returnResp(protocol.ToSimpleStr("QUEUED"))
			transactions = append(transactions, cmd)
			continue
		}

		// Handle the command
		var respStr string
		switch cmdName {
		case "MULTI":
			isMultiCmdExecuted = true
			respStr = protocol.ToSimpleStr("OK")

		case "EXEC":
			respStr = s.handleExec(isMultiCmdExecuted, transactions)
			isMultiCmdExecuted = false
			transactions = nil

		case "DISCARD":
			respStr = s.handleDiscard(isMultiCmdExecuted)
			isMultiCmdExecuted = false
			transactions = nil

		default:
			respStr = s.registry.Execute(cmdName, cmd, s.store)
		}

		returnResp(respStr)
	}
}

func (s *Server) handleExec(isMultiCmdExecuted bool, transactions [][]*protocol.RespVal) string {
	if !isMultiCmdExecuted {
		return protocol.ToSimpErr(errors.ErrExecWoMulti.Error())
	}

	// Execute all the commands of the queue
	resps := []string{}
	for _, cmd := range transactions {
		cmdName := strings.ToUpper(cmd[0].BulkStrs())
		resps = append(resps, s.registry.Execute(cmdName, cmd, s.store))
	}

	return protocol.ToArray(resps)
}

func (s *Server) handleDiscard(isMultiCmdExecuted bool) string {
	if !isMultiCmdExecuted {
		return protocol.ToSimpErr(errors.ErrDiscardWoMulti.Error())
	}

	return protocol.ToSimpleStr("OK")
}
