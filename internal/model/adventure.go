package model

import (
	"database/sql"
	"errors"
)

type Adventure struct {
	AthleteId           int64
	StartLocation       int
	EndLocation         int
	CurrentLocationName string
	CurrentDistance     float32
	TotalDistance       float32
	Completed           int
	StartDate           int
	EndDate             int
}

func (adventure *Adventure) Load(athlId int64, startLocation int, endLocation int, db *sql.DB, tx *sql.Tx) (bool, error) {
	query, params := PrepareQuery("SELECT * FROM Adventure", map[string]any{
		"athlete_id":     athlId,
		"start_location": startLocation,
		"end_location":   endLocation,
	})

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, params...)
	} else {
		row = db.QueryRow(query, params...)
	}

	err := row.Scan(&adventure.AthleteId, &adventure.StartLocation, &adventure.EndLocation, &adventure.CurrentLocationName, &adventure.CurrentDistance, &adventure.TotalDistance, &adventure.Completed, &adventure.StartDate, &adventure.EndDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

func (adv *Adventure) Save(db *sql.DB, tx *sql.Tx) error {
	var err error

	var found bool
	found, err = AdventureExists(adv.AthleteId, adv.StartLocation, adv.EndLocation, db, tx)
	if err != nil {
		return err
	}

	if found {
		query := "UPDATE Adventure SET current_location_name=?, current_distance=?, total_distance=?, completed=?, start_date=?, end_date=? WHERE athlete_id=? AND start_location=? AND end_location=?"

		if tx != nil {
			_, err = tx.Exec(query, adv.CurrentLocationName, adv.CurrentDistance, adv.TotalDistance, adv.Completed, adv.StartDate, adv.EndDate, adv.AthleteId, adv.StartLocation, adv.EndLocation)
		} else {
			_, err = db.Exec(query, adv.CurrentLocationName, adv.CurrentDistance, adv.TotalDistance, adv.Completed, adv.StartDate, adv.EndDate, adv.AthleteId, adv.StartLocation, adv.EndLocation)
		}
	} else {
		query := "INSERT INTO Adventure VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)"

		if tx != nil {
			_, err = tx.Exec(query, adv.AthleteId, adv.StartLocation, adv.EndLocation, adv.CurrentLocationName, adv.CurrentDistance, adv.TotalDistance, adv.Completed, adv.StartDate, adv.EndDate)
		} else {
			_, err = db.Exec(query, adv.AthleteId, adv.StartLocation, adv.EndLocation, adv.CurrentLocationName, adv.CurrentDistance, adv.TotalDistance, adv.Completed, adv.StartDate, adv.EndDate)
		}
	}

	return err
}

func AdventureExists(athlId int64, startLocation int, endLocation int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp Adventure

	return temp.Load(athlId, startLocation, endLocation, db, tx)
}

func AllAdventures(db *sql.DB, tx *sql.Tx, filter map[string]any) ([]Adventure, error) {
	var err error

	var rows *sql.Rows
	query, params := PrepareQuery("SELECT * FROM Adventure", filter)
	if tx != nil {
		rows, err = tx.Query(query, params...)
	} else {
		rows, err = db.Query(query, params...)
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var adventures []Adventure
	for rows.Next() {
		adventures = append(adventures, Adventure{})

		adventureToEdit := &adventures[len(adventures)-1]
		if err = rows.Scan(&adventureToEdit.AthleteId, &adventureToEdit.StartLocation, &adventureToEdit.EndLocation,
			&adventureToEdit.CurrentLocationName, &adventureToEdit.CurrentDistance, &adventureToEdit.TotalDistance,
			&adventureToEdit.Completed, &adventureToEdit.StartDate, &adventureToEdit.EndDate); err != nil {
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return adventures, nil
}
