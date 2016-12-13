/*
* CODE GENERATED AUTOMATICALLY WITH github.com/ernesto-jimenez/gogen/automock
* THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package cacher

import (
	"fmt"
	mock "github.com/stretchr/testify/mock"

	common "github.com/fresh8/go-cache/engine/common"
	time "time"
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
func (m *CacherMock) Get(p0 string, p1 time.Time, p2 func() []byte) ([]byte, error) {

	ret := m.Called(p0, p1, p2)

	var r0 []byte
	switch res := ret.Get(0).(type) {
	case nil:
	case []byte:
		r0 = res
	default:
		panic(fmt.Sprintf("unexpected type: %v", res))
	}

	var r1 error
	switch res := ret.Get(1).(type) {
	case nil:
	case error:
		r1 = res
	default:
		panic(fmt.Sprintf("unexpected type: %v", res))
	}

	return r0, r1

}

// Setup mocked method
func (m *CacherMock) Setup(p0 common.Engine) {

	m.Called(p0)

}
