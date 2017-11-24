package basiccacher

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fresh8/go-cache/engine/common"
)

func TestGet(t *testing.T) {
	expectedData := []byte("hello")
	// set up
	engine := &common.EngineMock{
		ExistsFunc: func(in1 string) bool {
			if strings.Contains(in1, "EXISTS") {
				return true
			}
			return false
		},
		GetFunc: func(in1 string) ([]byte, error) {
			return expectedData, nil
		},
		IsExpiredFunc: func(in1 string) bool {
			if strings.Contains(in1, "EXPIRED") {
				return true
			}
			return false
		},
	}
	cacher := NewCacher(engine, 5, 5)

	t.Run("key doesn't exist", func(*testing.T) {
		data, err := cacher.Get("NOPE")
		if data != nil {
			t.Errorf("no data expected, got %s", data)
		}
		if err != nil {
			t.Errorf("no error expected, got %s", err.Error())
		}
	})

	t.Run("key exists, not expired", func(*testing.T) {
		data, err := cacher.Get("EXISTS")
		if bytes.Compare(expectedData, data) != 0 {
			t.Errorf("%s expected, got %s", expectedData, data)
		}
		if err != nil {
			t.Errorf("no error expected, got %s", err.Error())
		}
	})

	t.Run("key exists, expired", func(*testing.T) {
		data, err := cacher.Get("EXISTS_EXPIRED")
		if data != nil {
			t.Errorf("no data expected, got %s", data)
		}
		if err != nil {
			t.Errorf("no error expected, got %s", err.Error())
		}
	})
}

func TestPut(t *testing.T) {
	engine := &common.EngineMock{
		IsLockedFunc: func(in1 string) bool {
			if strings.Contains(in1, "LOCKED") {
				return true
			}
			return false
		},
		LockFunc: func(in1 string) error {
			if strings.Contains(in1, "LOCKERROR") {
				return errors.New("lock error")
			}
			return nil
		},
		UnlockFunc: func(in1 string) error {
			if strings.Contains(in1, "UNLOCKERROR") {
				return errors.New("unlock error")
			}
			return nil
		},
		IsExpiredFunc: func(in1 string) bool {
			if strings.Contains(in1, "EXPIRED") {
				return true
			}
			return false
		},
		PutFunc: func(in1 string, data []byte, ttl time.Time) error {
			if strings.Contains(in1, "PUTERROR") {
				return errors.New("put error")
			}
			return nil
		},
	}

	cacher := NewCacher(engine, 5, 5)

	expires := time.Now()
	data := []byte("hello")

	t.Run("engine locked", func(*testing.T) {
		err := cacher.Put("LOCKED", expires, data)
		if err != common.ErrEngineLocked {
			t.Errorf("expected error %s, got %s", common.ErrEngineLocked, err.Error())
		}
	})

	t.Run("lock error", func(*testing.T) {
		err := cacher.Put("LOCKERROR", expires, data)
		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("put error", func(*testing.T) {
		err := cacher.Put("PUTERROR", expires, data)
		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("unlock error", func(*testing.T) {
		err := cacher.Put("UNLOCKERROR", expires, data)
		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("put valid", func(*testing.T) {
		err := cacher.Put("anything else", expires, data)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
	})
}

func TestExpire(t *testing.T) {
	engine := &common.EngineMock{
		ExpireFunc: func(in1 string) error {
			if strings.Contains(in1, "EXPIREERROR") {
				return errors.New("error")
			}
			return nil
		},
	}

	cacher := NewCacher(engine, 5, 5)

	t.Run("expire error", func(*testing.T) {
		err := cacher.Expire("EXPIREERROR")
		if err == nil {
			t.Errorf("expected  error, got none")
		}
	})

	t.Run("expire ok", func(*testing.T) {
		err := cacher.Expire("anything else")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
	})
}
