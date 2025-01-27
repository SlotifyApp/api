package api

import (
	"database/sql"
	"fmt"

	"github.com/SlotifyApp/slotify-backend/database"
	"go.uber.org/zap"
)

type TeamRepositoryInterface interface {
	AddTeam(TeamCreate) (Team, error)
	// AddUserToTeam takes in a teamID and a userID.
	AddUserToTeam(int, int) error
	CheckTeamExistsByID(int) (bool, error)
	DeleteTeamByID(int) error
	GetAllTeamMembers(int) (Users, error)
	GetTeamByID(int) (Team, error)
	GetTeamsByQueryParams(GetTeamsParams) (Teams, error)
}

type TeamRepository struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewTeamRepository(logger *zap.SugaredLogger, db *sql.DB) TeamRepository {
	return TeamRepository{
		logger: logger,
		db:     db,
	}
}

// check TeamRepository conforms to the interface.
var _ TeamRepositoryInterface = (*TeamRepository)(nil)

func (tr TeamRepository) AddUserToTeam(teamID int, userID int) error {
	// TODO: What error is returned when teamID or userID don't exist? Make this a 400, and the rest 500
	stmt, err := tr.db.Prepare("INSERT INTO UserToTeam (user_id, team_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("team repository failed to add team member: %s: %w", PrepareStmtFail, err)
	}
	defer database.CloseStmt(stmt, tr.logger)
	res, err := stmt.Exec(userID, teamID)
	if err != nil {
		return fmt.Errorf("team repository failed to execute insert stmt: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("team repository failed to get rows affected: %w", err)
	}

	if rows != 1 {
		return fmt.Errorf("team repository: %w", database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: 1})
	}

	return nil
}

func (tr TeamRepository) CheckTeamExistsByID(teamID int) (bool, error) {
	var exists bool
	if err := tr.db.QueryRow("SELECT EXISTS(SELECT 1 FROM Team WHERE id=?)", teamID).Scan(&exists); err != nil {
		return false, fmt.Errorf("team repository: error checking team existence: %w", err)
	}
	return exists, nil
}

func (tr TeamRepository) GetAllTeamMembers(teamID int) (Users, error) {
	// Check team exists first, the below join won't throw a FK error
	exists, err := tr.CheckTeamExistsByID(teamID)
	if err != nil {
		return Users{}, fmt.Errorf("team repository: %w", err)
	}
	if !exists {
		return Users{}, fmt.Errorf("team repository: %w: ", database.ErrTeamIDInvalid)
	}

	query := "SELECT u.* FROM Team t JOIN UserToTeam utt ON t.id=utt.team_id JOIN User u ON u.id=utt.user_id WHERE t.id=?"
	stmt, err := tr.db.Prepare(query)
	if err != nil {
		return Users{}, fmt.Errorf("team repository failed to get all team members: %s: %w", PrepareStmtFail, err)
	}
	defer database.CloseStmt(stmt, tr.logger)

	rows, err := stmt.Query(teamID)
	if err != nil {
		return Users{}, fmt.Errorf("team repository failed to execute query: %w", err)
	}

	defer database.CloseRows(rows, tr.logger)

	users := Users{}
	for rows.Next() {
		var user User
		if err = rows.Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName); err != nil {
			return Users{}, fmt.Errorf("team repository failed to scan rows: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return Users{}, fmt.Errorf("team repository sql row failure: %w", err)
	}
	return users, nil
}

func (tr TeamRepository) GetTeamByID(teamID int) (Team, error) {
	stmt, err := tr.db.Prepare("SELECT * FROM Team WHERE id=?")
	if err != nil {
		return Team{}, fmt.Errorf("team repository failed to get team by id: %s: %w", PrepareStmtFail, err)
	}
	defer database.CloseStmt(stmt, tr.logger)
	var team Team
	if err = stmt.QueryRow(teamID).Scan(&team.Id, &team.Name); err != nil {
		return Team{}, fmt.Errorf("team repository failed to get team by id, query row error, %w", err)
	}
	return team, nil
}

func (tr TeamRepository) DeleteTeamByID(teamID int) error {
	stmt, err := tr.db.Prepare("DELETE FROM Team WHERE id=?")
	if err != nil {
		return fmt.Errorf("team repository failed to delete team: %s: %w", PrepareStmtFail, err)
	}

	defer database.CloseStmt(stmt, tr.logger)
	res, err := stmt.Exec(teamID)
	if err != nil {
		return fmt.Errorf("team repository failed to delete team %d: %w", teamID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("team repository failed to get rows affected %d: %w", teamID, err)
	}

	if rows != 1 {
		return fmt.Errorf("team repository: %w", database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: 1})
	}
	return nil
}

// GetTeamsByQueryParams returns a list of teams matching the given
// query parameters.
func (tr TeamRepository) GetTeamsByQueryParams(params GetTeamsParams) (Teams, error) {
	var args []any

	query := "SELECT * FROM Team"

	if params.Name != nil {
		query += " WHERE name=?"
		args = append(args, *params.Name)
	}

	stmt, err := tr.db.Prepare(query)
	if err != nil {
		return Teams{}, fmt.Errorf("%s: %w", PrepareStmtFail, err)
	}
	defer database.CloseStmt(stmt, tr.logger)

	var rows *sql.Rows
	if rows, err = stmt.Query(args...); err != nil {
		return Teams{}, fmt.Errorf("team repository: %s: %w", QueryDBFail, err)
	}

	defer database.CloseRows(rows, tr.logger)

	teams := Teams{}
	for rows.Next() {
		var team Team
		if err = rows.Scan(&team.Id, &team.Name); err != nil {
			return Teams{}, fmt.Errorf("team repository: failed to scan row: %w", err)
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return Teams{}, fmt.Errorf("team repository: sql rows error: %w", err)
	}
	return teams, nil
}

// AddTeam will insert a team into the Team table.
func (tr TeamRepository) AddTeam(teamCreate TeamCreate) (Team, error) {
	stmt, err := tr.db.Prepare("INSERT INTO Team (name) VALUES (?)")
	if err != nil {
		return Team{}, err
	}
	defer database.CloseStmt(stmt, tr.logger)
	var res sql.Result
	if res, err = stmt.Exec(teamCreate.Name); err != nil {
		return Team{}, err
	}
	var rows int64
	if rows, err = res.RowsAffected(); err != nil {
		return Team{}, fmt.Errorf("team repository: %w", err)
	}
	if rows != 1 {
		return Team{}, fmt.Errorf("team repository: %w", database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: 1})
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return Team{}, fmt.Errorf("team repository: %w", err)
	}
	team := Team{
		Id:   int(id),
		Name: teamCreate.Name,
	}
	return team, nil
}
