package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gokv/app/internal/errors"
	"gokv/app/internal/protocol"
	"gokv/app/internal/storage"
)

func handleSet(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 3 {
		return "", errors.ErrInvalidCmd
	}

	var exp time.Duration
	if len(cmd) > 3 { // Optional exp arg is present
		if len(cmd) != 5 {
			return "", errors.ErrInvalidCmd
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

	store.Set(cmd[1].BulkStrs(), cmd[2].BulkStrs(), exp)
	return protocol.ToSimpleStr("OK"), nil
}

func handleGet(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 2 {
		return "", errors.ErrInvalidCmd
	}

	if val, ok := store.Get(cmd[1].BulkStrs()); !ok {
		return protocol.ToNulls(), nil
	} else {
		return protocol.ToBulkStr(val), nil
	}
}

func handleIncr(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 2 {
		return "", errors.ErrInvalidCmd
	}

	val, err := store.Incr(cmd[1].BulkStrs())
	if err != nil {
		return "", err
	}

	return protocol.ToNumeric(val), nil
}
