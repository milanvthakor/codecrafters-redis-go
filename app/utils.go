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
func parseStreamID(id string, isSeqNumOpt bool) (int64, int64, error) {
	tokens := strings.Split(id, "-")
	if len(tokens) != 2 {
		if !isSeqNumOpt {
			return 0, 0, fmt.Errorf("id doesn't have <milliseconds>-<sequenceNumber> format")
		} else {
			tokens = append(tokens, "*")
		}
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
	ms, seqNum, err := parseStreamID(id, false)
	if err != nil {
		return "", err
	}
	if ms == 0 && seqNum == 0 {
		return "", errXaddIdIsZero
	}

	// Parse the last ID
	prevMs, prevSeqNum, _ := parseStreamID(lastID, false)

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

// getStartIdxByElemID gets the index in the stream from which the elements with ID greater than or equals to start.
func getStartIdxByElemID(id string, stream Stream) (int, error) {
	ms, seqNum, err := parseStreamID(id, true)
	if err != nil {
		return -1, err
	}

	low, high := 0, len(stream)-1
	for low <= high {
		mid := low + (high-low)/2

		mms, mSeqNum, _ := parseStreamID(stream[mid].ID, false)
		if mms < ms {
			low = mid + 1
		} else if mms > ms {
			high = mid - 1
		} else {
			// Check if seqNum is provided or not
			if seqNum == -1 {
				high = mid - 1 // Move towards the first sequence number
			} else if mSeqNum == seqNum {
				return mid, nil
			} else if mSeqNum < seqNum {
				low = mid + 1
			} else {
				high = mid - 1
			}
		}
	}

	return low, nil
}

// getEndIdxByElemID gets the index in the stream till which the elements with ID less than or equals to end.
func getEndIdxByElemID(id string, stream Stream) (int, error) {
	ms, seqNum, err := parseStreamID(id, true)
	if err != nil {
		return -1, err
	}

	low, high := 0, len(stream)-1
	for low <= high {
		mid := low + (high-low)/2

		mms, mSeqNum, _ := parseStreamID(stream[mid].ID, false)
		if mms < ms {
			low = mid + 1
		} else if mms > ms {
			high = mid - 1
		} else {
			// Check if seqNum is provided or not
			if seqNum == -1 {
				low = mid + 1 // Move towards the last sequence number
			} else if mSeqNum == seqNum {
				return mid, nil
			} else if mSeqNum < seqNum {
				low = mid + 1
			} else {
				high = mid - 1
			}
		}
	}

	return high, nil
}
