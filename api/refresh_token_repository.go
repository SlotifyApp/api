package api

import (
	"database/sql"
	"fmt"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/jwt"
	"go.uber.org/zap"
)

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
