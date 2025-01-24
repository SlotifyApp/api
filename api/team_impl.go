package api

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// (GET /teams) Get a team by query params.
func (s Server) GetTeams(w http.ResponseWriter, _ *http.Request, params GetTeamsParams) {
	var args []interface{}

	query := "SELECT * FROM Team"

	if params.Name != nil {
		query += " WHERE name=?"
		args = append(args, *params.Name)
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
		err = rows.Close()
		if err != nil {
			s.Logger.Error("failed to close rows", zap.Error(err))
		}
	}()

	teams := Teams{}
	for rows.Next() {
		var team Team
		err = rows.Scan(&team.Id, &team.Name)
		if err != nil {
			s.Logger.Error("failed to scan row", zap.Error(err))
			sendError(w, http.StatusInternalServerError, "failed to process team data")
			return
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		s.Logger.Error("sql row error", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "could not retrieve teams")
		return
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, teams)
}

// (POST /teams) Create a new team.
func (s Server) PostTeams(w http.ResponseWriter, r *http.Request) {
	var teamBody PostTeamsJSONRequestBody
	// Ignore err returned because that would be caught by the middleware
	if err := json.NewDecoder(r.Body).Decode(&teamBody); err != nil {
		errMsg := "failed to unmarshal request body correctly"
		s.Logger.Error(errMsg, zap.Object("body", teamBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}
	team, err := s.TeamService.InsertTeam(teamBody)
	if err != nil {
		s.Logger.Error("failed to create team", zap.Object("body", teamBody), zap.Error(err))
		sendError(w, http.StatusBadRequest, "team creation unsuccessful")
		return
	}
	SetHeaderAndWriteResponse(w, http.StatusCreated, team)
}

// (DELETE /teams/{teamID}) Delete a team by id.
// nolint: dupl //Duplicate of DeleteUsersUserID but this is more readable
func (s Server) DeleteTeamsTeamID(w http.ResponseWriter, _ *http.Request, teamID int) {
	stmt, err := s.DB.Prepare("DELETE FROM Team WHERE id=?")
	defer CloseStmt(stmt, s.Logger)

	if err != nil {
		errMsg := PrepareStmtFail
		s.Logger.Error(errMsg, zap.Int("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusInternalServerError, errMsg)
		return
	}
	res, err := stmt.Exec(teamID)
	if err != nil {
		errMsg := "failed to delete team"
		s.Logger.Error(errMsg, zap.Int("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}

	rows, err := res.RowsAffected()
	if err != nil {
		s.Logger.Error("cant get rows affected", zap.Int("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusBadRequest, "team not deleted from db")
		return
	}

	if rows != 1 {
		s.Logger.Error("rows affected was not equal to 1 after trying deleting team",
			zap.Int64("rowsAffected", rows), zap.Int("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusBadRequest, "team not deleted from db")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode("team deleted successfully")
}

// (GET /teams/{teamID}) Get a team by id.
func (s Server) GetTeamsTeamID(w http.ResponseWriter, _ *http.Request, teamID int) {
	stmt, err := s.DB.Prepare("SELECT * FROM Team WHERE id=?")
	defer CloseStmt(stmt, s.Logger)

	if err != nil {
		errMsg := PrepareStmtFail
		s.Logger.Error(errMsg, zap.Int("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusInternalServerError, errMsg)
		return
	}
	var team Team
	if err = stmt.QueryRow(teamID).Scan(&team); err != nil {
		errMsg := "failed to get team"
		s.Logger.Error(errMsg, zap.Int("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(team)
}

// (GET /teams/{teamID}/users) Get all members of a team.
func (s Server) GetTeamsTeamIDUsers(w http.ResponseWriter, _ *http.Request, teamID int) {
	query := "SELECT u.* FROM Team t JOIN UserToTeam utt ON t.id=utt.team_id JOIN User u ON u.id=utt.user_id WHERE t.id=?"
	stmt, err := s.DB.Prepare(query)
	if err != nil {
		errMsg := PrepareStmtFail
		s.Logger.Error(errMsg, zap.Int("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusInternalServerError, errMsg)
		return
	}
	defer CloseStmt(stmt, s.Logger)

	rows, err := stmt.Query(teamID)
	if err != nil {
		errMsg := "failed to query db"
		s.Logger.Error(errMsg, zap.Int("teamID", teamID), zap.Error(err))
		sendError(w, http.StatusInternalServerError, errMsg)
		return
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			s.Logger.Error("failed to close rows", zap.Error(err))
		}
	}()

	var users Users
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

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(users); err != nil {
		s.Logger.Error("Failed to encode JSON", zap.Array("users", users))
		sendError(w, http.StatusInternalServerError, "Failed to encode JSON")
		return
	}
}

// (POST /teams/{teamID}/users/{userID}) Add a user to a team.
func (s Server) PostTeamsTeamIDUsersUserID(w http.ResponseWriter, _ *http.Request, teamID int, userID int) {
	stmt, err := s.DB.Prepare("INSERT INTO UserToTeam (user_id, team_id) VALUES (?, ?)")
	defer CloseStmt(stmt, s.Logger)

	if err != nil {
		errMsg := PrepareStmtFail
		s.Logger.Error(errMsg, zap.Int("teamID", teamID), zap.Int("userID", userID), zap.Error(err))
		sendError(w, http.StatusInternalServerError, errMsg)
		return
	}
	res, err := stmt.Exec(userID, teamID)
	if err != nil {
		errMsg := "failed to add user to team"
		s.Logger.Error(errMsg, zap.Int("teamID", teamID), zap.Int("userID", userID), zap.Error(err))
		sendError(w, http.StatusBadRequest, errMsg)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		s.Logger.Error("cant get rows affected",
			zap.Int("teamID", teamID), zap.Int("userID", userID), zap.Error(err))
		sendError(w, http.StatusBadRequest, "team not inserted into db")
		return
	}
	if rows != 1 {
		s.Logger.Error("rows affected was not equal to 1",
			zap.Int("teamID", teamID), zap.Int("userID", userID), zap.Error(err))
		sendError(w, http.StatusBadRequest, "team not inserted into db")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode("User successfully joined the team.")
}
