package database

import (
	"errors"
	"fmt"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
)

var ErrTeamIDInvalid = errors.New("team id does not exist")

type WrongNumberSQLRowsError struct {
	ActualRows   int64
	ExpectedRows int64
}

func (e WrongNumberSQLRowsError) Error() string {
	return fmt.Sprintf("expected %d affected rows, but got %d affected rows", e.ExpectedRows, e.ActualRows)
}

func (e WrongNumberSQLRowsError) Is(target error) bool {
	_, ok := target.(WrongNumberSQLRowsError)
	return ok
}

// IsDuplicateEntrySQLError refers to MySQL error 1062,
// a 'Duplicate entry' SQL error.
func IsDuplicateEntrySQLError(err error) bool {
	return isSpecificMySQLError(err, mysqlerr.ER_DUP_ENTRY)
}

// IsRowDoesNotExistSQLError refers to MySQL error 1452,
// happens when the fk reference does not exist.
func IsRowDoesNotExistSQLError(err error) bool {
	return isSpecificMySQLError(err, mysqlerr.ER_NO_REFERENCED_ROW_2)
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
