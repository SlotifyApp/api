package database

import "errors"

// affected SQL rows after a stmt.Exec.
var ErrWrongNumberRows = errors.New("teamsservice: unexpected number of sql affected rows")
