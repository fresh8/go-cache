package aerospike

import (
	"time"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/fresh8/go-cache/engine/common"
)

type cl interface {
	Put(policy *as.WritePolicy, key *as.Key, binMap as.BinMap) error
	Get(policy *as.BasePolicy, key *as.Key, binNames ...string) (*as.Record, error)
	Delete(policy *as.WritePolicy, key *as.Key) (bool, error)
}

// Engine is the default Redis storage engine
type Engine struct {
	namespace string
	set       string
	client    cl

	cleanupTimeout time.Duration
}

// NewAerospikeStore creates a new standard Aerospike-backed store
func NewAerospikeStore(namespace, set string, client cl, cleanupTimeout time.Duration) *Engine {
	return &Engine{
		namespace:      namespace,
		set:            set,
		client:         client,
		cleanupTimeout: cleanupTimeout,
	}
}

func getRecord(e *Engine, key string) (*as.Record, error) {
	asKey, err := as.NewKey(e.namespace, e.set, key)
	if err != nil {
		return nil, err
	}

	return e.client.Get(nil, asKey)
}

// Exists checks to see if a key exists in the store
func (e *Engine) Exists(key string) bool {
	record, err := getRecord(e, key)
	if err != nil {
		// TODO: Handle this error properly
		return false
	}

	if record == nil {
		return false
	}

	return true
}

// Get retrieves data from the store based on key, if it exists, else it returns an error
func (e *Engine) Get(key string) (data []byte, err error) {
	record, err := getRecord(e, key)
	if err != nil {
		return
	}

	if record == nil {
		err = common.ErrNonExistentKey
		return
	}

	data, ok := record.Bins["data"].([]byte)
	if !ok {
		// TODO: Handle this error properly
	}

	return
}

// Put stores data against a key, else it returns an error
func (e *Engine) Put(key string, data []byte, expires time.Time) error {
	asKey, err := as.NewKey(e.namespace, e.set, key)
	if err != nil {
		return err
	}

	writePolicy := as.NewWritePolicy(0, uint32(e.cleanupTimeout.Seconds()))

	bins := as.BinMap{
		"expires": expires.Unix(),
		"locked":  0,
		"data":    data,
	}

	return e.client.Put(writePolicy, asKey, bins)
}

// IsExpired checks to see if the key has expired
func (e *Engine) IsExpired(key string) bool {
	record, err := getRecord(e, key)
	if err != nil {
		// TODO: Handle this error properly
		return true
	}

	if record == nil {
		return true
	}

	expires, ok := record.Bins["expires"].(int)
	if !ok {
		// TODO: Handle this error properly
		return false
	}

	return time.Now().Unix() > int64(expires)
}

// Expire marks the key as expired, and removes it from the storage engine
func (e *Engine) Expire(key string) error {
	asKey, err := as.NewKey(e.namespace, e.set, key)
	if err != nil {
		return err
	}

	_, err = e.client.Delete(nil, asKey)
	return err
}

// IsLocked checks to see if the key has been locked
func (e *Engine) IsLocked(key string) bool {
	record, err := getRecord(e, key)
	if err != nil {
		// TODO: Handle this error properly
		return false
	}

	if record == nil {
		return false
	}

	locked, ok := record.Bins["locked"].(int64)
	if !ok {
		// TODO: Handle this error properly
	}

	return locked == 1
}

// Lock sets a lock against the given key
func (e *Engine) Lock(key string) error {
	return nil
}

// Unlock removes the lock from a given key, if it doesn't exist it returns an error
func (e *Engine) Unlock(key string) error {
	return nil
}
