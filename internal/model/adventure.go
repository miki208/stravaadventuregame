package model

import (
	"database/sql"
	"errors"

	"github.com/miki208/stravaadventuregame/internal/database"
)

type Adventure struct {
	AthleteId           int
	StartLocation       int
	EndLocation         int
	CurrentLocationName string
	CurrentDistance     float32
	TotalDistance       float32
	Completed           int
}

type AdventureCompletionFilter int

const (
	DontFilterCompletion AdventureCompletionFilter = iota
	FilterCompleted
	FilterNotCompleted
)

func GetAdventuresByAthlete(athlId int, complFilter AdventureCompletionFilter, db *sql.DB, tx *sql.Tx) ([]*Adventure, error) {
	var query string
	var queryParams []any
	var err error

	if complFilter == DontFilterCompletion {
		query = "SELECT * FROM Adventure"
	} else {
		query = "SELECT * FROM Adventure WHERE completed=?"

		if complFilter == FilterCompleted {
			queryParams = append(queryParams, 1)
		} else {
			queryParams = append(queryParams, 0)
		}
	}

	var result []*Adventure

	var rows *sql.Rows
	if tx != nil {
		rows, err = tx.Query(query, queryParams...)
	} else {
		rows, err = db.Query(query, queryParams...)
	}

	if err != nil {
		if tx != nil {
			tx.Rollback()
		}

		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		result = append(result, &Adventure{})

		adventure := result[len(result)-1]
		if err = rows.Scan(&adventure.AthleteId, &adventure.StartLocation, &adventure.EndLocation, &adventure.CurrentLocationName, &adventure.CurrentDistance, &adventure.TotalDistance, &adventure.Completed); err != nil {
			if tx != nil {
				tx.Rollback()
			}

			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		if tx != nil {
			tx.Rollback()
		}

		return nil, err
	}

	return result, nil
}

func (adventure *Adventure) Load(athlId int, startLocation int, endLocation int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var row *sql.Row
	query := "SELECT * FROM Adventure WHERE athlete_id=? AND start_location=? AND end_location=?"

	if tx != nil {
		row = tx.QueryRow(query, athlId, startLocation, endLocation)
	} else {
		row = db.QueryRow(query, athlId, startLocation, endLocation)
	}

	err := row.Scan(&adventure.AthleteId, &adventure.StartLocation, &adventure.EndLocation, &adventure.CurrentLocationName, &adventure.CurrentDistance, &adventure.TotalDistance, &adventure.Completed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		if tx != nil {
			tx.Rollback()
		}

		return false, err
	}

	return true, nil
}

func AdventureExists(athlId int, startLocation int, endLocation int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp Adventure

	exists, err := temp.Load(athlId, startLocation, endLocation, db, tx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (adv *Adventure) Save(db *sql.DB, tx *sql.Tx) error {
	isExternalTx, tx, err := database.GetOrCreateSQLiteTransaction(db, tx)
	if err != nil {
		return err
	}

	exists, err := AdventureExists(adv.AthleteId, adv.StartLocation, adv.EndLocation, db, tx)
	if err != nil {
		return err
	}

	if exists {
		_, err = tx.Exec("UPDATE Adventure SET current_location_name=?, current_distance=?, total_distance=?, completed=? WHERE athlete_id=? AND start_location=? AND end_location=?", adv.CurrentLocationName, adv.CurrentDistance, adv.TotalDistance, adv.Completed, adv.AthleteId, adv.StartLocation, adv.EndLocation)
		if err != nil {
			tx.Rollback()

			return err
		}
	} else {
		_, err = tx.Exec("INSERT INTO Adventure VALUES(?, ?, ?, ?, ?, ?, ?)", adv.AthleteId, adv.StartLocation, adv.EndLocation, adv.CurrentLocationName, adv.CurrentDistance, adv.TotalDistance, adv.Completed)
		if err != nil {
			tx.Rollback()

			return err
		}
	}

	if !isExternalTx {
		err = database.CommitOrRollbackSQLiteTransaction(tx)
		if err != nil {
			return err
		}
	}

	return nil
}
