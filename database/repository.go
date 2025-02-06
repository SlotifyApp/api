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
	if q.addTeamStmt, err = db.PrepareContext(ctx, addTeam); err != nil {
		return nil, fmt.Errorf("error preparing query AddTeam: %w", err)
	}
	if q.addUserToTeamStmt, err = db.PrepareContext(ctx, addUserToTeam); err != nil {
		return nil, fmt.Errorf("error preparing query AddUserToTeam: %w", err)
	}
	if q.countTeamByIDStmt, err = db.PrepareContext(ctx, countTeamByID); err != nil {
		return nil, fmt.Errorf("error preparing query CountTeamByID: %w", err)
	}
	if q.countUserByEmailStmt, err = db.PrepareContext(ctx, countUserByEmail); err != nil {
		return nil, fmt.Errorf("error preparing query CountUserByEmail: %w", err)
	}
	if q.countUserByIDStmt, err = db.PrepareContext(ctx, countUserByID); err != nil {
		return nil, fmt.Errorf("error preparing query CountUserByID: %w", err)
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
	if q.deleteRefreshTokenByUserIDStmt, err = db.PrepareContext(ctx, deleteRefreshTokenByUserID); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteRefreshTokenByUserID: %w", err)
	}
	if q.deleteTeamByIDStmt, err = db.PrepareContext(ctx, deleteTeamByID); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteTeamByID: %w", err)
	}
	if q.deleteUserByIDStmt, err = db.PrepareContext(ctx, deleteUserByID); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteUserByID: %w", err)
	}
	if q.getAllTeamMembersStmt, err = db.PrepareContext(ctx, getAllTeamMembers); err != nil {
		return nil, fmt.Errorf("error preparing query GetAllTeamMembers: %w", err)
	}
	if q.getJoinableTeamsStmt, err = db.PrepareContext(ctx, getJoinableTeams); err != nil {
		return nil, fmt.Errorf("error preparing query GetJoinableTeams: %w", err)
	}
	if q.getRefreshTokenByUserIDStmt, err = db.PrepareContext(ctx, getRefreshTokenByUserID); err != nil {
		return nil, fmt.Errorf("error preparing query GetRefreshTokenByUserID: %w", err)
	}
	if q.getTeamByIDStmt, err = db.PrepareContext(ctx, getTeamByID); err != nil {
		return nil, fmt.Errorf("error preparing query GetTeamByID: %w", err)
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
	if q.getUsersTeamsStmt, err = db.PrepareContext(ctx, getUsersTeams); err != nil {
		return nil, fmt.Errorf("error preparing query GetUsersTeams: %w", err)
	}
	if q.listTeamsStmt, err = db.PrepareContext(ctx, listTeams); err != nil {
		return nil, fmt.Errorf("error preparing query ListTeams: %w", err)
	}
	if q.listUsersStmt, err = db.PrepareContext(ctx, listUsers); err != nil {
		return nil, fmt.Errorf("error preparing query ListUsers: %w", err)
	}
	if q.markNotificationAsReadStmt, err = db.PrepareContext(ctx, markNotificationAsRead); err != nil {
		return nil, fmt.Errorf("error preparing query MarkNotificationAsRead: %w", err)
	}
	if q.updateUserHomeAccountIDStmt, err = db.PrepareContext(ctx, updateUserHomeAccountID); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateUserHomeAccountID: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.addTeamStmt != nil {
		if cerr := q.addTeamStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addTeamStmt: %w", cerr)
		}
	}
	if q.addUserToTeamStmt != nil {
		if cerr := q.addUserToTeamStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addUserToTeamStmt: %w", cerr)
		}
	}
	if q.countTeamByIDStmt != nil {
		if cerr := q.countTeamByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing countTeamByIDStmt: %w", cerr)
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
	if q.deleteRefreshTokenByUserIDStmt != nil {
		if cerr := q.deleteRefreshTokenByUserIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteRefreshTokenByUserIDStmt: %w", cerr)
		}
	}
	if q.deleteTeamByIDStmt != nil {
		if cerr := q.deleteTeamByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteTeamByIDStmt: %w", cerr)
		}
	}
	if q.deleteUserByIDStmt != nil {
		if cerr := q.deleteUserByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteUserByIDStmt: %w", cerr)
		}
	}
	if q.getAllTeamMembersStmt != nil {
		if cerr := q.getAllTeamMembersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getAllTeamMembersStmt: %w", cerr)
		}
	}
	if q.getJoinableTeamsStmt != nil {
		if cerr := q.getJoinableTeamsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getJoinableTeamsStmt: %w", cerr)
		}
	}
	if q.getRefreshTokenByUserIDStmt != nil {
		if cerr := q.getRefreshTokenByUserIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getRefreshTokenByUserIDStmt: %w", cerr)
		}
	}
	if q.getTeamByIDStmt != nil {
		if cerr := q.getTeamByIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getTeamByIDStmt: %w", cerr)
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
	if q.getUsersTeamsStmt != nil {
		if cerr := q.getUsersTeamsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUsersTeamsStmt: %w", cerr)
		}
	}
	if q.listTeamsStmt != nil {
		if cerr := q.listTeamsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listTeamsStmt: %w", cerr)
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
	db                             DBTX
	tx                             *sql.Tx
	addTeamStmt                    *sql.Stmt
	addUserToTeamStmt              *sql.Stmt
	countTeamByIDStmt              *sql.Stmt
	countUserByEmailStmt           *sql.Stmt
	countUserByIDStmt              *sql.Stmt
	createNotificationStmt         *sql.Stmt
	createRefreshTokenStmt         *sql.Stmt
	createUserStmt                 *sql.Stmt
	createUserNotificationStmt     *sql.Stmt
	deleteRefreshTokenByUserIDStmt *sql.Stmt
	deleteTeamByIDStmt             *sql.Stmt
	deleteUserByIDStmt             *sql.Stmt
	getAllTeamMembersStmt          *sql.Stmt
	getJoinableTeamsStmt           *sql.Stmt
	getRefreshTokenByUserIDStmt    *sql.Stmt
	getTeamByIDStmt                *sql.Stmt
	getUnreadUserNotificationsStmt *sql.Stmt
	getUserByEmailStmt             *sql.Stmt
	getUserByIDStmt                *sql.Stmt
	getUsersTeamsStmt              *sql.Stmt
	listTeamsStmt                  *sql.Stmt
	listUsersStmt                  *sql.Stmt
	markNotificationAsReadStmt     *sql.Stmt
	updateUserHomeAccountIDStmt    *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                             tx,
		tx:                             tx,
		addTeamStmt:                    q.addTeamStmt,
		addUserToTeamStmt:              q.addUserToTeamStmt,
		countTeamByIDStmt:              q.countTeamByIDStmt,
		countUserByEmailStmt:           q.countUserByEmailStmt,
		countUserByIDStmt:              q.countUserByIDStmt,
		createNotificationStmt:         q.createNotificationStmt,
		createRefreshTokenStmt:         q.createRefreshTokenStmt,
		createUserStmt:                 q.createUserStmt,
		createUserNotificationStmt:     q.createUserNotificationStmt,
		deleteRefreshTokenByUserIDStmt: q.deleteRefreshTokenByUserIDStmt,
		deleteTeamByIDStmt:             q.deleteTeamByIDStmt,
		deleteUserByIDStmt:             q.deleteUserByIDStmt,
		getAllTeamMembersStmt:          q.getAllTeamMembersStmt,
		getJoinableTeamsStmt:           q.getJoinableTeamsStmt,
		getRefreshTokenByUserIDStmt:    q.getRefreshTokenByUserIDStmt,
		getTeamByIDStmt:                q.getTeamByIDStmt,
		getUnreadUserNotificationsStmt: q.getUnreadUserNotificationsStmt,
		getUserByEmailStmt:             q.getUserByEmailStmt,
		getUserByIDStmt:                q.getUserByIDStmt,
		getUsersTeamsStmt:              q.getUsersTeamsStmt,
		listTeamsStmt:                  q.listTeamsStmt,
		listUsersStmt:                  q.listUsersStmt,
		markNotificationAsReadStmt:     q.markNotificationAsReadStmt,
		updateUserHomeAccountIDStmt:    q.updateUserHomeAccountIDStmt,
	}
}
