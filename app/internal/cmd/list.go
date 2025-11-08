package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/internal/errors"
	"github.com/codecrafters-io/redis-starter-go/app/internal/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

func handleRpush(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 3 {
		return "", errors.ErrInvalidCmd
	}

	vals := make([]any, len(cmd)-2)
	for i := 2; i < len(cmd); i++ {
		vals[i-2] = cmd[i].BulkStrs()
	}

	listLen := store.Rpush(cmd[1].BulkStrs(), vals...)
	return protocol.ToIntegers(int64(listLen)), nil
}

func handleLrange(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 4 {
		return "", errors.ErrInvalidCmd
	}

	start, err := strconv.Atoi(cmd[2].BulkStrs())
	if err != nil {
		return "", fmt.Errorf("invalid 'start' index")
	}

	stop, err := strconv.Atoi(cmd[3].BulkStrs())
	if err != nil {
		return "", fmt.Errorf("invalid 'stop' index")
	}

	vals := store.Lrange(cmd[1].BulkStrs(), start, stop)
	return protocol.ToArray(protocol.ToBulkStrArr(vals)), nil
}

func handleLpush(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 3 {
		return "", errors.ErrInvalidCmd
	}

	vals := make([]any, len(cmd)-2)
	for i := 2; i < len(cmd); i++ {
		vals[i-2] = cmd[i].BulkStrs()
	}

	listLen := store.Lpush(cmd[1].BulkStrs(), vals...)
	return protocol.ToIntegers(int64(listLen)), nil
}

func handleLlen(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 2 {
		return "", errors.ErrInvalidCmd
	}

	len := store.Llen(cmd[1].BulkStrs())
	return protocol.ToIntegers(int64(len)), nil
}

func handleLpop(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 2 {
		return "", errors.ErrInvalidCmd
	}

	remCnt := 1
	if len(cmd) == 3 {
		val, err := strconv.Atoi(cmd[2].BulkStrs())
		if err != nil {
			return "", fmt.Errorf("invalid 'count' value")
		}

		remCnt = val
	}

	removed := store.Lpop(cmd[1].BulkStrs(), remCnt)
	if removed == nil {
		return protocol.ToNulls(), nil
	}
	if len(removed) == 1 {
		return protocol.ToBulkStr(removed[0]), nil
	}

	return protocol.ToArray(protocol.ToBulkStrArr(removed)), nil
}

func handleBlpop(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 3 {
		return "", errors.ErrInvalidCmd
	}

	dur, err := strconv.ParseFloat(cmd[2].BulkStrs(), 64)
	if err != nil {
		return "", fmt.Errorf("invalid expiry value")
	}

	key := cmd[1].BulkStrs()
	removed := store.Blpop(key, time.Duration(dur*float64(time.Second)))
	if removed == nil {
		return protocol.ToArray(nil), nil
	}

	return protocol.ToArray(protocol.ToBulkStrArr([]any{key, removed})), nil
}
