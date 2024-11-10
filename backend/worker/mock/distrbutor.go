// Code generated by MockGen. DO NOT EDIT.
// Source: simple_bank/worker (interfaces: TaskDistributor)
//
// Generated by this command:
//
//	mockgen -package mockwk -destination worker/mock/distrbutor.go simple_bank/worker TaskDistributor
//

// Package mockwk is a generated GoMock package.
package mockwk

import (
	context "context"
	reflect "reflect"
	worker "simple_bank/worker"

	asynq "github.com/hibiken/asynq"
	gomock "go.uber.org/mock/gomock"
)

// MockTaskDistributor is a mock of TaskDistributor interface.
type MockTaskDistributor struct {
	ctrl     *gomock.Controller
	recorder *MockTaskDistributorMockRecorder
	isgomock struct{}
}

// MockTaskDistributorMockRecorder is the mock recorder for MockTaskDistributor.
type MockTaskDistributorMockRecorder struct {
	mock *MockTaskDistributor
}

// NewMockTaskDistributor creates a new mock instance.
func NewMockTaskDistributor(ctrl *gomock.Controller) *MockTaskDistributor {
	mock := &MockTaskDistributor{ctrl: ctrl}
	mock.recorder = &MockTaskDistributorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskDistributor) EXPECT() *MockTaskDistributorMockRecorder {
	return m.recorder
}

// DistributeTaskSendVerifyEmail mocks base method.
func (m *MockTaskDistributor) DistributeTaskSendVerifyEmail(ctx context.Context, payload *worker.PayloadSendVerifyEmail, opt ...asynq.Option) error {
	m.ctrl.T.Helper()
	varargs := []any{ctx, payload}
	for _, a := range opt {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DistributeTaskSendVerifyEmail", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DistributeTaskSendVerifyEmail indicates an expected call of DistributeTaskSendVerifyEmail.
func (mr *MockTaskDistributorMockRecorder) DistributeTaskSendVerifyEmail(ctx, payload any, opt ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, payload}, opt...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DistributeTaskSendVerifyEmail", reflect.TypeOf((*MockTaskDistributor)(nil).DistributeTaskSendVerifyEmail), varargs...)
}
