package memory

import (
	"bytes"
	"testing"
	"time"

	"github.com/fresh8/go-cache/engine/common"
)

func TestInMemory_NewMemoryStore(t *testing.T) {
	memStore := NewMemoryStore(time.Second * 60)

	if len(memStore.store) > 0 {
		t.Fatalf("store length should be 0 on initialisation, %d given", len(memStore.store))
	}

	if len(memStore.locks) > 0 {
		t.Fatalf("locks length should be 0 on initialisation, %d given", len(memStore.locks))
	}

	if len(memStore.expire) > 0 {
		t.Fatalf("expire length should be 0 on initalisation, %d given", len(memStore.expire))
	}
}

func TestInMemory_Exists(t *testing.T) {
	content := []byte("hello")

	memStore := NewMemoryStore(time.Second * 60)

	if memStore.Exists("existing") {
		t.Fatal("key does not exist, marked as existing")
	}

	memStore.store["existing"] = content

	if !memStore.Exists("existing") {
		t.Fatal("key exist, marked as non-existent")
	}

	delete(memStore.store, "existing")

	if memStore.Exists("existing") {
		t.Fatal("key does not exist, marked as existing")
	}
}

func TestInMemory_Get(t *testing.T) {
	content := []byte("hello")

	memStore := NewMemoryStore(time.Second * 60)
	memStore.store["existing"] = content

	_, err := memStore.Get("non-existant")
	if err != common.ErrNonExistentKey {
		t.Fatalf("key does not exist, should return error")
	}

	data, err := memStore.Get("existing")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("%s expected, %s given", content, data)
	}

	newContent := append(content, []byte("existing")...)
	memStore.store["existing"] = newContent
	data, err = memStore.Get("existing")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if bytes.Compare(data, content) == 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", newContent, data)
	}
}

