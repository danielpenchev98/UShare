// Code generated by MockGen. DO NOT EDIT.
// Source: fm_dao.go

// Package dao_mocks is a generated GoMock package.
package dao_mocks

import (
	models "github.com/danielpenchev98/UShare/web-server/internal/db/models"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockFmDAO is a mock of FmDAO interface
type MockFmDAO struct {
	ctrl     *gomock.Controller
	recorder *MockFmDAOMockRecorder
}

// MockFmDAOMockRecorder is the mock recorder for MockFmDAO
type MockFmDAOMockRecorder struct {
	mock *MockFmDAO
}

// NewMockFmDAO creates a new mock instance
func NewMockFmDAO(ctrl *gomock.Controller) *MockFmDAO {
	mock := &MockFmDAO{ctrl: ctrl}
	mock.recorder = &MockFmDAOMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFmDAO) EXPECT() *MockFmDAOMockRecorder {
	return m.recorder
}

// AddFileInfo mocks base method
func (m *MockFmDAO) AddFileInfo(userID uint, fileName, groupName string) (uint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddFileInfo", userID, fileName, groupName)
	ret0, _ := ret[0].(uint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddFileInfo indicates an expected call of AddFileInfo
func (mr *MockFmDAOMockRecorder) AddFileInfo(userID, fileName, groupName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddFileInfo", reflect.TypeOf((*MockFmDAO)(nil).AddFileInfo), userID, fileName, groupName)
}

// GetFileInfo mocks base method
func (m *MockFmDAO) GetFileInfo(userID, fileID uint, groupName string) (models.FileInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFileInfo", userID, fileID, groupName)
	ret0, _ := ret[0].(models.FileInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFileInfo indicates an expected call of GetFileInfo
func (mr *MockFmDAOMockRecorder) GetFileInfo(userID, fileID, groupName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileInfo", reflect.TypeOf((*MockFmDAO)(nil).GetFileInfo), userID, fileID, groupName)
}

// GetAllFilesInfo mocks base method
func (m *MockFmDAO) GetAllFilesInfo(userID uint, groupName string) ([]models.FileInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllFilesInfo", userID, groupName)
	ret0, _ := ret[0].([]models.FileInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllFilesInfo indicates an expected call of GetAllFilesInfo
func (mr *MockFmDAOMockRecorder) GetAllFilesInfo(userID, groupName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllFilesInfo", reflect.TypeOf((*MockFmDAO)(nil).GetAllFilesInfo), userID, groupName)
}

// RemoveFileInfo mocks base method
func (m *MockFmDAO) RemoveFileInfo(userID, fileID uint, groupName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveFileInfo", userID, fileID, groupName)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveFileInfo indicates an expected call of RemoveFileInfo
func (mr *MockFmDAOMockRecorder) RemoveFileInfo(userID, fileID, groupName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveFileInfo", reflect.TypeOf((*MockFmDAO)(nil).RemoveFileInfo), userID, fileID, groupName)
}

// Migrate mocks base method
func (m *MockFmDAO) Migrate() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Migrate")
	ret0, _ := ret[0].(error)
	return ret0
}

// Migrate indicates an expected call of Migrate
func (mr *MockFmDAOMockRecorder) Migrate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Migrate", reflect.TypeOf((*MockFmDAO)(nil).Migrate))
}