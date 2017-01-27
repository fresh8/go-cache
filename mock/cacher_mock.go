/*
* CODE GENERATED AUTOMATICALLY WITH github.com/ernesto-jimenez/gogen/automock
* THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package mock

import (
	"fmt"
	"time"

	mock "github.com/stretchr/testify/mock"
)

// CacherMock mock
type CacherMock struct {
	mock.Mock
}

// Expire mocked method
func (m *CacherMock) Expire(p0 string) error {

	ret := m.Called(p0)

	var r0 error
	switch res := ret.Get(0).(type) {
	case nil:
	case error:
		r0 = res
	default:
		panic(fmt.Sprintf("unexpected type: %v", res))
	}

	return r0

}

// Get mocked method
func (m *CacherMock) Get(p0 string, p1 time.Time, p2 func() ([]byte, error)) func() ([]byte, error) {

	ret := m.Called(p0, p1, p2)

	var r0 func() ([]byte, error)
	switch res := ret.Get(0).(type) {
	case nil:
	case func() ([]byte, error):
		r0 = res
	default:
		panic(fmt.Sprintf("unexpected type: %v", res))
	}

	return r0

}
