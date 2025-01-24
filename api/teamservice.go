package api

import (
	"database/sql"
	"fmt"

	"github.com/SlotifyApp/slotify-backend/database"
	"go.uber.org/zap"
)

type TeamServiceInterface interface {
	InsertTeam(TeamCreate) (Team, error)
}

type TeamService struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewTeamService(logger *zap.SugaredLogger, db *sql.DB) TeamService {
	return TeamService{
		logger: logger,
		db:     db,
	}
}

// check TeamService conforms to the interface.
var _ TeamServiceInterface = (*TeamService)(nil)

// InsertTeam will insert a team into the Team table.
func (t TeamService) InsertTeam(teamCreate TeamCreate) (Team, error) {
	stmt, err := t.db.Prepare("INSERT INTO Team (name) VALUES (?)")
	if err != nil {
		return Team{}, err
	}
	defer CloseStmt(stmt, t.logger)
	var res sql.Result
	if res, err = stmt.Exec(teamCreate.Name); err != nil {
		return Team{}, err
	}
	var rows int64
	if rows, err = res.RowsAffected(); err != nil {
		return Team{}, fmt.Errorf("teamsservice database: %w", err)
	}
	if rows != 1 {
		return Team{}, fmt.Errorf("teamsservice database: %w", database.ErrWrongNumberRows)
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return Team{}, fmt.Errorf("teamsservice database: %w", err)
	}
	team := Team{
		Id:   int(id),
		Name: teamCreate.Name,
	}
	return team, nil
}
