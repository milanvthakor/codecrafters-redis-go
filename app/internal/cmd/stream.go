package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/internal/errors"
	"github.com/codecrafters-io/redis-starter-go/app/internal/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

func handleType(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 2 {
		return "", errors.ErrInvalidCmd
	}

	typ := store.Type(cmd[1].BulkStrs())
	return protocol.ToSimpleStr(typ), nil
}

func handleXadd(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 3 {
		return "", errors.ErrInvalidCmd
	}

	key := cmd[1].BulkStrs()
	id := cmd[2].BulkStrs()
	rawPairs := cmd[3:]
	// Check if key-val pairs are provided correctly
	if len(rawPairs)%2 != 0 {
		return "", fmt.Errorf("invalid key-value pairs")
	}

	pairs := make(map[string]string)
	for i := 0; i < len(rawPairs); i += 2 {
		pairs[rawPairs[i].BulkStrs()] = rawPairs[i+1].BulkStrs()
	}

	storedID, err := store.Xadd(key, &storage.StreamElem{
		ID:    id,
		Pairs: pairs,
	})
	if err != nil {
		return "", err
	}

	return protocol.ToBulkStr(storedID), nil
}

func handleXrange(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 4 {
		return "", errors.ErrInvalidCmd
	}

	stream, err := store.Xrange(cmd[1].BulkStrs(), cmd[2].BulkStrs(), cmd[3].BulkStrs())
	if err != nil {
		return "", err
	}

	return streamToArray(stream), nil
}

func handleXread(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
	if len(cmd) < 2 {
		return "", errors.ErrInvalidCmd
	}

	var (
		timeout  time.Duration = -1
		keyAnIds []*protocol.RespVal
	)
	switch strings.ToUpper(cmd[1].BulkStrs()) {
	case "BLOCK":
		ms, err := strconv.ParseInt(cmd[2].BulkStrs(), 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid 'timeout' value")
		}

		timeout = time.Duration(ms * int64(time.Millisecond))
		keyAnIds = cmd[4:]

	case "STREAMS":
		keyAnIds = cmd[2:]
	}

	if len(keyAnIds)%2 != 0 {
		return "", fmt.Errorf("invalid list of stream keys and ids")
	}

	// Get the keys
	keys := make([]string, 0, len(keyAnIds)/2)
	for _, k := range keyAnIds[:len(keyAnIds)/2] {
		keys = append(keys, k.BulkStrs())
	}
	// Get the IDs
	ids := make([]string, 0, len(keyAnIds)/2)
	for _, k := range keyAnIds[len(keyAnIds)/2:] {
		ids = append(ids, k.BulkStrs())
	}

	// Get the streams
	streams, err := store.Xread(keys, ids, timeout)
	if err != nil {
		return "", err
	}

	if streams == nil {
		return protocol.ToNullArray(), nil
	}

	// Prepare the response
	result := fmt.Sprintf("*%d\r\n", len(streams))
	for i, stream := range streams {
		result += "*2\r\n"
		result += protocol.ToBulkStr(keys[i])
		result += streamToArray(stream)
	}

	return result, nil
}

// streamToArray converts a stream to RESP array format
func streamToArray(stream storage.Stream) string {
	str := fmt.Sprintf("*%d\r\n", len(stream))
	for _, streElem := range stream {
		str += "*2\r\n"
		str += protocol.ToBulkStr(streElem.ID)

		str += fmt.Sprintf("*%d\r\n", len(streElem.Pairs)*2)
		for k, v := range streElem.Pairs {
			str += protocol.ToBulkStr(k) + protocol.ToBulkStr(v)
		}
	}

	return str
}
