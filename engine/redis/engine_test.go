package redis

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/fresh8/go-cache/engine/common"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type mockPool struct {
	conn redigo.Conn
}

func (m mockPool) Get() redigo.Conn {
	return m.conn
}

func TestRedisEngine_Exists(t *testing.T) {
	fakeConn := redigomock.NewConn()
	engine := NewRedisStore("testing", &mockPool{
		conn: fakeConn,
	}, 1*time.Minute)

	cmd := fakeConn.Command("EXISTS", "testing:non-existing").Expect([]byte("false"))
	if engine.Exists("non-existing") {
		t.Fatal("key does not exist, marked as existing")
	}

	if fakeConn.Stats(cmd) != 1 {
		t.Fatal("exists command was not used")
	}

	cmd = fakeConn.Command("EXISTS", "testing:existing").Expect([]byte("true"))
	if !engine.Exists("existing") {
		t.Fatal("key exist, marked as non-existent")
	}

	if fakeConn.Stats(cmd) != 1 {
		t.Fatal("exists command was not used")
	}

	cmd = fakeConn.Command("EXISTS", "testing:existing").ExpectError(fmt.Errorf("Random error!"))
	if engine.Exists("non-existing") {
		t.Fatal("error for existing should return false")
	}

	if fakeConn.Stats(cmd) != 1 {
		t.Fatal("exists command was not used")
	}
}

func TestRedisEngine_Get(t *testing.T) {
	fakeConn := redigomock.NewConn()
	engine := NewRedisStore("testing", &mockPool{
		conn: fakeConn,
	}, 1*time.Minute)

	fakeConn.Command("GET", "testing:non-existing").ExpectError(redigo.ErrNil)

	data, err := engine.Get("non-existing")
	if err != common.ErrNonExistentKey {
		t.Fatalf("non-existing key error expected, %s given", err)
	}

	if data != nil {
		t.Fatal("data returned for non-existing key")
	}

	content := []byte("hello")
	fakeConn.Command("GET", "testing:existing").Expect(content)

	data, err = engine.Get("existing")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("%s expected, %s given", content, data)
	}
}

func TestRedisEngine_Put(t *testing.T) {
	fakeConn := redigomock.NewConn()
	cleanupTimeout := 1 * time.Minute
	engine := NewRedisStore("testing", &mockPool{
		conn: fakeConn,
	}, cleanupTimeout)

	content := []byte("hello")
	expires := time.Now().Add(1 * time.Hour)

	expectedErr := fmt.Errorf("Random error!")
	cmd1 := fakeConn.Command("SETEX", "testing:new-key", cleanupTimeout.Seconds(), content)
	cmd2 := fakeConn.Command("SET", "testing:expire:new-key", expires.Unix())
	cmd3 := fakeConn.Command("EXEC").Expect([]interface{}{"OK", "OK"}).ExpectError(expectedErr)

	err := engine.Put("new-key", []byte("hello"), expires)
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if fakeConn.Stats(cmd1) != 1 {
		t.Fatal("setex command was not used")
	}

	if fakeConn.Stats(cmd2) != 1 {
		t.Fatal("set command was not used")
	}

	if fakeConn.Stats(cmd3) != 1 {
		t.Fatal("exec command was not used")
	}

	err = engine.Put("new-key", []byte("hello"), expires)
	if err != expectedErr {
		t.Fatalf("random error expected, %s given", err)
	}
}

func TestRedisEngine_IsExpired(t *testing.T) {
	fakeConn := redigomock.NewConn()
	engine := NewRedisStore("testing", &mockPool{
		conn: fakeConn,
	}, 1*time.Minute)

	cmd1 := fakeConn.Command("EXISTS", "testing:expire:non-existing").Expect([]byte("false"))
	if engine.IsExpired("non-existing") {
		t.Fatal("key does not exist, marked as existing")
	}

	if fakeConn.Stats(cmd1) != 1 {
		t.Fatal("exists command was not used")
	}

	fakeConn.Clear()

	cmd2 := fakeConn.Command("EXISTS", "testing:expire:existing").Expect([]byte("true"))
	cmd3 := fakeConn.Command("GET", "testing:expire:existing").Expect(time.Now().Add(1 * time.Minute).Unix())
	if engine.IsExpired("existing") {
		t.Fatal("key exist, marked as non-existent")
	}

	if fakeConn.Stats(cmd2) != 1 {
		t.Fatal("exists command was not used")
	}

	if fakeConn.Stats(cmd3) != 1 {
		t.Fatal("get command was not used")
	}

	fakeConn.Clear()

	cmd4 := fakeConn.Command("EXISTS", "testing:expire:existing").Expect([]byte("true"))
	cmd5 := fakeConn.Command("GET", "testing:expire:existing").Expect(time.Now().Add(-1 * time.Minute).Unix())
	if !engine.IsExpired("existing") {
		t.Fatal("key exist, marked as non-existent")
	}

	if fakeConn.Stats(cmd4) != 1 {
		t.Fatal("exists command was not used")
	}

	if fakeConn.Stats(cmd5) != 1 {
		t.Fatal("get command was not used")
	}

	fakeConn.Clear()

	fakeConn.Command("EXISTS", "testing:expire:existing").ExpectError(fmt.Errorf("Random error!"))
	if engine.IsExpired("existing") {
		t.Fatal("error for existing should return false")
	}
}

