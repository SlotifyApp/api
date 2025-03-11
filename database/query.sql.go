// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package database

import (
	"context"
	"database/sql"
	"time"
)

const addSlotifyGroup = `-- name: AddSlotifyGroup :execlastid
INSERT INTO SlotifyGroup (name) VALUES (?)
`

func (q *Queries) AddSlotifyGroup(ctx context.Context, name string) (int64, error) {
	result, err := q.exec(ctx, q.addSlotifyGroupStmt, addSlotifyGroup, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const addUserToSlotifyGroup = `-- name: AddUserToSlotifyGroup :execrows
INSERT INTO UserToSlotifyGroup (user_id, slotify_group_id) VALUES (?, ?)
`

type AddUserToSlotifyGroupParams struct {
	UserID         uint32 `json:"userId"`
	SlotifyGroupID uint32 `json:"slotifyGroupId"`
}

func (q *Queries) AddUserToSlotifyGroup(ctx context.Context, arg AddUserToSlotifyGroupParams) (int64, error) {
	result, err := q.exec(ctx, q.addUserToSlotifyGroupStmt, addUserToSlotifyGroup, arg.UserID, arg.SlotifyGroupID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const batchDeleteWeekOldInvites = `-- name: BatchDeleteWeekOldInvites :execrows
DELETE FROM Invite
WHERE created_at <= CURDATE() - INTERVAL 1 WEEK
  AND id >= (SELECT MIN(id) FROM Invite WHERE DATE(created_at) <= CURDATE() - INTERVAL 1 WEEK)
ORDER BY id
LIMIT ?
`

func (q *Queries) BatchDeleteWeekOldInvites(ctx context.Context, limit int32) (int64, error) {
	result, err := q.exec(ctx, q.batchDeleteWeekOldInvitesStmt, batchDeleteWeekOldInvites, limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const batchDeleteWeekOldNotifications = `-- name: BatchDeleteWeekOldNotifications :execrows
DELETE FROM Notification
WHERE created <= CURDATE() - INTERVAL 1 WEEK
  AND id >= (SELECT MIN(id) FROM Notification WHERE created <= CURDATE() - INTERVAL 1 WEEK)
ORDER BY id
LIMIT ?
`

func (q *Queries) BatchDeleteWeekOldNotifications(ctx context.Context, limit int32) (int64, error) {
	result, err := q.exec(ctx, q.batchDeleteWeekOldNotificationsStmt, batchDeleteWeekOldNotifications, limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const batchExpireInvites = `-- name: BatchExpireInvites :execrows
UPDATE Invite SET status = 'expired'
WHERE expiry_date <= CURDATE()
  AND status != 'expired'
  AND id >= (SELECT MIN(id) FROM Invite WHERE expiry_date <= CURDATE() AND status != 'expired')
ORDER BY id
LIMIT ?
`

func (q *Queries) BatchExpireInvites(ctx context.Context, limit int32) (int64, error) {
	result, err := q.exec(ctx, q.batchExpireInvitesStmt, batchExpireInvites, limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const checkMemberInSlotifyGroup = `-- name: CheckMemberInSlotifyGroup :one
SELECT COUNT(*) FROM UserToSlotifyGroup
WHERE user_id=? AND slotify_group_id=?
`

type CheckMemberInSlotifyGroupParams struct {
	UserID         uint32 `json:"userId"`
	SlotifyGroupID uint32 `json:"slotifyGroupId"`
}

func (q *Queries) CheckMemberInSlotifyGroup(ctx context.Context, arg CheckMemberInSlotifyGroupParams) (int64, error) {
	row := q.queryRow(ctx, q.checkMemberInSlotifyGroupStmt, checkMemberInSlotifyGroup, arg.UserID, arg.SlotifyGroupID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countExpiredInvites = `-- name: CountExpiredInvites :one
SELECT COUNT(*) FROM Invite
WHERE status='expired'
`

func (q *Queries) CountExpiredInvites(ctx context.Context) (int64, error) {
	row := q.queryRow(ctx, q.countExpiredInvitesStmt, countExpiredInvites)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countSlotifyGroupByID = `-- name: CountSlotifyGroupByID :one
SELECT COUNT(*) FROM SlotifyGroup WHERE id=?
`

func (q *Queries) CountSlotifyGroupByID(ctx context.Context, id uint32) (int64, error) {
	row := q.queryRow(ctx, q.countSlotifyGroupByIDStmt, countSlotifyGroupByID, id)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countSlotifyGroupMembers = `-- name: CountSlotifyGroupMembers :one
SELECT COUNT(*) FROM SlotifyGroup sg
JOIN UserToSlotifyGroup utsg ON sg.id=utsg.slotify_group_id
JOIN User u ON u.id=utsg.user_id 
WHERE sg.id=?
`

func (q *Queries) CountSlotifyGroupMembers(ctx context.Context, id uint32) (int64, error) {
	row := q.queryRow(ctx, q.countSlotifyGroupMembersStmt, countSlotifyGroupMembers, id)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countUserByEmail = `-- name: CountUserByEmail :one
SELECT COUNT(*) FROM User WHERE email=?
`

func (q *Queries) CountUserByEmail(ctx context.Context, email string) (int64, error) {
	row := q.queryRow(ctx, q.countUserByEmailStmt, countUserByEmail, email)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countUserByID = `-- name: CountUserByID :one
SELECT COUNT(*) FROM User WHERE id=?
`

func (q *Queries) CountUserByID(ctx context.Context, id uint32) (int64, error) {
	row := q.queryRow(ctx, q.countUserByIDStmt, countUserByID, id)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countWeekOldInvites = `-- name: CountWeekOldInvites :one
SELECT COUNT(*) FROM Invite
WHERE DATE(created_at) <= CURDATE() - INTERVAL 1 WEEK
`

func (q *Queries) CountWeekOldInvites(ctx context.Context) (int64, error) {
	row := q.queryRow(ctx, q.countWeekOldInvitesStmt, countWeekOldInvites)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countWeekOldNotifications = `-- name: CountWeekOldNotifications :one
SELECT COUNT(*) FROM Notification
WHERE DATE(created) <= CURDATE() - INTERVAL 1 WEEK
`

func (q *Queries) CountWeekOldNotifications(ctx context.Context) (int64, error) {
	row := q.queryRow(ctx, q.countWeekOldNotificationsStmt, countWeekOldNotifications)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createInvite = `-- name: CreateInvite :execlastid
INSERT INTO Invite (slotify_group_id, from_user_id, to_user_id, message, status, expiry_date, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
`

type CreateInviteParams struct {
	SlotifyGroupID uint32       `json:"slotifyGroupId"`
	FromUserID     uint32       `json:"fromUserId"`
	ToUserID       uint32       `json:"toUserId"`
	Message        string       `json:"message"`
	Status         InviteStatus `json:"status"`
	ExpiryDate     time.Time    `json:"expiryDate"`
	CreatedAt      time.Time    `json:"createdAt"`
}

func (q *Queries) CreateInvite(ctx context.Context, arg CreateInviteParams) (int64, error) {
	result, err := q.exec(ctx, q.createInviteStmt, createInvite,
		arg.SlotifyGroupID,
		arg.FromUserID,
		arg.ToUserID,
		arg.Message,
		arg.Status,
		arg.ExpiryDate,
		arg.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createMeeting = `-- name: CreateMeeting :execlastid
INSERT INTO Meeting (meeting_pref_id, owner_id, msft_meeting_id) VALUES (?,?,?)
`

type CreateMeetingParams struct {
	MeetingPrefID uint32 `json:"meetingPrefId"`
	OwnerID       uint32 `json:"ownerId"`
	MsftMeetingID string `json:"msftMeetingId"`
}

func (q *Queries) CreateMeeting(ctx context.Context, arg CreateMeetingParams) (int64, error) {
	result, err := q.exec(ctx, q.createMeetingStmt, createMeeting, arg.MeetingPrefID, arg.OwnerID, arg.MsftMeetingID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createMeetingPreferences = `-- name: CreateMeetingPreferences :execlastid
INSERT INTO MeetingPreferences (meeting_start_time, start_date_range, end_date_range) VALUES (?,?,?)
`

type CreateMeetingPreferencesParams struct {
	MeetingStartTime time.Time `json:"meetingStartTime"`
	StartDateRange   time.Time `json:"startDateRange"`
	EndDateRange     time.Time `json:"endDateRange"`
}

func (q *Queries) CreateMeetingPreferences(ctx context.Context, arg CreateMeetingPreferencesParams) (int64, error) {
	result, err := q.exec(ctx, q.createMeetingPreferencesStmt, createMeetingPreferences, arg.MeetingStartTime, arg.StartDateRange, arg.EndDateRange)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createNotification = `-- name: CreateNotification :execlastid
INSERT INTO Notification (message, created) VALUES(?, ?)
`

type CreateNotificationParams struct {
	Message string    `json:"message"`
	Created time.Time `json:"created"`
}

func (q *Queries) CreateNotification(ctx context.Context, arg CreateNotificationParams) (int64, error) {
	result, err := q.exec(ctx, q.createNotificationStmt, createNotification, arg.Message, arg.Created)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createPlaceholderMeeting = `-- name: CreatePlaceholderMeeting :execlastid
INSERT INTO PlaceholderMeeting (request_id, title, start_time, end_time, location, duration, start_date_range, end_date_range) VALUES (?,?,?,?,?,?,?,?)
`

type CreatePlaceholderMeetingParams struct {
	RequestID      uint32    `json:"requestId"`
	Title          string    `json:"title"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	Location       string    `json:"location"`
	Duration       time.Time `json:"duration"`
	StartDateRange time.Time `json:"startDateRange"`
	EndDateRange   time.Time `json:"endDateRange"`
}

func (q *Queries) CreatePlaceholderMeeting(ctx context.Context, arg CreatePlaceholderMeetingParams) (int64, error) {
	result, err := q.exec(ctx, q.createPlaceholderMeetingStmt, createPlaceholderMeeting,
		arg.RequestID,
		arg.Title,
		arg.StartTime,
		arg.EndTime,
		arg.Location,
		arg.Duration,
		arg.StartDateRange,
		arg.EndDateRange,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createPlaceholderMeetingAttendee = `-- name: CreatePlaceholderMeetingAttendee :execlastid
INSERT INTO PlaceholderMeetingAttendee (meeting_id, user_id) VALUES (?,?)
`

type CreatePlaceholderMeetingAttendeeParams struct {
	MeetingID uint32 `json:"meetingId"`
	UserID    uint32 `json:"userId"`
}

func (q *Queries) CreatePlaceholderMeetingAttendee(ctx context.Context, arg CreatePlaceholderMeetingAttendeeParams) (int64, error) {
	result, err := q.exec(ctx, q.createPlaceholderMeetingAttendeeStmt, createPlaceholderMeetingAttendee, arg.MeetingID, arg.UserID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createRefreshToken = `-- name: CreateRefreshToken :execrows
REPLACE INTO RefreshToken (user_id, token) VALUES (?, ?)
`

type CreateRefreshTokenParams struct {
	UserID uint32 `json:"userId"`
	Token  string `json:"token"`
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (int64, error) {
	result, err := q.exec(ctx, q.createRefreshTokenStmt, createRefreshToken, arg.UserID, arg.Token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const createRequestToMeeting = `-- name: CreateRequestToMeeting :execlastid
INSERT INTO RequestToMeeting (request_id, meeting_id) VALUES (?,?)
`

type CreateRequestToMeetingParams struct {
	RequestID uint32 `json:"requestId"`
	MeetingID uint32 `json:"meetingId"`
}

func (q *Queries) CreateRequestToMeeting(ctx context.Context, arg CreateRequestToMeetingParams) (int64, error) {
	result, err := q.exec(ctx, q.createRequestToMeetingStmt, createRequestToMeeting, arg.RequestID, arg.MeetingID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createReschedulingRequest = `-- name: CreateReschedulingRequest :execlastid
INSERT INTO ReschedulingRequest (requested_by) VALUES (?)
`

func (q *Queries) CreateReschedulingRequest(ctx context.Context, requestedBy uint32) (int64, error) {
	result, err := q.exec(ctx, q.createReschedulingRequestStmt, createReschedulingRequest, requestedBy)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createUser = `-- name: CreateUser :execlastid
INSERT INTO User (email, first_name, last_name) VALUES (?, ?, ?)
`

type CreateUserParams struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int64, error) {
	result, err := q.exec(ctx, q.createUserStmt, createUser, arg.Email, arg.FirstName, arg.LastName)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createUserNotification = `-- name: CreateUserNotification :execrows
INSERT INTO UserToNotification (user_id, notification_id, is_read) VALUES(?, ?, FALSE)
`

type CreateUserNotificationParams struct {
	UserID         uint32 `json:"userId"`
	NotificationID uint32 `json:"notificationId"`
}

func (q *Queries) CreateUserNotification(ctx context.Context, arg CreateUserNotificationParams) (int64, error) {
	result, err := q.exec(ctx, q.createUserNotificationStmt, createUserNotification, arg.UserID, arg.NotificationID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteInviteByID = `-- name: DeleteInviteByID :execrows
DELETE FROM Invite WHERE id=?
`

func (q *Queries) DeleteInviteByID(ctx context.Context, id uint32) (int64, error) {
	result, err := q.exec(ctx, q.deleteInviteByIDStmt, deleteInviteByID, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteRefreshTokenByUserID = `-- name: DeleteRefreshTokenByUserID :execrows
DELETE FROM RefreshToken WHERE user_id=?
`

func (q *Queries) DeleteRefreshTokenByUserID(ctx context.Context, userID uint32) (int64, error) {
	result, err := q.exec(ctx, q.deleteRefreshTokenByUserIDStmt, deleteRefreshTokenByUserID, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteSlotifyGroupByID = `-- name: DeleteSlotifyGroupByID :execrows
DELETE FROM SlotifyGroup WHERE id=?
`

func (q *Queries) DeleteSlotifyGroupByID(ctx context.Context, id uint32) (int64, error) {
	result, err := q.exec(ctx, q.deleteSlotifyGroupByIDStmt, deleteSlotifyGroupByID, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteUserByID = `-- name: DeleteUserByID :execrows
DELETE FROM User WHERE id=?
`

func (q *Queries) DeleteUserByID(ctx context.Context, id uint32) (int64, error) {
	result, err := q.exec(ctx, q.deleteUserByIDStmt, deleteUserByID, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getAllRequestsForUser = `-- name: GetAllRequestsForUser :many
SELECT rr.request_id, rr.requested_by, rr.status, rr.created_at, m.msft_meeting_id, m.id, pm.meeting_id, pm.title, pm.start_time, pm.end_time, pm.duration, pm.location  FROM ReschedulingRequest rr JOIN RequestToMeeting rtm ON rr.request_id = rtm.request_id JOIN Meeting m ON rtm.meeting_id = m.id LEFT JOIN PlaceholderMeeting pm ON r.request_id = pm.request_id WHERE m.owner_id = ?
`

type GetAllRequestsForUserRow struct {
	RequestID     uint32                    `json:"requestId"`
	RequestedBy   uint32                    `json:"requestedBy"`
	Status        ReschedulingrequestStatus `json:"status"`
	CreatedAt     time.Time                 `json:"createdAt"`
	MsftMeetingID string                    `json:"msftMeetingId"`
	ID            uint32                    `json:"id"`
	MeetingID     sql.NullInt32             `json:"meetingId"`
	Title         sql.NullString            `json:"title"`
	StartTime     sql.NullTime              `json:"startTime"`
	EndTime       sql.NullTime              `json:"endTime"`
	Duration      sql.NullTime              `json:"duration"`
	Location      sql.NullString            `json:"location"`
}

func (q *Queries) GetAllRequestsForUser(ctx context.Context, ownerID uint32) ([]GetAllRequestsForUserRow, error) {
	rows, err := q.query(ctx, q.getAllRequestsForUserStmt, getAllRequestsForUser, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllRequestsForUserRow{}
	for rows.Next() {
		var i GetAllRequestsForUserRow
		if err := rows.Scan(
			&i.RequestID,
			&i.RequestedBy,
			&i.Status,
			&i.CreatedAt,
			&i.MsftMeetingID,
			&i.ID,
			&i.MeetingID,
			&i.Title,
			&i.StartTime,
			&i.EndTime,
			&i.Duration,
			&i.Location,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllSlotifyGroupMembers = `-- name: GetAllSlotifyGroupMembers :many
SELECT u.id, u.email, u.first_name, u.last_name FROM SlotifyGroup sg
JOIN UserToSlotifyGroup utsg ON sg.id=utsg.slotify_group_id
JOIN User u ON u.id=utsg.user_id 
WHERE sg.id=?
`

type GetAllSlotifyGroupMembersRow struct {
	ID        uint32 `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (q *Queries) GetAllSlotifyGroupMembers(ctx context.Context, id uint32) ([]GetAllSlotifyGroupMembersRow, error) {
	rows, err := q.query(ctx, q.getAllSlotifyGroupMembersStmt, getAllSlotifyGroupMembers, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllSlotifyGroupMembersRow{}
	for rows.Next() {
		var i GetAllSlotifyGroupMembersRow
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.FirstName,
			&i.LastName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllSlotifyGroupMembersExcept = `-- name: GetAllSlotifyGroupMembersExcept :many
SELECT u.id FROM SlotifyGroup sg
JOIN UserToSlotifyGroup utsg ON sg.id=utsg.slotify_group_id
JOIN User u ON u.id=utsg.user_id 
WHERE sg.id=? AND u.id!=?
`

type GetAllSlotifyGroupMembersExceptParams struct {
	SlotifyGroupID uint32 `json:"slotifyGroupID"`
	UserID         uint32 `json:"userID"`
}

func (q *Queries) GetAllSlotifyGroupMembersExcept(ctx context.Context, arg GetAllSlotifyGroupMembersExceptParams) ([]uint32, error) {
	rows, err := q.query(ctx, q.getAllSlotifyGroupMembersExceptStmt, getAllSlotifyGroupMembersExcept, arg.SlotifyGroupID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []uint32{}
	for rows.Next() {
		var id uint32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getInviteByID = `-- name: GetInviteByID :one
SELECT id, slotify_group_id, from_user_id, to_user_id, message, status, expiry_date, created_at FROM Invite
WHERE id=?
`

func (q *Queries) GetInviteByID(ctx context.Context, id uint32) (Invite, error) {
	row := q.queryRow(ctx, q.getInviteByIDStmt, getInviteByID, id)
	var i Invite
	err := row.Scan(
		&i.ID,
		&i.SlotifyGroupID,
		&i.FromUserID,
		&i.ToUserID,
		&i.Message,
		&i.Status,
		&i.ExpiryDate,
		&i.CreatedAt,
	)
	return i, err
}

const getMeetingByID = `-- name: GetMeetingByID :one
SELECT id, meeting_pref_id, owner_id, msft_meeting_id FROM Meeting
WHERE id=?
`

func (q *Queries) GetMeetingByID(ctx context.Context, id uint32) (Meeting, error) {
	row := q.queryRow(ctx, q.getMeetingByIDStmt, getMeetingByID, id)
	var i Meeting
	err := row.Scan(
		&i.ID,
		&i.MeetingPrefID,
		&i.OwnerID,
		&i.MsftMeetingID,
	)
	return i, err
}

const getMeetingByMSFTID = `-- name: GetMeetingByMSFTID :one
SELECT id, meeting_pref_id, owner_id, msft_meeting_id FROM Meeting
WHERE msft_meeting_id=?
`

func (q *Queries) GetMeetingByMSFTID(ctx context.Context, msftMeetingID string) (Meeting, error) {
	row := q.queryRow(ctx, q.getMeetingByMSFTIDStmt, getMeetingByMSFTID, msftMeetingID)
	var i Meeting
	err := row.Scan(
		&i.ID,
		&i.MeetingPrefID,
		&i.OwnerID,
		&i.MsftMeetingID,
	)
	return i, err
}

const getMeetingPreferences = `-- name: GetMeetingPreferences :one
SELECT id, meeting_start_time, start_date_range, end_date_range FROM MeetingPreferences
WHERE id=?
`

func (q *Queries) GetMeetingPreferences(ctx context.Context, id uint32) (Meetingpreferences, error) {
	row := q.queryRow(ctx, q.getMeetingPreferencesStmt, getMeetingPreferences, id)
	var i Meetingpreferences
	err := row.Scan(
		&i.ID,
		&i.MeetingStartTime,
		&i.StartDateRange,
		&i.EndDateRange,
	)
	return i, err
}

const getRefreshTokenByUserID = `-- name: GetRefreshTokenByUserID :one
SELECT id, user_id, token, revoked FROM RefreshToken WHERE user_id=?
`

func (q *Queries) GetRefreshTokenByUserID(ctx context.Context, userID uint32) (RefreshToken, error) {
	row := q.queryRow(ctx, q.getRefreshTokenByUserIDStmt, getRefreshTokenByUserID, userID)
	var i RefreshToken
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.Revoked,
	)
	return i, err
}

const getSlotifyGroupByID = `-- name: GetSlotifyGroupByID :one
SELECT id, name FROM SlotifyGroup WHERE id=?
`

func (q *Queries) GetSlotifyGroupByID(ctx context.Context, id uint32) (SlotifyGroup, error) {
	row := q.queryRow(ctx, q.getSlotifyGroupByIDStmt, getSlotifyGroupByID, id)
	var i SlotifyGroup
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getUnreadUserNotifications = `-- name: GetUnreadUserNotifications :many
SELECT n.id, n.message, n.created FROM UserToNotification utn
JOIN Notification n ON n.id=utn.notification_id 
WHERE utn.user_id=? AND utn.is_read=FALSE
ORDER BY n.created DESC
`

func (q *Queries) GetUnreadUserNotifications(ctx context.Context, userID uint32) ([]Notification, error) {
	rows, err := q.query(ctx, q.getUnreadUserNotificationsStmt, getUnreadUserNotifications, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Notification{}
	for rows.Next() {
		var i Notification
		if err := rows.Scan(&i.ID, &i.Message, &i.Created); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, first_name, last_name, msft_home_account_id FROM User WHERE email=?
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.queryRow(ctx, q.getUserByEmailStmt, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.FirstName,
		&i.LastName,
		&i.MsftHomeAccountID,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, email, first_name, last_name, msft_home_account_id FROM User WHERE id=?
`

func (q *Queries) GetUserByID(ctx context.Context, id uint32) (User, error) {
	row := q.queryRow(ctx, q.getUserByIDStmt, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.FirstName,
		&i.LastName,
		&i.MsftHomeAccountID,
	)
	return i, err
}

const getUsersSlotifyGroups = `-- name: GetUsersSlotifyGroups :many
SELECT sg.id, sg.name FROM UserToSlotifyGroup utsg
JOIN SlotifyGroup sg ON utsg.slotify_group_id=sg.id 
WHERE utsg.user_id=?
`

func (q *Queries) GetUsersSlotifyGroups(ctx context.Context, userID uint32) ([]SlotifyGroup, error) {
	rows, err := q.query(ctx, q.getUsersSlotifyGroupsStmt, getUsersSlotifyGroups, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SlotifyGroup{}
	for rows.Next() {
		var i SlotifyGroup
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listInvitesByGroup = `-- name: ListInvitesByGroup :many
SELECT 
   i.id AS invite_id, i.message, i.status, i.created_at, i.expiry_date, fu.email AS from_user_email, fu.first_name AS from_user_first_name, fu.last_name AS from_user_last_name, tu.email AS to_user_email, tu.first_name AS to_user_first_name, tu.last_name AS to_user_last_name FROM Invite i
JOIN User fu ON fu.id=i.from_user_id
JOIN User tu ON tu.id=i.to_user_id
WHERE i.status = ifnull(?, i.status) 
AND i.slotify_group_id=?
`

type ListInvitesByGroupParams struct {
	Status         interface{} `json:"status"`
	SlotifyGroupID uint32      `json:"slotifyGroupId"`
}

type ListInvitesByGroupRow struct {
	InviteID          uint32       `json:"inviteId"`
	Message           string       `json:"message"`
	Status            InviteStatus `json:"status"`
	CreatedAt         time.Time    `json:"createdAt"`
	ExpiryDate        time.Time    `json:"expiryDate"`
	FromUserEmail     string       `json:"fromUserEmail"`
	FromUserFirstName string       `json:"fromUserFirstName"`
	FromUserLastName  string       `json:"fromUserLastName"`
	ToUserEmail       string       `json:"toUserEmail"`
	ToUserFirstName   string       `json:"toUserFirstName"`
	ToUserLastName    string       `json:"toUserLastName"`
}

func (q *Queries) ListInvitesByGroup(ctx context.Context, arg ListInvitesByGroupParams) ([]ListInvitesByGroupRow, error) {
	rows, err := q.query(ctx, q.listInvitesByGroupStmt, listInvitesByGroup, arg.Status, arg.SlotifyGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListInvitesByGroupRow{}
	for rows.Next() {
		var i ListInvitesByGroupRow
		if err := rows.Scan(
			&i.InviteID,
			&i.Message,
			&i.Status,
			&i.CreatedAt,
			&i.ExpiryDate,
			&i.FromUserEmail,
			&i.FromUserFirstName,
			&i.FromUserLastName,
			&i.ToUserEmail,
			&i.ToUserFirstName,
			&i.ToUserLastName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listInvitesMe = `-- name: ListInvitesMe :many
SELECT i.id AS invite_id, i.message, i.status,i.created_at, i.expiry_date, fu.email AS from_user_email, fu.first_name AS from_user_first_name, fu.last_name AS from_user_last_name, sg.name AS slotify_group_name FROM Invite i
JOIN User fu ON fu.id=i.from_user_id
JOIN SlotifyGroup sg ON sg.id=i.slotify_group_id
WHERE i.status = ifnull(?, i.status) 
AND i.to_user_id=?
`

type ListInvitesMeParams struct {
	Status   interface{} `json:"status"`
	ToUserID uint32      `json:"toUserId"`
}

type ListInvitesMeRow struct {
	InviteID          uint32       `json:"inviteId"`
	Message           string       `json:"message"`
	Status            InviteStatus `json:"status"`
	CreatedAt         time.Time    `json:"createdAt"`
	ExpiryDate        time.Time    `json:"expiryDate"`
	FromUserEmail     string       `json:"fromUserEmail"`
	FromUserFirstName string       `json:"fromUserFirstName"`
	FromUserLastName  string       `json:"fromUserLastName"`
	SlotifyGroupName  string       `json:"slotifyGroupName"`
}

func (q *Queries) ListInvitesMe(ctx context.Context, arg ListInvitesMeParams) ([]ListInvitesMeRow, error) {
	rows, err := q.query(ctx, q.listInvitesMeStmt, listInvitesMe, arg.Status, arg.ToUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListInvitesMeRow{}
	for rows.Next() {
		var i ListInvitesMeRow
		if err := rows.Scan(
			&i.InviteID,
			&i.Message,
			&i.Status,
			&i.CreatedAt,
			&i.ExpiryDate,
			&i.FromUserEmail,
			&i.FromUserFirstName,
			&i.FromUserLastName,
			&i.SlotifyGroupName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listSlotifyGroups = `-- name: ListSlotifyGroups :many
SELECT id, name FROM SlotifyGroup
WHERE name = ifnull(?, name)
`

func (q *Queries) ListSlotifyGroups(ctx context.Context, name interface{}) ([]SlotifyGroup, error) {
	rows, err := q.query(ctx, q.listSlotifyGroupsStmt, listSlotifyGroups, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SlotifyGroup{}
	for rows.Next() {
		var i SlotifyGroup
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUsers = `-- name: ListUsers :many
SELECT id, email, first_name, last_name FROM User
`

type ListUsersRow struct {
	ID        uint32 `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (q *Queries) ListUsers(ctx context.Context) ([]ListUsersRow, error) {
	rows, err := q.query(ctx, q.listUsersStmt, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUsersRow{}
	for rows.Next() {
		var i ListUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.FirstName,
			&i.LastName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const markNotificationAsRead = `-- name: MarkNotificationAsRead :execrows
UPDATE UserToNotification SET is_read=TRUE
WHERE user_id=? AND notification_id=?
`

type MarkNotificationAsReadParams struct {
	UserID         uint32 `json:"userId"`
	NotificationID uint32 `json:"notificationId"`
}

func (q *Queries) MarkNotificationAsRead(ctx context.Context, arg MarkNotificationAsReadParams) (int64, error) {
	result, err := q.exec(ctx, q.markNotificationAsReadStmt, markNotificationAsRead, arg.UserID, arg.NotificationID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const removeSlotifyGroup = `-- name: RemoveSlotifyGroup :execrows
DELETE FROM SlotifyGroup
WHERE id=?
`

func (q *Queries) RemoveSlotifyGroup(ctx context.Context, id uint32) (int64, error) {
	result, err := q.exec(ctx, q.removeSlotifyGroupStmt, removeSlotifyGroup, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const removeSlotifyGroupMember = `-- name: RemoveSlotifyGroupMember :execrows
DELETE FROM UserToSlotifyGroup
WHERE user_id=? AND slotify_group_id=?
`

type RemoveSlotifyGroupMemberParams struct {
	UserID         uint32 `json:"userId"`
	SlotifyGroupID uint32 `json:"slotifyGroupId"`
}

func (q *Queries) RemoveSlotifyGroupMember(ctx context.Context, arg RemoveSlotifyGroupMemberParams) (int64, error) {
	result, err := q.exec(ctx, q.removeSlotifyGroupMemberStmt, removeSlotifyGroupMember, arg.UserID, arg.SlotifyGroupID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const searchUsersByEmail = `-- name: SearchUsersByEmail :many
SELECT id, email, first_name, last_name FROM User
WHERE LOWER(email) LIKE LOWER(CONCAT('%', ?, '%'))
LIMIT ?
`

type SearchUsersByEmailParams struct {
	Email interface{} `json:"email"`
	Limit int32       `json:"limit"`
}

type SearchUsersByEmailRow struct {
	ID        uint32 `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (q *Queries) SearchUsersByEmail(ctx context.Context, arg SearchUsersByEmailParams) ([]SearchUsersByEmailRow, error) {
	rows, err := q.query(ctx, q.searchUsersByEmailStmt, searchUsersByEmail, arg.Email, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SearchUsersByEmailRow{}
	for rows.Next() {
		var i SearchUsersByEmailRow
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.FirstName,
			&i.LastName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchUsersByName = `-- name: SearchUsersByName :many
SELECT id, email, first_name, last_name FROM User
WHERE LOWER(CONCAT(first_name, ' ', last_name)) LIKE LOWER(CONCAT('%', ?, '%'))
LIMIT ?
`

type SearchUsersByNameParams struct {
	Name  interface{} `json:"name"`
	Limit int32       `json:"limit"`
}

type SearchUsersByNameRow struct {
	ID        uint32 `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (q *Queries) SearchUsersByName(ctx context.Context, arg SearchUsersByNameParams) ([]SearchUsersByNameRow, error) {
	rows, err := q.query(ctx, q.searchUsersByNameStmt, searchUsersByName, arg.Name, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SearchUsersByNameRow{}
	for rows.Next() {
		var i SearchUsersByNameRow
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.FirstName,
			&i.LastName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateInviteMessage = `-- name: UpdateInviteMessage :execrows
UPDATE Invite SET message=?
WHERE id=? AND from_user_id=?
`

type UpdateInviteMessageParams struct {
	Message    string `json:"message"`
	ID         uint32 `json:"id"`
	FromUserID uint32 `json:"fromUserId"`
}

func (q *Queries) UpdateInviteMessage(ctx context.Context, arg UpdateInviteMessageParams) (int64, error) {
	result, err := q.exec(ctx, q.updateInviteMessageStmt, updateInviteMessage, arg.Message, arg.ID, arg.FromUserID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const updateInviteStatus = `-- name: UpdateInviteStatus :execrows
UPDATE Invite SET status=?
WHERE id=?
`

type UpdateInviteStatusParams struct {
	Status InviteStatus `json:"status"`
	ID     uint32       `json:"id"`
}

func (q *Queries) UpdateInviteStatus(ctx context.Context, arg UpdateInviteStatusParams) (int64, error) {
	result, err := q.exec(ctx, q.updateInviteStatusStmt, updateInviteStatus, arg.Status, arg.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const updateUserHomeAccountID = `-- name: UpdateUserHomeAccountID :execrows
UPDATE User SET msft_home_account_id=? WHERE id=?
`

type UpdateUserHomeAccountIDParams struct {
	MsftHomeAccountID sql.NullString `json:"msftHomeAccountId"`
	ID                uint32         `json:"id"`
}

func (q *Queries) UpdateUserHomeAccountID(ctx context.Context, arg UpdateUserHomeAccountIDParams) (int64, error) {
	result, err := q.exec(ctx, q.updateUserHomeAccountIDStmt, updateUserHomeAccountID, arg.MsftHomeAccountID, arg.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
