package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	errXaddIdIsZero      = fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	errXaddIdIsEqOrSmall = fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
)

// parseStreamID parses the given Xadd stream id
func parseStreamID(id string) (int64, int64, error) {
	tokens := strings.Split(id, "-")
	if len(tokens) != 2 {
		return 0, 0, fmt.Errorf("id doesn't have <milliseconds>-<sequenceNumber> format")
	}

	var ms, seqNum int64

	if tokens[0] == "*" {
		ms = -1
	} else if val, err := strconv.ParseInt(tokens[0], 10, 64); err != nil {
		return 0, 0, fmt.Errorf("invalid <milliseconds> value in the ID")
	} else {
		ms = val
	}

	if tokens[1] == "*" {
		seqNum = -1
	} else if val, err := strconv.ParseInt(tokens[1], 10, 64); err != nil {
		return 0, 0, fmt.Errorf("invalid <sequenceFormat> value in the ID")
	} else {
		seqNum = val
	}

	return ms, seqNum, nil
}

// isValidStreamID checks if the new Xadd stread id is valid as per the format and last id.
// If either part of the ID contains "*", it auto-generates the ID and return it.
// Otherwise, returns the same id.
func isValidStreamID(id, lastID string) (string, error) {
	if id == "*" {
		return fmt.Sprintf("%d-%d", time.Now().UnixMilli(), 0), nil
	}

	// Parse the new ID
	ms, seqNum, err := parseStreamID(id)
	if err != nil {
		return "", err
	}
	if ms == 0 && seqNum == 0 {
		return "", errXaddIdIsZero
	}

	// Parse the last ID
	prevMs, prevSeqNum, _ := parseStreamID(lastID)

	// Generate the auto-increated sequence number
	if seqNum == -1 {
		if ms == 0 {
			seqNum = 1
		} else if prevMs == 0 {
			seqNum = 0
		} else {
			seqNum = prevSeqNum + 1
		}
	}

	// Validate the new ID against the last one
	if ms == prevMs {
		if seqNum <= prevSeqNum {
			return "", errXaddIdIsEqOrSmall
		}
	} else if ms <= prevMs {
		return "", errXaddIdIsEqOrSmall
	}

	return fmt.Sprintf("%d-%d", ms, seqNum), nil
}
