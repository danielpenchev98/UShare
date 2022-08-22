// Code generated by MockGen. DO NOT EDIT.
// Source: auth.go

// Package auth_mocks is a generated GoMock package.
package auth_mocks

import (
	auth "github.com/danielpenchev98/UShare/web-server/internal/auth"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockJwtCreator is a mock of JwtCreator interface
type MockJwtCreator struct {
	ctrl     *gomock.Controller
	recorder *MockJwtCreatorMockRecorder
}

// MockJwtCreatorMockRecorder is the mock recorder for MockJwtCreator
type MockJwtCreatorMockRecorder struct {
	mock *MockJwtCreator
}

// NewMockJwtCreator creates a new mock instance
func NewMockJwtCreator(ctrl *gomock.Controller) *MockJwtCreator {
	mock := &MockJwtCreator{ctrl: ctrl}
	mock.recorder = &MockJwtCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockJwtCreator) EXPECT() *MockJwtCreatorMockRecorder {
	return m.recorder
}

// GenerateToken mocks base method
func (m *MockJwtCreator) GenerateToken(arg0 uint) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateToken", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateToken indicates an expected call of GenerateToken
func (mr *MockJwtCreatorMockRecorder) GenerateToken(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateToken", reflect.TypeOf((*MockJwtCreator)(nil).GenerateToken), arg0)
}

// ValidateToken mocks base method
func (m *MockJwtCreator) ValidateToken(arg0 string) (*auth.JwtClaim, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateToken", arg0)
	ret0, _ := ret[0].(*auth.JwtClaim)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateToken indicates an expected call of ValidateToken
func (mr *MockJwtCreatorMockRecorder) ValidateToken(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateToken", reflect.TypeOf((*MockJwtCreator)(nil).ValidateToken), arg0)
}
