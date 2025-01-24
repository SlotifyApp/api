package database

import (
	"database/sql"
	"errors"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
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

// IsDuplicateEntrySQLError will check if the error
// matches 'Duplicate entry' SQL error.
func IsDuplicateEntrySQLError(err error) bool {
	return isSpecificMySQLError(err, mysqlerr.ER_DUP_ENTRY)
}

// code is equal to target error code.
func isSpecificMySQLError(err error, errorCode uint16) bool {
	var mysqlErr *mysql.MySQLError
	// Check if err is instance of MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == errorCode
	}
	return false
}
