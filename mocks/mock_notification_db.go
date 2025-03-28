// Code generated by MockGen. DO NOT EDIT.
// Source: ../database/notification.go
//
// Generated by this command:
//
//	mockgen -source=../database/notification.go -destination=../mocks/mock_notification_db.go -package mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	database "github.com/SlotifyApp/slotify-backend/database"
	gomock "go.uber.org/mock/gomock"
)

// MockNotificationDatabase is a mock of NotificationDatabase interface.
type MockNotificationDatabase struct {
	ctrl     *gomock.Controller
	recorder *MockNotificationDatabaseMockRecorder
	isgomock struct{}
}

// MockNotificationDatabaseMockRecorder is the mock recorder for MockNotificationDatabase.
type MockNotificationDatabaseMockRecorder struct {
	mock *MockNotificationDatabase
}

// NewMockNotificationDatabase creates a new mock instance.
func NewMockNotificationDatabase(ctrl *gomock.Controller) *MockNotificationDatabase {
	mock := &MockNotificationDatabase{ctrl: ctrl}
	mock.recorder = &MockNotificationDatabaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNotificationDatabase) EXPECT() *MockNotificationDatabaseMockRecorder {
	return m.recorder
}

// CreateNotification mocks base method.
func (m *MockNotificationDatabase) CreateNotification(ctx context.Context, arg database.CreateNotificationParams) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateNotification", ctx, arg)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateNotification indicates an expected call of CreateNotification.
func (mr *MockNotificationDatabaseMockRecorder) CreateNotification(ctx, arg any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNotification", reflect.TypeOf((*MockNotificationDatabase)(nil).CreateNotification), ctx, arg)
}

// CreateUserNotification mocks base method.
func (m *MockNotificationDatabase) CreateUserNotification(ctx context.Context, arg database.CreateUserNotificationParams) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUserNotification", ctx, arg)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUserNotification indicates an expected call of CreateUserNotification.
func (mr *MockNotificationDatabaseMockRecorder) CreateUserNotification(ctx, arg any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserNotification", reflect.TypeOf((*MockNotificationDatabase)(nil).CreateUserNotification), ctx, arg)
}
