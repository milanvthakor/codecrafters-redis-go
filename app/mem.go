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
