package main

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

// StreamElem represents the single element/item of the stream.
type StreamElem struct {
	ID    string
	Pairs map[string]string
}

// Stream represents the single stream of elements associated with particular key.
// It maps the ID of stream elements with its struct.
type Stream []*StreamElem

// ListBlockPop is used to handle the blocking "BLPOP" command.
type ListBlockPop struct {
	// waitQ is map of list key and the list of channels of the connection
	// waiting for the new element to get inserted in the list
	waitQ map[string][]chan struct{}
	mu    sync.Mutex
}

type Mem struct {
	mu  sync.RWMutex
	mp  map[string]any // TODO: Make sure a key holds the value of only one type. If user tries to change it, they shouldn't be able to do so if the value exists for it.
	lbp *ListBlockPop
}

var memCache *Mem

func NewMem() *Mem {
	return &Mem{
		mp: make(map[string]any),
		lbp: &ListBlockPop{
			waitQ: make(map[string][]chan struct{}),
		},
	}
}

func init() {
	memCache = NewMem()
}

func (m *Mem) Get(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.mp[key]
	return val, ok
}

func (m *Mem) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.mp, key)
}

func (m *Mem) Set(key string, val any, exp time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mp[key] = val

	if exp > 0 {
		go func() {
			time.Sleep(exp)
			m.Delete(key)
		}()
	}
}

// handleListInsert sends the signal to available connections waiting for the element to be inserted
func (m *Mem) handleListInsert(key string) {
	// Get the length of the list
	var listLen int
	m.mu.RLock()
	if list, ok := m.mp[key].([]any); ok {
		listLen = len(list)
	}
	m.mu.RUnlock()

	// Sent the signal to the connections waiting in the queue
	m.lbp.mu.Lock()
	defer m.lbp.mu.Unlock()

	waitList, ok := m.lbp.waitQ[key]
	if !ok || len(waitList) <= 0 {
		return
	}

waitLoop:
	for i, w := range waitList {
		select {
		case w <- struct{}{}:
			listLen--
		default:
		}

		if listLen == 0 { // Break once there are no more elements to be removed
			if i == -1 || i == len(waitList)-1 {
				delete(m.lbp.waitQ, key)
			} else {
				m.lbp.waitQ[key] = waitList[i+1:]
			}
			break waitLoop
		}
	}
}

func (m *Mem) Rpush(key string, vals ...any) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	existVals, ok := m.mp[key].([]any)
	if !ok {
		existVals = []any{}
	}

	existVals = append(existVals, vals...)
	m.mp[key] = existVals

	go m.handleListInsert(key)

	return len(existVals)
}

func (m *Mem) Lpush(key string, vals ...any) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	existVals, ok := m.mp[key].([]any)
	if !ok {
		existVals = []any{}
	}

	slices.Reverse(vals)
	existVals = append(vals, existVals...)
	m.mp[key] = existVals

	go m.handleListInsert(key)

	return len(existVals)
}

func (m *Mem) Lrange(key string, start, stop int) []any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vals, ok := m.mp[key].([]any)
	if !ok {
		return []any{}
	}

	// Handle negative indexes
	if start < 0 {
		start = max(len(vals)+start, 0)
	}
	if stop < 0 {
		stop = max(len(vals)+stop, 0)
	}

	if start < 0 || stop < 0 || start > stop || start > len(vals) {
		return []any{}
	}

	if stop >= len(vals) {
		stop = len(vals) - 1
	}
	return vals[start : stop+1]
}

func (m *Mem) Llen(key string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vals, ok := m.mp[key].([]any)
	if !ok {
		return 0
	}

	return len(vals)
}

func (m *Mem) Lpop(key string, remCnt int) []any {
	m.mu.Lock()
	defer m.mu.Unlock()

	vals, ok := m.mp[key].([]any)
	if !ok || len(vals) == 0 {
		return nil
	}

	if remCnt >= len(vals) {
		delete(m.mp, key)
		return vals
	}

	removed := vals[:remCnt]
	m.mp[key] = vals[remCnt:]

	return removed
}

func (m *Mem) Blpop(key string, timeout time.Duration) any {
	// Remove the first element, if present.
	removed := m.Lpop(key, 1)
	if removed != nil {
		return removed[0]
	}

	// Wait for an element to be present to get removed
	elemPresSign := make(chan struct{})
	m.lbp.mu.Lock()
	m.lbp.waitQ[key] = append(m.lbp.waitQ[key], elemPresSign)
	m.lbp.mu.Unlock()

	// Handles the no timeout
	if timeout == 0 {
		<-elemPresSign
		return m.Lpop(key, 1)[0]
	}

	// Handles the timeout
	for {
		select {
		case <-time.After(timeout):
			return nil

		case <-elemPresSign:
			return m.Lpop(key, 1)[0]
		}
	}
}

func (m *Mem) Type(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.mp[key]
	if !ok {
		return "none"
	}

	switch val.(type) {
	case string:
		return "string"
	case Stream:
		return "stream"
	case []any:
		return "list"
	default:
		return "none"
	}
}

func (m *Mem) Xadd(key string, elem *StreamElem) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stream, ok := m.mp[key].(Stream)
	if !ok {
		stream = make(Stream, 0)
	}

	// Check if the new ID is valid
	var (
		id  string
		err error
	)
	if len(stream) <= 0 {
		id, err = isValidStreamID(elem.ID, "")
	} else {
		id, err = isValidStreamID(elem.ID, stream[len(stream)-1].ID)
	}
	if err != nil {
		return "", err
	}

	elem.ID = id
	stream = append(stream, elem)
	m.mp[key] = stream

	return id, nil
}

func (m *Mem) Xrange(key, startId, endId string) (Stream, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stream, ok := m.mp[key].(Stream)
	if !ok {
		return nil, nil
	}

	startIdx, err := getStartIdxByElemID(startId, stream)
	if err != nil {
		return nil, err
	}

	endIdx, err := getEndIdxByElemID(endId, stream)
	if err != nil {
		return nil, err
	}

	if startIdx < 0 || startIdx >= len(stream) || endIdx < 0 || endIdx >= len(stream) || startIdx > endIdx {
		return nil, fmt.Errorf("invalid 'startId' or 'endId' parameter were provided")
	}

	result := make(Stream, endIdx-startIdx+1)
	copy(result, stream[startIdx:endIdx+1])

	return result, nil
}
