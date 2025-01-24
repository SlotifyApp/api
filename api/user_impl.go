package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// (GET /users) Get a user by query params.
func (s Server) GetUsers(w http.ResponseWriter, _ *http.Request, params GetUsersParams) {
	var conditions []string
	var args []interface{}

	if params.Email != nil {
		conditions = append(conditions, "email=?")
		args = append(args, string(*params.Email))
	}

	if params.FirstName != nil {
		conditions = append(conditions, "first_name=?")
		args = append(args, *params.FirstName)
	}

	if params.LastName != nil {
		conditions = append(conditions, "last_name=?")
		args = append(args, *params.LastName)
	}

	query := "SELECT * FROM User"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	stmt, err := s.DB.Prepare(query)
	defer CloseStmt(stmt, s.Logger)

	if err != nil {
		errMsg := PrepareStmtFail
		s.Logger.Error(errMsg, zap.Error(err))
		sendError(w, http.StatusInternalServerError, errMsg)
		return
	}
	rows, err := stmt.Query(args...)
	if err != nil {
		s.Logger.Error(QueryDBFail, zap.Error(err))
		sendError(w, http.StatusInternalServerError, QueryDBFail)
		return
	}

	defer func() {
		if err = rows.Close(); err != nil {
			s.Logger.Error("failed to close rows", zap.Error(err))
		}
	}()

	users := Users{}
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName)
		if err != nil {
			s.Logger.Error("failed to scan row", zap.Error(err))
			sendError(w, http.StatusInternalServerError, "failed to process user data")
			return
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		s.Logger.Error("sql row error", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "could not retrieve users")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, users)
}

// (POST /users) Create a new user.
func (s Server) PostUsers(w http.ResponseWriter, r *http.Request) {
	var userBody PostUsersJSONRequestBody
	var err error
	defer func() {
		if err = r.Body.Close(); err != nil {
			s.Logger.Warn("could not close request body", zap.Error(err))
		}
	}()
	// Ignore err returned because that would be caught by the middleware
	if err = json.NewDecoder(r.Body).Decode(&userBody); err != nil {
		errMsg := "failed to unmarshal request body correctly"
		s.Logger.Error(errMsg, zap.Object("body", userBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}

	var stmt *sql.Stmt
	if stmt, err = s.DB.Prepare("INSERT INTO User (email, first_name, last_name) VALUES (?, ?, ?)"); err != nil {
		errMsg := PrepareStmtFail
		s.Logger.Error(errMsg, zap.Object("reqBody", userBody), zap.Error(err))
		sendError(w, http.StatusInternalServerError, errMsg)
		return
	}
	defer CloseStmt(stmt, s.Logger)

	var res sql.Result
	if res, err = stmt.Exec(userBody.Email, userBody.FirstName, userBody.LastName); err != nil {
		errMsg := "failed to insert user"
		s.Logger.Error(errMsg, zap.Object("reqBody", userBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}
	var rows int64
	if rows, err = res.RowsAffected(); err != nil {
		s.Logger.Error("cant get rows affected", zap.Object("body", userBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, "user not inserted into db")
		return
	}
	if rows != 1 {
		s.Logger.Error("rows affected was not equal to 1",
			zap.Int64("rowsAffected", rows), zap.Object("body", userBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, "user not inserted into db")
		return
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		s.Logger.Error("cant get last insert id", zap.Object("body", userBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, "cant get last insert id")
		return
	}
	user := User{
		Id:        int(id),
		Email:     userBody.Email,
		FirstName: userBody.FirstName,
		LastName:  userBody.LastName,
	}

	SetHeaderAndWriteResponse(w, http.StatusCreated, user)
}

// (DELETE /users/{userID}) Delete a user by id.
// nolint: dupl //Duplicate of DeleteTeamsTeamID but this is more readable
func (s Server) DeleteUsersUserID(w http.ResponseWriter, _ *http.Request, userID int) {
	var err error
	var stmt *sql.Stmt
	if stmt, err = s.DB.Prepare("DELETE FROM User WHERE id=?"); err != nil {
		errMsg := PrepareStmtFail
		s.Logger.Error(errMsg, zap.Int("userID", userID), zap.Error(err))
		sendError(w, http.StatusInternalServerError, errMsg)
		return
	}
	defer CloseStmt(stmt, s.Logger)

	var res sql.Result
	if res, err = stmt.Exec(userID); err != nil {
		errMsg := "failed to delete user"
		s.Logger.Error(errMsg, zap.Int("userID", userID), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}

	var rows int64
	if rows, err = res.RowsAffected(); err != nil {
		s.Logger.Error("cant get rows affected", zap.Int("userID", userID), zap.Error(err))
		sendError(w, http.StatusBadRequest, "user not deleted from db")
		return
	}
	if rows != 1 {
		s.Logger.Error("rows affected was not equal to 1 after trying deleting user",
			zap.Int64("rowsAffected", rows), zap.Int("userID", userID), zap.Error(err))
		sendError(w, http.StatusBadRequest, "user not deleted from db")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, "user deleted successfully")
}

// (GET /users/{userID}) Get a user by id.
func (s Server) GetUsersUserID(w http.ResponseWriter, _ *http.Request, userID int) {
	stmt, err := s.DB.Prepare("SELECT * FROM User WHERE id=?")
	defer CloseStmt(stmt, s.Logger)
	if err != nil {
		errMsg := PrepareStmtFail
		s.Logger.Error(errMsg, zap.Int("userID", userID), zap.Error(err))
		sendError(w, http.StatusInternalServerError, errMsg)
		return
	}
	var user User
	if err = stmt.QueryRow(userID).Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errMsg := "user doesn't exist"
			s.Logger.Error(errMsg, zap.Int("userID", userID), zap.Error(err))
			sendError(w, http.StatusNotFound, errMsg)
		} else {
			errMsg := "failed to get user"
			s.Logger.Error(errMsg, zap.Int("userID", userID), zap.Error(err))
			sendError(w, http.StatusBadRequest, errMsg)
		}
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, user)
}
