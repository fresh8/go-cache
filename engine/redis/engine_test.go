package redis

import (
	"testing"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"time"
	"fmt"
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

	fakeConn.Command("EXISTS", "testing:non-existing").Expect([]byte("false"))

	if engine.Exists("non-existing") {
		t.Fatal("key does not exist, marked as existing")
	}

	fakeConn.Command("EXISTS", "testing:existing").Expect([]byte("true"))

	if !engine.Exists("existing") {
		t.Fatal("key exist, marked as non-existent")
	}

	fakeConn.Command("EXISTS", "testing:existing").ExpectError(fmt.Errorf("Random error!"))
	if engine.Exists("non-existing") {
		t.Fatal("error for existing should return false")
	}
}
