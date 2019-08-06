// Code generated by MockGen. DO NOT EDIT.
// Source: lib/cdk/cdk.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockClienter is a mock of Clienter interface
type MockClienter struct {
	ctrl     *gomock.Controller
	recorder *MockClienterMockRecorder
}

// MockClienterMockRecorder is the mock recorder for MockClienter
type MockClienterMockRecorder struct {
	mock *MockClienter
}

// NewMockClienter creates a new mock instance
func NewMockClienter(ctrl *gomock.Controller) *MockClienter {
	mock := &MockClienter{ctrl: ctrl}
	mock.recorder = &MockClienterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClienter) EXPECT() *MockClienterMockRecorder {
	return m.recorder
}

// Setup mocks base method
func (m *MockClienter) Setup(repoPath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Setup", repoPath)
	ret0, _ := ret[0].(error)
	return ret0
}

// Setup indicates an expected call of Setup
func (mr *MockClienterMockRecorder) Setup(repoPath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Setup", reflect.TypeOf((*MockClienter)(nil).Setup), repoPath)
}

// List mocks base method
func (m *MockClienter) List(repoPath string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", repoPath)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockClienterMockRecorder) List(repoPath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockClienter)(nil).List), repoPath)
}

// Diff mocks base method
func (m *MockClienter) Diff(repoPath string) (string, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Diff", repoPath)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Diff indicates an expected call of Diff
func (mr *MockClienterMockRecorder) Diff(repoPath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Diff", reflect.TypeOf((*MockClienter)(nil).Diff), repoPath)
}

// Deploy mocks base method
func (m *MockClienter) Deploy(repoPath, stacks string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Deploy", repoPath, stacks)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Deploy indicates an expected call of Deploy
func (mr *MockClienterMockRecorder) Deploy(repoPath, stacks interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Deploy", reflect.TypeOf((*MockClienter)(nil).Deploy), repoPath, stacks)
}
