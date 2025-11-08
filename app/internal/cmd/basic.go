package cmd

import (
	"github.com/codecrafters-io/redis-starter-go/app/internal/errors"
	"github.com/codecrafters-io/redis-starter-go/app/internal/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

func handlePing(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	return protocol.ToSimpleStr("PONG"), nil
}

func handleEcho(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 2 {
		return "", errors.ErrInvalidCmd
	}
	return protocol.ToSimpleStr(cmd[1].BulkStrs()), nil
}
