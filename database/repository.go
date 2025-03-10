// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.addSlotifyGroupStmt, err = db.PrepareContext(ctx, addSlotifyGroup); err != nil {
		return nil, fmt.Errorf("error preparing query AddSlotifyGroup: %w", err)
	}
	if q.addUserToSlotifyGroupStmt, err = db.PrepareContext(ctx, addUserToSlotifyGroup); err != nil {
		return nil, fmt.Errorf("error preparing query AddUserToSlotifyGroup: %w", err)
	}
	if q.batchDeleteWeekOldInvitesStmt, err = db.PrepareContext(ctx, batchDeleteWeekOldInvites); err != nil {
		return nil, fmt.Errorf("error preparing query BatchDeleteWeekOldInvites: %w", err)
	}
	if q.batchDeleteWeekOldNotificationsStmt, err = db.PrepareContext(ctx, batchDeleteWeekOldNotifications); err != nil {
		return nil, fmt.Errorf("error preparing query BatchDeleteWeekOldNotifications: %w", err)
	}
	if q.batchExpireInvitesStmt, err = db.PrepareContext(ctx, batchExpireInvites); err != nil {
		return nil, fmt.Errorf("error preparing query BatchExpireInvites: %w", err)
	}
	if q.checkMemberInSlotifyGroupStmt, err = db.PrepareContext(ctx, checkMemberInSlotifyGroup); err != nil {
		return nil, fmt.Errorf("error preparing query CheckMemberInSlotifyGroup: %w", err)
	}
	if q.countExpiredInvitesStmt, err = db.PrepareContext(ctx, countExpiredInvites); err != nil {
		return nil, fmt.Errorf("error preparing query CountExpiredInvites: %w", err)
	}
	if q.countSlotifyGroupByIDStmt, err = db.PrepareContext(ctx, countSlotifyGroupByID); err != nil {
		return nil, fmt.Errorf("error preparing query CountSlotifyGroupByID: %w", err)
	}
	if q.countSlotifyGroupMembersStmt, err = db.PrepareContext(ctx, countSlotifyGroupMembers); err != nil {
		return nil, fmt.Errorf("error preparing query CountSlotifyGroupMembers: %w", err)
	}
	if q.countUserByEmailStmt, err = db.PrepareContext(ctx, countUserByEmail); err != nil {
		return nil, fmt.Errorf("error preparing query CountUserByEmail: %w", err)
	}
	if q.countUserByIDStmt, err = db.PrepareContext(ctx, countUserByID); err != nil {
		return nil, fmt.Errorf("error preparing query CountUserByID: %w", err)
	}
	if q.countWeekOldInvitesStmt, err = db.PrepareContext(ctx, countWeekOldInvites); err != nil {
		return nil, fmt.Errorf("error preparing query CountWeekOldInvites: %w", err)
	}
	if q.countWeekOldNotificationsStmt, err = db.PrepareContext(ctx, countWeekOldNotifications); err != nil {
		return nil, fmt.Errorf("error preparing query CountWeekOldNotifications: %w", err)
	}
	if q.createInviteStmt, err = db.PrepareContext(ctx, createInvite); err != nil {
		return nil, fmt.Errorf("error preparing query CreateInvite: %w", err)
	}
	if q.createNotificationStmt, err = db.PrepareContext(ctx, createNotification); err != nil {
		return nil, fmt.Errorf("error preparing query CreateNotification: %w", err)
	}
	if q.createRefreshTokenStmt, err = db.PrepareContext(ctx, createRefreshToken); err != nil {
		return nil, fmt.Errorf("error preparing query CreateRefreshToken: %w", err)
	}
	if q.createUserStmt, err = db.PrepareContext(ctx, createUser); err != nil {
		return nil, fmt.Errorf("error preparing query CreateUser: %w", err)
	}
	if q.createUserNotificationStmt, err = db.PrepareContext(ctx, createUserNotification); err != nil {
		return nil, fmt.Errorf("error preparing query CreateUserNotification: %w", err)
	}
	if q.deleteInviteByIDStmt, err = db.PrepareContext(ctx, deleteInviteByID); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteInviteByID: %w", err)
	}
	if q.deleteRefreshTokenByUserIDStmt, err = db.PrepareContext(ctx, deleteRefreshTokenByUserID); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteRefreshTokenByUserID: %w", err)
	}
	if q.deleteSlotifyGroupByIDStmt, err = db.PrepareContext(ctx, deleteSlotifyGroupByID); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteSlotifyGroupByID: %w", err)
	}
	if q.deleteUserByIDStmt, err = db.PrepareContext(ctx, deleteUserByID); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteUserByID: %w", err)
	}
	if q.getAllSlotifyGroupMembersStmt, err = db.PrepareContext(ctx, getAllSlotifyGroupMembers); err != nil {
		return nil, fmt.Errorf("error preparing query GetAllSlotifyGroupMembers: %w", err)
	}
	if q.getAllSlotifyGroupMembersExceptStmt, err = db.PrepareContext(ctx, getAllSlotifyGroupMembersExcept); err != nil {
		return nil, fmt.Errorf("error preparing query GetAllSlotifyGroupMembersExcept: %w", err)
	}
	if q.getInviteByIDStmt, err = db.PrepareContext(ctx, getInviteByID); err != nil {
		return nil, fmt.Errorf("error preparing query GetInviteByID: %w", err)
	}
	if q.getRefreshTokenByUserIDStmt, err = db.PrepareContext(ctx, getRefreshTokenByUserID); err != nil {
		return nil, fmt.Errorf("error preparing query GetRefreshTokenByUserID: %w", err)
	}
	if q.getSlotifyGroupByIDStmt, err = db.PrepareContext(ctx, getSlotifyGroupByID); err != nil {
		return nil, fmt.Errorf("error preparing query GetSlotifyGroupByID: %w", err)
	}
	if q.getUnreadUserNotificationsStmt, err = db.PrepareContext(ctx, getUnreadUserNotifications); err != nil {
		return nil, fmt.Errorf("error preparing query GetUnreadUserNotifications: %w", err)
	}
	if q.getUserByEmailStmt, err = db.PrepareContext(ctx, getUserByEmail); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserByEmail: %w", err)
	}
	if q.getUserByIDStmt, err = db.PrepareContext(ctx, getUserByID); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserByID: %w", err)
	}
	if q.getUsersSlotifyGroupsStmt, err = db.PrepareContext(ctx, getUsersSlotifyGroups); err != nil {
		return nil, fmt.Errorf("error preparing query GetUsersSlotifyGroups: %w", err)
	}
	if q.listInvitesByGroupStmt, err = db.PrepareContext(ctx, listInvitesByGroup); err != nil {
		return nil, fmt.Errorf("error preparing query ListInvitesByGroup: %w", err)
	}
	if q.listInvitesMeStmt, err = db.PrepareContext(ctx, listInvitesMe); err != nil {
		return nil, fmt.Errorf("error preparing query ListInvitesMe: %w", err)
	}
	if q.listSlotifyGroupsStmt, err = db.PrepareContext(ctx, listSlotifyGroups); err != nil {
		return nil, fmt.Errorf("error preparing query ListSlotifyGroups: %w", err)
	}
	if q.listUsersStmt, err = db.PrepareContext(ctx, listUsers); err != nil {
		return nil, fmt.Errorf("error preparing query ListUsers: %w", err)
	}
	if q.markNotificationAsReadStmt, err = db.PrepareContext(ctx, markNotificationAsRead); err != nil {
		return nil, fmt.Errorf("error preparing query MarkNotificationAsRead: %w", err)
	}
	if q.removeSlotifyGroupStmt, err = db.PrepareContext(ctx, removeSlotifyGroup); err != nil {
		return nil, fmt.Errorf("error preparing query RemoveSlotifyGroup: %w", err)
	}
	if q.removeSlotifyGroupMemberStmt, err = db.PrepareContext(ctx, removeSlotifyGroupMember); err != nil {
		return nil, fmt.Errorf("error preparing query RemoveSlotifyGroupMember: %w", err)
	}
	if q.searchUsersByEmailStmt, err = db.PrepareContext(ctx, searchUsersByEmail); err != nil {
		return nil, fmt.Errorf("error preparing query SearchUsersByEmail: %w", err)
	}
	if q.searchUsersByNameStmt, err = db.PrepareContext(ctx, searchUsersByName); err != nil {
		return nil, fmt.Errorf("error preparing query SearchUsersByName: %w", err)
	}
	if q.updateInviteMessageStmt, err = db.PrepareContext(ctx, updateInviteMessage); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateInviteMessage: %w", err)
	}
	if q.updateInviteStatusStmt, err = db.PrepareContext(ctx, updateInviteStatus); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateInviteStatus: %w", err)
	}
	if q.updateUserHomeAccountIDStmt, err = db.PrepareContext(ctx, updateUserHomeAccountID); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateUserHomeAccountID: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.addSlotifyGroupStmt != nil {
		if cerr := q.addSlotifyGroupStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addSlotifyGroupStmt: %w", cerr)
		}
	}
	if q.addUserToSlotifyGroupStmt != nil {
		if cerr := q.addUserToSlotifyGroupStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addUserToSlotifyGroupStmt: %w", cerr)
		}
	}
	if q.batchDeleteWeekOldInvitesStmt != nil {
		if cerr := q.batchDeleteWeekOldInvitesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing batchDeleteWeekOldInvitesStmt: %w", cerr)
		}
	}
	if q.batchDeleteWeekOldNotificationsStmt != nil {
		if cerr := q.batchDeleteWeekOldNotificationsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing batchDeleteWeekOldNotificationsStmt: %w", cerr)
		}
	}
	if q.batchExpireInvitesStmt != nil {
		if cerr := q.batchExpireInvitesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing batchExpireInvitesStmt: %w", cerr)
		}
	}
	if q.checkMemberInSlotifyGroupStmt != nil {
		if cerr := q.checkMemberInSlotifyGroupStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing checkMemberInSlotifyGroupStmt: %w", cerr)
		}
	}
	if q.countExpiredInvitesStmt != nil {
		if cerr := q.countExpiredInvitesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing countExpiredInvitesStmt: %w", cerr)
		}
	}
	if q.countSlotifyGroupByIDStmt != nil {
		if cerr := q.countSlotifyGroupByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing countSlotifyGroupByIDStmt: %w", cerr)
		}
	}
	if q.countSlotifyGroupMembersStmt != nil {
		if cerr := q.countSlotifyGroupMembersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing countSlotifyGroupMembersStmt: %w", cerr)
		}
	}
	if q.countUserByEmailStmt != nil {
		if cerr := q.countUserByEmailStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing countUserByEmailStmt: %w", cerr)
		}
	}
	if q.countUserByIDStmt != nil {
		if cerr := q.countUserByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing countUserByIDStmt: %w", cerr)
		}
	}
	if q.countWeekOldInvitesStmt != nil {
		if cerr := q.countWeekOldInvitesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing countWeekOldInvitesStmt: %w", cerr)
		}
	}
	if q.countWeekOldNotificationsStmt != nil {
		if cerr := q.countWeekOldNotificationsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing countWeekOldNotificationsStmt: %w", cerr)
		}
	}
	if q.createInviteStmt != nil {
		if cerr := q.createInviteStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createInviteStmt: %w", cerr)
		}
	}
	if q.createNotificationStmt != nil {
		if cerr := q.createNotificationStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createNotificationStmt: %w", cerr)
		}
	}
	if q.createRefreshTokenStmt != nil {
		if cerr := q.createRefreshTokenStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createRefreshTokenStmt: %w", cerr)
		}
	}
	if q.createUserStmt != nil {
		if cerr := q.createUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createUserStmt: %w", cerr)
		}
	}
	if q.createUserNotificationStmt != nil {
		if cerr := q.createUserNotificationStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createUserNotificationStmt: %w", cerr)
		}
	}
	if q.deleteInviteByIDStmt != nil {
		if cerr := q.deleteInviteByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteInviteByIDStmt: %w", cerr)
		}
	}
	if q.deleteRefreshTokenByUserIDStmt != nil {
		if cerr := q.deleteRefreshTokenByUserIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteRefreshTokenByUserIDStmt: %w", cerr)
		}
	}
	if q.deleteSlotifyGroupByIDStmt != nil {
		if cerr := q.deleteSlotifyGroupByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteSlotifyGroupByIDStmt: %w", cerr)
		}
	}
	if q.deleteUserByIDStmt != nil {
		if cerr := q.deleteUserByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteUserByIDStmt: %w", cerr)
		}
	}
	if q.getAllSlotifyGroupMembersStmt != nil {
		if cerr := q.getAllSlotifyGroupMembersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getAllSlotifyGroupMembersStmt: %w", cerr)
		}
	}
	if q.getAllSlotifyGroupMembersExceptStmt != nil {
		if cerr := q.getAllSlotifyGroupMembersExceptStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getAllSlotifyGroupMembersExceptStmt: %w", cerr)
		}
	}
	if q.getInviteByIDStmt != nil {
		if cerr := q.getInviteByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getInviteByIDStmt: %w", cerr)
		}
	}
	if q.getRefreshTokenByUserIDStmt != nil {
		if cerr := q.getRefreshTokenByUserIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getRefreshTokenByUserIDStmt: %w", cerr)
		}
	}
	if q.getSlotifyGroupByIDStmt != nil {
		if cerr := q.getSlotifyGroupByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getSlotifyGroupByIDStmt: %w", cerr)
		}
	}
	if q.getUnreadUserNotificationsStmt != nil {
		if cerr := q.getUnreadUserNotificationsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUnreadUserNotificationsStmt: %w", cerr)
		}
	}
	if q.getUserByEmailStmt != nil {
		if cerr := q.getUserByEmailStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserByEmailStmt: %w", cerr)
		}
	}
	if q.getUserByIDStmt != nil {
		if cerr := q.getUserByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserByIDStmt: %w", cerr)
		}
	}
	if q.getUsersSlotifyGroupsStmt != nil {
		if cerr := q.getUsersSlotifyGroupsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUsersSlotifyGroupsStmt: %w", cerr)
		}
	}
	if q.listInvitesByGroupStmt != nil {
		if cerr := q.listInvitesByGroupStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listInvitesByGroupStmt: %w", cerr)
		}
	}
	if q.listInvitesMeStmt != nil {
		if cerr := q.listInvitesMeStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listInvitesMeStmt: %w", cerr)
		}
	}
	if q.listSlotifyGroupsStmt != nil {
		if cerr := q.listSlotifyGroupsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listSlotifyGroupsStmt: %w", cerr)
		}
	}
	if q.listUsersStmt != nil {
		if cerr := q.listUsersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listUsersStmt: %w", cerr)
		}
	}
	if q.markNotificationAsReadStmt != nil {
		if cerr := q.markNotificationAsReadStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing markNotificationAsReadStmt: %w", cerr)
		}
	}
	if q.removeSlotifyGroupStmt != nil {
		if cerr := q.removeSlotifyGroupStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing removeSlotifyGroupStmt: %w", cerr)
		}
	}
	if q.removeSlotifyGroupMemberStmt != nil {
		if cerr := q.removeSlotifyGroupMemberStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing removeSlotifyGroupMemberStmt: %w", cerr)
		}
	}
	if q.searchUsersByEmailStmt != nil {
		if cerr := q.searchUsersByEmailStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing searchUsersByEmailStmt: %w", cerr)
		}
	}
	if q.searchUsersByNameStmt != nil {
		if cerr := q.searchUsersByNameStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing searchUsersByNameStmt: %w", cerr)
		}
	}
	if q.updateInviteMessageStmt != nil {
		if cerr := q.updateInviteMessageStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateInviteMessageStmt: %w", cerr)
		}
	}
	if q.updateInviteStatusStmt != nil {
		if cerr := q.updateInviteStatusStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateInviteStatusStmt: %w", cerr)
		}
	}
	if q.updateUserHomeAccountIDStmt != nil {
		if cerr := q.updateUserHomeAccountIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateUserHomeAccountIDStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                                  DBTX
	tx                                  *sql.Tx
	addSlotifyGroupStmt                 *sql.Stmt
	addUserToSlotifyGroupStmt           *sql.Stmt
	batchDeleteWeekOldInvitesStmt       *sql.Stmt
	batchDeleteWeekOldNotificationsStmt *sql.Stmt
	batchExpireInvitesStmt              *sql.Stmt
	checkMemberInSlotifyGroupStmt       *sql.Stmt
	countExpiredInvitesStmt             *sql.Stmt
	countSlotifyGroupByIDStmt           *sql.Stmt
	countSlotifyGroupMembersStmt        *sql.Stmt
	countUserByEmailStmt                *sql.Stmt
	countUserByIDStmt                   *sql.Stmt
	countWeekOldInvitesStmt             *sql.Stmt
	countWeekOldNotificationsStmt       *sql.Stmt
	createInviteStmt                    *sql.Stmt
	createNotificationStmt              *sql.Stmt
	createRefreshTokenStmt              *sql.Stmt
	createUserStmt                      *sql.Stmt
	createUserNotificationStmt          *sql.Stmt
	deleteInviteByIDStmt                *sql.Stmt
	deleteRefreshTokenByUserIDStmt      *sql.Stmt
	deleteSlotifyGroupByIDStmt          *sql.Stmt
	deleteUserByIDStmt                  *sql.Stmt
	getAllSlotifyGroupMembersStmt       *sql.Stmt
	getAllSlotifyGroupMembersExceptStmt *sql.Stmt
	getInviteByIDStmt                   *sql.Stmt
	getRefreshTokenByUserIDStmt         *sql.Stmt
	getSlotifyGroupByIDStmt             *sql.Stmt
	getUnreadUserNotificationsStmt      *sql.Stmt
	getUserByEmailStmt                  *sql.Stmt
	getUserByIDStmt                     *sql.Stmt
	getUsersSlotifyGroupsStmt           *sql.Stmt
	listInvitesByGroupStmt              *sql.Stmt
	listInvitesMeStmt                   *sql.Stmt
	listSlotifyGroupsStmt               *sql.Stmt
	listUsersStmt                       *sql.Stmt
	markNotificationAsReadStmt          *sql.Stmt
	removeSlotifyGroupStmt              *sql.Stmt
	removeSlotifyGroupMemberStmt        *sql.Stmt
	searchUsersByEmailStmt              *sql.Stmt
	searchUsersByNameStmt               *sql.Stmt
	updateInviteMessageStmt             *sql.Stmt
	updateInviteStatusStmt              *sql.Stmt
	updateUserHomeAccountIDStmt         *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                                  tx,
		tx:                                  tx,
		addSlotifyGroupStmt:                 q.addSlotifyGroupStmt,
		addUserToSlotifyGroupStmt:           q.addUserToSlotifyGroupStmt,
		batchDeleteWeekOldInvitesStmt:       q.batchDeleteWeekOldInvitesStmt,
		batchDeleteWeekOldNotificationsStmt: q.batchDeleteWeekOldNotificationsStmt,
		batchExpireInvitesStmt:              q.batchExpireInvitesStmt,
		checkMemberInSlotifyGroupStmt:       q.checkMemberInSlotifyGroupStmt,
		countExpiredInvitesStmt:             q.countExpiredInvitesStmt,
		countSlotifyGroupByIDStmt:           q.countSlotifyGroupByIDStmt,
		countSlotifyGroupMembersStmt:        q.countSlotifyGroupMembersStmt,
		countUserByEmailStmt:                q.countUserByEmailStmt,
		countUserByIDStmt:                   q.countUserByIDStmt,
		countWeekOldInvitesStmt:             q.countWeekOldInvitesStmt,
		countWeekOldNotificationsStmt:       q.countWeekOldNotificationsStmt,
		createInviteStmt:                    q.createInviteStmt,
		createNotificationStmt:              q.createNotificationStmt,
		createRefreshTokenStmt:              q.createRefreshTokenStmt,
		createUserStmt:                      q.createUserStmt,
		createUserNotificationStmt:          q.createUserNotificationStmt,
		deleteInviteByIDStmt:                q.deleteInviteByIDStmt,
		deleteRefreshTokenByUserIDStmt:      q.deleteRefreshTokenByUserIDStmt,
		deleteSlotifyGroupByIDStmt:          q.deleteSlotifyGroupByIDStmt,
		deleteUserByIDStmt:                  q.deleteUserByIDStmt,
		getAllSlotifyGroupMembersStmt:       q.getAllSlotifyGroupMembersStmt,
		getAllSlotifyGroupMembersExceptStmt: q.getAllSlotifyGroupMembersExceptStmt,
		getInviteByIDStmt:                   q.getInviteByIDStmt,
		getRefreshTokenByUserIDStmt:         q.getRefreshTokenByUserIDStmt,
		getSlotifyGroupByIDStmt:             q.getSlotifyGroupByIDStmt,
		getUnreadUserNotificationsStmt:      q.getUnreadUserNotificationsStmt,
		getUserByEmailStmt:                  q.getUserByEmailStmt,
		getUserByIDStmt:                     q.getUserByIDStmt,
		getUsersSlotifyGroupsStmt:           q.getUsersSlotifyGroupsStmt,
		listInvitesByGroupStmt:              q.listInvitesByGroupStmt,
		listInvitesMeStmt:                   q.listInvitesMeStmt,
		listSlotifyGroupsStmt:               q.listSlotifyGroupsStmt,
		listUsersStmt:                       q.listUsersStmt,
		markNotificationAsReadStmt:          q.markNotificationAsReadStmt,
		removeSlotifyGroupStmt:              q.removeSlotifyGroupStmt,
		removeSlotifyGroupMemberStmt:        q.removeSlotifyGroupMemberStmt,
		searchUsersByEmailStmt:              q.searchUsersByEmailStmt,
		searchUsersByNameStmt:               q.searchUsersByNameStmt,
		updateInviteMessageStmt:             q.updateInviteMessageStmt,
		updateInviteStatusStmt:              q.updateInviteStatusStmt,
		updateUserHomeAccountIDStmt:         q.updateUserHomeAccountIDStmt,
	}
}
