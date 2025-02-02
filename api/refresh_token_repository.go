package api

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/jwt"
	"go.uber.org/zap"
)

// Refers to the UserToMSFTRefreshToken table.
type UserToMSFTRefreshTokenRepositoryInterface interface {
	// Stores/Updates the Microsoft OAuth refresh token tied to a specific user
	StoreMicrosoftRefreshToken(int, string) error
	// Returns Microsoft refresh token of a user. If the user doesn't have one, then error is returned
	GetMicrosoftRefreshToken(int) (string, error)
}

type UserToMSFTRefreshTokenRepository struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewUserToMSFTRefreshTokenRepository(logger *zap.SugaredLogger, db *sql.DB) UserToMSFTRefreshTokenRepository {
	return UserToMSFTRefreshTokenRepository{
		logger: logger,
		db:     db,
	}
}

// check UserToMSFTRefreshTokenRepository conforms to the interface.
var _ UserToMSFTRefreshTokenRepositoryInterface = (*UserToMSFTRefreshTokenRepository)(nil)

func (utm UserToMSFTRefreshTokenRepository) GetMicrosoftRefreshToken(userID int) (string, error) {
	var stmt *sql.Stmt
	var err error
	if stmt, err = utm.db.Prepare("SELECT * FROM UserToMSFTRefreshToken WHERE user_id=?"); err != nil {
		return "", fmt.Errorf("user to msft refresh tok repository: %s: %w", PrepareStmtFail, err)
	}
	defer database.CloseStmt(stmt, utm.logger)

	var refreshToken string
	if err = stmt.QueryRow(userID).Scan(&refreshToken); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("user to msft refresh tok repository failed to scan: %w", err)
	}

	return refreshToken, nil
}

func (utm UserToMSFTRefreshTokenRepository) StoreMicrosoftRefreshToken(userID int, msftTok string) error {
	var stmt *sql.Stmt
	var err error
	// If that row already exists for a user, then
	// replace it. Otherwise insert
	if stmt, err = utm.db.Prepare("REPLACE INTO UserToMSFTRefreshToken (user_id, token) VALUES (?, ?)"); err != nil {
		return fmt.Errorf("userToMicrosoftRefreshToken repository: %s: %w", PrepareStmtFail, err)
	}
	defer database.CloseStmt(stmt, utm.logger)

	var res sql.Result
	if res, err = stmt.Exec(userID, msftTok); err != nil {
		return fmt.Errorf("userToMicrosoftRefreshToken repository: %s: %w", QueryDBFail, err)
	}
	var rows int64
	if rows, err = res.RowsAffected(); err != nil {
		return fmt.Errorf("userToMicrosoftRefreshToken repository failed to get rows affected: %w", err)
	}
	// REPLACE will affect either 1 or 2 rows dependent on whether
	// the db is updating an existing row
	if rows < 1 || rows > 2 {
		return fmt.Errorf("userToMicrosoftRefreshToken repository: %w",
			database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1, 2}},
		)
	}

	return nil
}

// Refers to the RefreshToken table.
type RefreshTokenRepositoryInterface interface {
	// Creates and returns a Slotify refresh token for a specific user
	CreateRefreshToken(int, string) (string, error)
	// Deletes a user's refresh token
	DeleteRefreshTokenByUserID(int) error
	// Returns refresh token of user. If the user doesn't have one, then error is returned:
	// sql.ErrNoRows is wrapped
	GetRefreshTokenByUserID(int) (jwt.RefreshToken, error)
}

type RefreshTokenRepository struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewRefreshTokenRepository(logger *zap.SugaredLogger, db *sql.DB) RefreshTokenRepository {
	return RefreshTokenRepository{
		logger: logger,
		db:     db,
	}
}

// check RefreshTokenRepository conforms to the interface.
var _ RefreshTokenRepositoryInterface = (*RefreshTokenRepository)(nil)

func (rr RefreshTokenRepository) CreateRefreshToken(userID int, email string) (string, error) {
	var stmt *sql.Stmt
	var err error
	// If that row already exists for a user, then
	// delete and insert a new one. Otherwise, just insert
	if stmt, err = rr.db.Prepare("REPLACE INTO RefreshToken (user_id, token) VALUES (?, ?)"); err != nil {
		return "", fmt.Errorf("refreshtoken repository: %s: %w", PrepareStmtFail, err)
	}
	defer database.CloseStmt(stmt, rr.logger)

	var refreshToken string
	if refreshToken, err = jwt.CreateNewJWT(userID, email, jwt.RefreshTokenJWTSecretEnv); err != nil {
		return "", fmt.Errorf("failed to CreateRefreshToken: %w", err)
	}

	var res sql.Result
	if res, err = stmt.Exec(userID, refreshToken); err != nil {
		return "", fmt.Errorf("refreshtoken repository: %s: %w", QueryDBFail, err)
	}
	var rows int64
	if rows, err = res.RowsAffected(); err != nil {
		return "", fmt.Errorf("refresh repository failed to get rows affected: %w", err)
	}

	// REPLACE will affect either 1 or 2 rows dependent on whether
	// the db is updating an existing row
	if rows < 1 || rows > 2 {
		return "", fmt.Errorf("refresh repository: %w",
			database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}})
	}

	return refreshToken, nil
}

func (rr RefreshTokenRepository) GetRefreshTokenByUserID(userID int) (jwt.RefreshToken, error) {
	var stmt *sql.Stmt
	var err error
	if stmt, err = rr.db.Prepare("SELECT * FROM RefreshToken WHERE user_id=?"); err != nil {
		return jwt.RefreshToken{}, fmt.Errorf("refreshtoken repository: %s: %w", PrepareStmtFail, err)
	}
	defer database.CloseStmt(stmt, rr.logger)

	refreshToken := jwt.RefreshToken{}
	if err = stmt.QueryRow(userID).Scan(&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.Revoked); err != nil {
		return jwt.RefreshToken{}, fmt.Errorf("refreshToken repository failed to scan: %w", err)
	}

	return refreshToken, nil
}

func (rr RefreshTokenRepository) DeleteRefreshTokenByUserID(userID int) error {
	var err error
	// Check existence of refresh token first
	if _, err = rr.GetRefreshTokenByUserID(userID); err != nil {
		return fmt.Errorf("refreshtoken repo: failed to delete: %w", err)
	}

	var res sql.Result

	if res, err = rr.db.Exec("DELETE FROM RefreshToken WHERE user_id=?", userID); err != nil {
		return fmt.Errorf("refreshtoken repository: %s: %w", QueryDBFail, err)
	}

	var rows int64
	if rows, err = res.RowsAffected(); err != nil {
		return fmt.Errorf("refreshtoken repository fails to get rows affected: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("refreshtoken repository: %w",
			database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}})
	}

	return nil
}
