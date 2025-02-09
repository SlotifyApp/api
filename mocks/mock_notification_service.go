// Code generated by MockGen. DO NOT EDIT.
// Source: ../notification/notification.go
//
// Generated by this command:
//
//	mockgen -source=../notification/notification.go -destination=../mocks/mock_notification_service.go -package mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	http "net/http"
	reflect "reflect"

	database "github.com/SlotifyApp/slotify-backend/database"
	logger "github.com/SlotifyApp/slotify-backend/logger"
	gomock "go.uber.org/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
	isgomock struct{}
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// DeleteUserConn mocks base method.
func (m *MockService) DeleteUserConn(l *logger.Logger, userID uint32, w http.ResponseWriter) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteUserConn", l, userID, w)
}

// DeleteUserConn indicates an expected call of DeleteUserConn.
func (mr *MockServiceMockRecorder) DeleteUserConn(l, userID, w any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUserConn", reflect.TypeOf((*MockService)(nil).DeleteUserConn), l, userID, w)
}

// RegisterUserClient mocks base method.
func (m *MockService) RegisterUserClient(l *logger.Logger, userID uint32, w http.ResponseWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterUserClient", l, userID, w)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterUserClient indicates an expected call of RegisterUserClient.
func (mr *MockServiceMockRecorder) RegisterUserClient(l, userID, w any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterUserClient", reflect.TypeOf((*MockService)(nil).RegisterUserClient), l, userID, w)
}

// SendNotification mocks base method.
func (m *MockService) SendNotification(ctx context.Context, l *logger.Logger, db database.NotificationDatabase, userIDs []uint32, notif database.CreateNotificationParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendNotification", ctx, l, db, userIDs, notif)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendNotification indicates an expected call of SendNotification.
func (mr *MockServiceMockRecorder) SendNotification(ctx, l, db, userIDs, notif any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendNotification", reflect.TypeOf((*MockService)(nil).SendNotification), ctx, l, db, userIDs, notif)
}
