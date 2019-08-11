// Code generated by MockGen. DO NOT EDIT.
// Source: functions/github/client/client.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	client "github.com/sambaiz/cdkbot/functions/github/client"
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

// CreateStatusOfLatestCommit mocks base method
func (m *MockClienter) CreateStatusOfLatestCommit(ctx context.Context, owner, repo string, number int, state client.State) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateStatusOfLatestCommit", ctx, owner, repo, number, state)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateStatusOfLatestCommit indicates an expected call of CreateStatusOfLatestCommit
func (mr *MockClienterMockRecorder) CreateStatusOfLatestCommit(ctx, owner, repo, number, state interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateStatusOfLatestCommit", reflect.TypeOf((*MockClienter)(nil).CreateStatusOfLatestCommit), ctx, owner, repo, number, state)
}

// CreateComment mocks base method
func (m *MockClienter) CreateComment(ctx context.Context, owner, repo string, number int, body string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateComment", ctx, owner, repo, number, body)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateComment indicates an expected call of CreateComment
func (mr *MockClienterMockRecorder) CreateComment(ctx, owner, repo, number, body interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateComment", reflect.TypeOf((*MockClienter)(nil).CreateComment), ctx, owner, repo, number, body)
}

// GetPullRequestLatestCommitHash mocks base method
func (m *MockClienter) GetPullRequestLatestCommitHash(ctx context.Context, owner, repo string, number int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPullRequestLatestCommitHash", ctx, owner, repo, number)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPullRequestLatestCommitHash indicates an expected call of GetPullRequestLatestCommitHash
func (mr *MockClienterMockRecorder) GetPullRequestLatestCommitHash(ctx, owner, repo, number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPullRequestLatestCommitHash", reflect.TypeOf((*MockClienter)(nil).GetPullRequestLatestCommitHash), ctx, owner, repo, number)
}

// GetPullRequestBaseBranch mocks base method
func (m *MockClienter) GetPullRequestBaseBranch(ctx context.Context, owner, repo string, number int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPullRequestBaseBranch", ctx, owner, repo, number)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPullRequestBaseBranch indicates an expected call of GetPullRequestBaseBranch
func (mr *MockClienterMockRecorder) GetPullRequestBaseBranch(ctx, owner, repo, number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPullRequestBaseBranch", reflect.TypeOf((*MockClienter)(nil).GetPullRequestBaseBranch), ctx, owner, repo, number)
}
