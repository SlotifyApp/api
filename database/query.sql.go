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

const addTeam = `-- name: AddTeam :execlastid
INSERT INTO Team (name) VALUES (?)
`

func (q *Queries) AddTeam(ctx context.Context, name string) (int64, error) {
	result, err := q.exec(ctx, q.addTeamStmt, addTeam, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const addUserToTeam = `-- name: AddUserToTeam :execrows
INSERT INTO UserToTeam (user_id, team_id) VALUES (?, ?)
`

type AddUserToTeamParams struct {
	UserID uint32 `json:"userId"`
	TeamID uint32 `json:"teamId"`
}

func (q *Queries) AddUserToTeam(ctx context.Context, arg AddUserToTeamParams) (int64, error) {
	result, err := q.exec(ctx, q.addUserToTeamStmt, addUserToTeam, arg.UserID, arg.TeamID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const countTeamByID = `-- name: CountTeamByID :one
SELECT COUNT(*) FROM Team WHERE id=?
`

func (q *Queries) CountTeamByID(ctx context.Context, id uint32) (int64, error) {
	row := q.queryRow(ctx, q.countTeamByIDStmt, countTeamByID, id)
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

const deleteTeamByID = `-- name: DeleteTeamByID :execrows
DELETE FROM Team WHERE id=?
`

func (q *Queries) DeleteTeamByID(ctx context.Context, id uint32) (int64, error) {
	result, err := q.exec(ctx, q.deleteTeamByIDStmt, deleteTeamByID, id)
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

const getAllTeamMembers = `-- name: GetAllTeamMembers :many
SELECT u.id, u.email, u.first_name, u.last_name FROM Team t
JOIN UserToTeam utt ON t.id=utt.team_id
JOIN User u ON u.id=utt.user_id 
WHERE t.id=?
`

type GetAllTeamMembersRow struct {
	ID        uint32 `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (q *Queries) GetAllTeamMembers(ctx context.Context, id uint32) ([]GetAllTeamMembersRow, error) {
	rows, err := q.query(ctx, q.getAllTeamMembersStmt, getAllTeamMembers, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllTeamMembersRow{}
	for rows.Next() {
		var i GetAllTeamMembersRow
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

const getAllTeamMembersExcept = `-- name: GetAllTeamMembersExcept :many
SELECT u.id FROM Team t
JOIN UserToTeam utt ON t.id=utt.team_id
JOIN User u ON u.id=utt.user_id 
WHERE t.id=? AND u.id!=?
`

type GetAllTeamMembersExceptParams struct {
	TeamID uint32 `json:"teamID"`
	UserID uint32 `json:"userID"`
}

func (q *Queries) GetAllTeamMembersExcept(ctx context.Context, arg GetAllTeamMembersExceptParams) ([]uint32, error) {
	rows, err := q.query(ctx, q.getAllTeamMembersExceptStmt, getAllTeamMembersExcept, arg.TeamID, arg.UserID)
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

const getJoinableTeams = `-- name: GetJoinableTeams :many
SELECT t.id, t.name FROM Team t
LEFT JOIN UserToTeam utt ON
     t.id = utt.team_id AND utt.user_id = ? 
WHERE utt.user_id IS NULL
`

func (q *Queries) GetJoinableTeams(ctx context.Context, userID uint32) ([]Team, error) {
	rows, err := q.query(ctx, q.getJoinableTeamsStmt, getJoinableTeams, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Team{}
	for rows.Next() {
		var i Team
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

const getTeamByID = `-- name: GetTeamByID :one
SELECT id, name FROM Team WHERE id=?
`

func (q *Queries) GetTeamByID(ctx context.Context, id uint32) (Team, error) {
	row := q.queryRow(ctx, q.getTeamByIDStmt, getTeamByID, id)
	var i Team
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

const getUsersTeams = `-- name: GetUsersTeams :many
SELECT t.id, t.name FROM UserToTeam utt
JOIN Team t ON utt.team_id=t.id 
WHERE utt.user_id=?
`

func (q *Queries) GetUsersTeams(ctx context.Context, userID uint32) ([]Team, error) {
	rows, err := q.query(ctx, q.getUsersTeamsStmt, getUsersTeams, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Team{}
	for rows.Next() {
		var i Team
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

const listTeams = `-- name: ListTeams :many
SELECT id, name FROM Team
WHERE name = ifnull(?, name)
`

func (q *Queries) ListTeams(ctx context.Context, name interface{}) ([]Team, error) {
	rows, err := q.query(ctx, q.listTeamsStmt, listTeams, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Team{}
	for rows.Next() {
		var i Team
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
WHERE email = ifnull(?, email)
   AND first_name = ifnull(?, first_name)
   AND last_name = ifnull(?, last_name)
`

type ListUsersParams struct {
	Email     interface{} `json:"email"`
	FirstName interface{} `json:"firstName"`
	LastName  interface{} `json:"lastName"`
}

type ListUsersRow struct {
	ID        uint32 `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]ListUsersRow, error) {
	rows, err := q.query(ctx, q.listUsersStmt, listUsers, arg.Email, arg.FirstName, arg.LastName)
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