func TestRedisEngine_Expire(t *testing.T) {
	fakeConn := redigomock.NewConn()
	engine := NewRedisStore("testing", &mockPool{
		conn: fakeConn,
	}, 1*time.Minute)

	expectedErr := fmt.Errorf("Random error!")
	cmd1 := fakeConn.Command("DEL", "testing:remove-key")
	cmd2 := fakeConn.Command("DEL", "testing:expire:remove-key")
	cmd3 := fakeConn.Command("DEL", "testing:lock:remove-key")
	cmd4 := fakeConn.Command("EXEC").Expect([]interface{}{"OK", "OK", "OK"}).ExpectError(expectedErr)

	err := engine.Expire("remove-key")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if fakeConn.Stats(cmd1) != 1 {
		t.Fatal("del key command was not used")
	}

	if fakeConn.Stats(cmd2) != 1 {
		t.Fatal("del expire command was not used")
	}

	if fakeConn.Stats(cmd3) != 1 {
		t.Fatal("del lock command was not used")
	}

	if fakeConn.Stats(cmd4) != 1 {
		t.Fatal("exec command was not used")
	}

	err = engine.Expire("remove-key")
	if err != expectedErr {
		t.Fatalf("random error expected, %s given", err)
	}
}

func TestRedisEngine_IsLocked(t *testing.T) {
	fakeConn := redigomock.NewConn()
	engine := NewRedisStore("testing", &mockPool{
		conn: fakeConn,
	}, 1*time.Minute)

	cmd := fakeConn.Command("EXISTS", "testing:lock:non-existing").Expect([]byte("false"))
	if engine.IsLocked("non-existing") {
		t.Fatal("key does not exist, marked as existing")
	}

	if fakeConn.Stats(cmd) != 1 {
		t.Fatal("exists command was not used")
	}

	cmd = fakeConn.Command("EXISTS", "testing:lock:existing").Expect([]byte("true"))
	if !engine.IsLocked("existing") {
		t.Fatal("key exist, marked as non-existent")
	}

	if fakeConn.Stats(cmd) != 1 {
		t.Fatal("exists command was not used")
	}

	cmd = fakeConn.Command("EXISTS", "testing:lock:existing").ExpectError(fmt.Errorf("Random error!"))
	if engine.IsLocked("non-existing") {
		t.Fatal("error for existing should return false")
	}

	if fakeConn.Stats(cmd) != 1 {
		t.Fatal("exists command was not used")
	}
}

func TestRedisEngine_Lock(t *testing.T) {
	fakeConn := redigomock.NewConn()
	engine := NewRedisStore("testing", &mockPool{
		conn: fakeConn,
	}, 1*time.Minute)

	expectedErr := fmt.Errorf("Random error!")
	cmd1 := fakeConn.Command("SET", "testing:lock:lock-key", []byte("1")).Expect("OK").ExpectError(expectedErr)

	err := engine.Lock("lock-key")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if fakeConn.Stats(cmd1) != 1 {
		t.Fatal("set command was not used")
	}

	err = engine.Lock("lock-key")
	if err != expectedErr {
		t.Fatalf("random error expected, %s given", err)
	}
}

func TestRedisEngine_Unlock(t *testing.T) {
	fakeConn := redigomock.NewConn()
	engine := NewRedisStore("testing", &mockPool{
		conn: fakeConn,
	}, 1*time.Minute)

	expectedErr := fmt.Errorf("Random error!")
	cmd1 := fakeConn.Command("DEL", "testing:lock:del-key").Expect("OK").ExpectError(expectedErr)

	err := engine.Unlock("del-key")
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if fakeConn.Stats(cmd1) != 1 {
		t.Fatal("del command was not used")
	}

	err = engine.Unlock("del-key")
	if err != expectedErr {
		t.Fatalf("random error expected, %s given", err)
	}
}
