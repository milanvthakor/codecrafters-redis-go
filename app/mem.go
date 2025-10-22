package main

import (
	"sync"
	"time"
)

type Mem struct {
	mu sync.RWMutex
	mp map[string]any
}

var memCache *Mem

func NewMem() *Mem {
	return &Mem{
		mu: sync.RWMutex{},
		mp: make(map[string]any),
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

func (m *Mem) Rpush(key string, vals ...any) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	existVals, ok := m.mp[key].([]any)
	if !ok {
		existVals = []any{}
	}

	existVals = append(existVals, vals...)
	m.mp[key] = existVals

	return len(existVals)
}

func (m *Mem) Lrange(key string, start, stop int) []any {
	if start > stop {
		return []any{}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	vals, ok := m.mp[key].([]any)
	if !ok || start > len(vals) {
		return []any{}
	}

	if stop >= len(vals) {
		stop = len(vals) - 1
	}

	return vals[start : stop+1]
}
