package database

import (
	"database/sql"

	"go.uber.org/zap"
)

// CloseStmt closes a sql.Stmt, use with defer.
func CloseStmt(stmt *sql.Stmt, logger *zap.SugaredLogger) {
	if stmt != nil {
		if err := stmt.Close(); err != nil {
			logger.Warn("database: failed to close rows", zap.Error(err))
		}
	}
}

// CloseRows closes a sql.Rows, use with defer.
func CloseRows(rows *sql.Rows, logger *zap.SugaredLogger) {
	if rows != nil {
		if err := rows.Close(); err != nil {
			logger.Warn("database: failed to close rows", zap.Error(err))
		}
	}
}
