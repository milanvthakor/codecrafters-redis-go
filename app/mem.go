package main

import "sync"

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

func (m *Mem) Set(key string, val any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mp[key] = val
}
