package distconf

import "sync"

type Mem struct {
	vals    map[string][]byte
	watches map[string][]backingCallbackFunction
	mu      sync.RWMutex
}

func (m *Mem) Read(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, exists := m.vals[key]
	if !exists {
		return nil, nil
	}
	return b, nil
}

func (m *Mem) Write(key string, value []byte) error {
	m.mu.Lock()
	if m.vals == nil {
		m.vals = make(map[string][]byte)
	}
	if value == nil {
		delete(m.vals, key)
	} else {
		m.vals[key] = value
	}
	m.mu.Unlock()

	for _, calls := range m.watches[key] {
		calls(key)
	}
	return nil
}

func (m *Mem) Watch(key string, callback backingCallbackFunction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.watches == nil {
		m.watches = make(map[string][]backingCallbackFunction)
	}
	_, existing := m.watches[key]
	if !existing {
		m.watches[key] = []backingCallbackFunction{}
	}
	m.watches[key] = append(m.watches[key], callback)
	return nil
}
