package common

import (
	"errors"
)

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

var (
	ErrNonExistentKey   = errors.New("non-existant key")
	ErrKeyAlreadyLocked = errors.New("key already locked")
)
