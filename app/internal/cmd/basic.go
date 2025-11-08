package cmd

import (
	"gokv/app/internal/errors"
	"gokv/app/internal/protocol"
	"gokv/app/internal/storage"
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
