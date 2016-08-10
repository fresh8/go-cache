package memory

import (
	"sync"

	"github.com/fresh8/go-cache/engine/common"
)

// Engine is the default memory storage engine
type Engine struct {
	store map[string][]byte
	locks map[string]bool
}

var (
	storeLock sync.RWMutex
	// TODO: Better name needed =|
	locksLock sync.RWMutex
)

// NewMemoryStore creates a new standard in memory store
func NewMemoryStore() *Engine {
	return &Engine{
		store: make(map[string][]byte),
		locks: make(map[string]bool),
	}
}

// Exists checks to see if a key exists in the store
func (e *Engine) Exists(key string) bool {
	_, ok := e.store[key]
	return ok
}

// Get retrieves data from the store based on key, if it exists, else it returns an error
func (e *Engine) Get(key string) (data []byte, err error) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	if !e.Exists(key) {
		err = common.ErrNonExistentKey
	}

	data = e.store[key]

	return
}

// Put stores data against a key, else it returns an error
func (e *Engine) Put(key string, data []byte) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	e.store[key] = data

	return nil
}

// IsExpired checks to see if the key has expired
func (e *Engine) IsExpired(string) bool {
	return false
}

// Expire marks the key as expired, and removes it from the storage engine
func (e *Engine) Expire(key string) error {
	_, ok := e.store[key]
	if !ok {
		return common.ErrNonExistentKey
	}

	delete(e.store, key)
	e.Unlock(key)

	return nil
}

// IsLocked checks to see if the key has been locked
func (e *Engine) IsLocked(key string) bool {
	locksLock.RLock()
	defer locksLock.RUnlock()

	_, ok := e.locks[key]

	return ok
}

// Lock sets a lock against the given key
func (e *Engine) Lock(key string) error {
	if e.IsLocked(key) {
		return common.ErrKeyAlreadyLocked
	}

	locksLock.Lock()
	defer locksLock.Unlock()

	e.locks[key] = true

	return nil
}

// Unlock removes the lock from a given key, if it doesn't exist it returns an error
func (e *Engine) Unlock(key string) error {
	_, ok := e.locks[key]
	if !ok {
		return common.ErrNonExistentKey
	}

	locksLock.Lock()
	defer locksLock.Unlock()

	delete(e.locks, key)

	return nil
}