func TestInMemory_Put(t *testing.T) {
	content := []byte("hello")

	memStore := NewMemoryStore(time.Second * 60)

	err := memStore.Put("new-key", content, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	data, ok := memStore.store["new-key"]
	if !ok {
		t.Fatal("key has not been set in internal store")
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("%s expected, %s given", content, data)
	}
}

func TestInMemory_IsLocked(t *testing.T) {
	memStore := NewMemoryStore(time.Second * 60)

	if memStore.IsLocked("not-locked") {
		t.Fatal("newly initialised store should contain no locks")
	}

	memStore.locks["locked-key"] = true

	if !memStore.IsLocked("locked-key") {
		t.Fatal("key should be locked")
	}

	delete(memStore.locks, "locked-key")

	if memStore.IsLocked("locked-key") {
		t.Fatal("key lock should have been released")
	}
}

func TestInMemory_Lock(t *testing.T) {
	memStore := NewMemoryStore(time.Second * 60)

	if len(memStore.locks) != 0 {
		t.Fatalf("locks length should be 0 on initialisation, %d given", len(memStore.locks))
	}

	err := memStore.Lock("lock-me")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if len(memStore.locks) != 1 {
		t.Fatalf("locks length should be 1 after single lock run, %d given", len(memStore.locks))
	}

	err = memStore.Lock("lock-me")
	if err != common.ErrKeyAlreadyLocked {
		t.Fatal("lock already exists, error expected")
	}

	err = memStore.Lock("another-lock")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if len(memStore.locks) != 2 {
		t.Fatalf("locks length should be 2 after double lock run, %d given", len(memStore.locks))
	}
}

func TestInMemory_Unlock(t *testing.T) {
	memStore := NewMemoryStore(time.Second * 60)

	if len(memStore.locks) != 0 {
		t.Fatalf("locks length should be 0 on initialisation, %d given", len(memStore.locks))
	}

	err := memStore.Unlock("locked-key")
	if err != common.ErrNonExistentKey {
		t.Fatalf("key does not exist, should return error")
	}

	memStore.locks["locked-key"] = true

	if len(memStore.locks) != 1 {
		t.Fatalf("locks length should be 1, %d given", len(memStore.locks))
	}

	err = memStore.Unlock("locked-key")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if len(memStore.locks) != 0 {
		t.Fatalf("locks length should be 0 after unlock, %d given", len(memStore.locks))
	}
}

func TestInMemory_IsExpired(t *testing.T) {
	content := []byte("hello")
	memStore := NewMemoryStore(time.Second * 10)
	//Check if key has expired
	if !memStore.IsExpired("existing") {
		t.Fatal("memory store should return true if the key has expired")
	}

	err := memStore.Put("existing", content, time.Now().Add(time.Hour*1))
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	//Check if key has expired
	if memStore.IsExpired("existing") {
		t.Fatal("memory store should return false if the key has not expired")
	}

	//Force expiry
	memStore.expire["existing"] = time.Now()

	//Wait until the cleanup poll has passed
	time.After(time.Second * 10)

	//Check if key has auto expired
	if !memStore.IsExpired("existing") {
		t.Fatal("memory store should return true if the key has expired")
	}
}

func TestInMemory_Expire(t *testing.T) {
	content := []byte("hello")

	memStore := NewMemoryStore(time.Second * 60)

	err := memStore.Expire("existing")
	if err != common.ErrNonExistentKey {
		t.Fatalf("key does not exist, should return error")
	}

	memStore.store["existing"] = content
	memStore.locks["existing"] = true

	err = memStore.Expire("existing")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if len(memStore.store) != 0 {
		t.Fatalf("store length should be 0 after expiring, %d given", len(memStore.store))
	}

	if len(memStore.locks) != 0 {
		t.Fatalf("locks length should be 0 after expiring, %d given", len(memStore.locks))
	}

	if len(memStore.expire) != 0 {
		t.Fatalf("expire length should be 0 after expiring, %d given", len(memStore.expire))
	}
}

func TestInMemory_PollExpire(t *testing.T) {
	content := []byte("hello")

	memStore := NewMemoryStore(time.Second * 1)

	err := memStore.Expire("existing")
	if err != common.ErrNonExistentKey {
		t.Fatalf("key does not exist, should return error")
	}

	memStore.store["existing"] = content
	memStore.locks["existing"] = true
	memStore.expire["existing"] = time.Now()

	//Wait until the cleanup poll has passed
	time.After(time.Second * 1)

	err = memStore.Expire("existing")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if len(memStore.store) != 0 {
		t.Fatalf("store length should be 0 after expiring, %d given", len(memStore.store))
	}

	if len(memStore.locks) != 0 {
		t.Fatalf("locks length should be 0 after expiring, %d given", len(memStore.locks))
	}

	if len(memStore.expire) != 0 {
		t.Fatalf("expire length should be 0 after expiring, %d given", len(memStore.expire))
	}
}

func TestInMemory_copyExpiredKeys(t *testing.T) {
	content := []byte("hello")
	memStore := NewMemoryStore(time.Second * 1)

	err := memStore.Put("new-key", content, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	for _, k := range memStore.copyExpiredKeys() {
		if !memStore.Exists(k) {
			t.Fatalf("no error expected, %s given", err)
		}
	}

}

func TestInMemory_processExpiredKeys(t *testing.T) {
	content := []byte("hello")

	memStore := NewMemoryStore(time.Second * 1)

	err := memStore.Expire("existing")
	if err != common.ErrNonExistentKey {
		t.Fatalf("key does not exist, should return error")
	}

	memStore.store["existing"] = content
	memStore.locks["existing"] = true
	memStore.expire["existing"] = time.Now()

	//Wait until the cleanup poll has passed
	time.After(time.Second * 1)

	//Run process to expire keys
	memStore.processExpiredKeys()

	err = memStore.Expire("existing")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if len(memStore.store) != 0 {
		t.Fatalf("store length should be 0 after expiring, %d given", len(memStore.store))
	}

	if len(memStore.locks) != 0 {
		t.Fatalf("locks length should be 0 after expiring, %d given", len(memStore.locks))
	}

	if len(memStore.expire) != 0 {
		t.Fatalf("expire length should be 0 after expiring, %d given", len(memStore.expire))
	}
}
