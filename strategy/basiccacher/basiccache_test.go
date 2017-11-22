package basiccacher

import (
	"bytes"
	"testing"
	"time"

	engine "github.com/fresh8/go-cache/engine/memory"
)

// TODO: This should be replaced by a mock, not use memory engine

func TestCacherGetPut(t *testing.T) {
	var (
		e       = engine.NewMemoryStore(time.Second * 60)
		cache   = NewCacher(e, 5, 5)
		content = []byte("hello")
	)

	// First try to get something which we know doens't exist
	data, err := cache.Get("existing")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if data != nil {
		t.Fatalf("no data expected, %s given", data)
	}

	// put something in the cache and expect it to be there

	err = cache.Put("existing", time.Now().Add(1*time.Minute), content)
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	data, err = cache.Get("existing")
	if bytes.Compare(content, data) != 0 {
		t.Fatalf("data expected to be the same, %s expected, %s given", content, data)
	}

	e.Expire("existing")

	// try to get it again
	data, err = cache.Get("existing")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if data != nil {
		t.Fatalf("no data expected, %s given", data)
	}

	// set it again with different content
	content = []byte("goodbye")

	// use a shorter cache TTL
	err = cache.Put("existing", time.Now().Add(1*time.Second), content)
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	// read it immediately
	data, err = cache.Get("existing")
	if bytes.Compare(content, data) != 0 {
		t.Fatalf("data expected to be the same, %s expected, %s given", content, data)
	}

	<-time.After(2 * time.Second)

	// try to read it again
	// try to get it again
	data, err = cache.Get("existing")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if data != nil {
		t.Fatalf("no data expected, %s given", data)
	}
}
