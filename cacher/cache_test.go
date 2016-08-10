package cacher

import (
	"testing"

	"bytes"
	engine "github.com/fresh8/go-cache/engine/memory"
)

func TestCacher_Setup(t *testing.T) {
	t.Skip()
}

func TestCacher_Get(t *testing.T) {
	e := engine.NewMemoryStore()
	cache := NewCacher(e)
	count := 0
	content := []byte("hello")
	regenerate := func() []byte {
		count = count + 1
		return content
	}

	data, err := cache.Get("existing", regenerate)
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if count != 1 {
		t.Fatalf("regenerate function run count should be 1, %d given", count)
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}

	data, err = cache.Get("existing", regenerate)
	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}

	if count != 1 {
		t.Fatalf("regenerate function run count should be 1, %d given", count)
	}

	e.Expire("existing")

	newContent := append(content, []byte("-world")...)
	data, err = cache.Get("existing", func() []byte {
		count = count + 1
		return newContent
	})

	if bytes.Compare(data, newContent) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", newContent, data)
	}

	if count != 2 {
		t.Fatalf("regenerate function run count should be 1, %d given", count)
	}
}
