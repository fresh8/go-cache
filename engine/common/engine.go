package common

import (
	"errors"
)

// Engine is the interface all caching engines must adhere to
type Engine interface {
	Exists(string) bool
	Get(string) ([]byte, error)
	Put(string, []byte) error

	Expire(string) error
	IsExpired(string) bool

	Lock(string) error
	Unlock(string) error
	IsLocked(string) bool
}

// Errors
var (
	ErrNonExistentKey   = errors.New("non-existant key")
	ErrKeyAlreadyLocked = errors.New("key already locked")
)
