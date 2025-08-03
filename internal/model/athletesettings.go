package model

import (
	"database/sql"
	"errors"
)

type AthleteSettings struct {
	AthleteId                     int64
	AutoUpdateActivityDescription int
	IsAdmin                       int
}

func (athleteSettings *AthleteSettings) Load(athlId int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	query, params := PrepareQuery("SELECT * FROM AthleteSettings", map[string]any{
		"athlete_id": athlId,
	})

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, params...)
	} else {
		row = db.QueryRow(query, params...)
	}

	err := row.Scan(&athleteSettings.AthleteId, &athleteSettings.AutoUpdateActivityDescription, &athleteSettings.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

func (athleteSettings *AthleteSettings) Save(db *sql.DB, tx *sql.Tx) error {
	var err error

	var found bool
	found, err = AthleteSettingsExists(athleteSettings.AthleteId, db, tx)
	if err != nil {
		return err
	}

	if found {
		query := "UPDATE AthleteSettings SET auto_update_activity_description=?, is_admin=? WHERE athlete_id=?"

		if tx != nil {
			_, err = tx.Exec(query, athleteSettings.AutoUpdateActivityDescription, athleteSettings.IsAdmin, athleteSettings.AthleteId)
		} else {
			_, err = db.Exec(query, athleteSettings.AutoUpdateActivityDescription, athleteSettings.IsAdmin, athleteSettings.AthleteId)
		}
	} else {
		query := "INSERT INTO AthleteSettings VALUES(?, ?, ?)"

		if tx != nil {
			_, err = tx.Exec(query, athleteSettings.AthleteId, athleteSettings.AutoUpdateActivityDescription, athleteSettings.IsAdmin)
		} else {
			_, err = db.Exec(query, athleteSettings.AthleteId, athleteSettings.AutoUpdateActivityDescription, athleteSettings.IsAdmin)
		}
	}

	return err
}

func AthleteSettingsExists(athlId int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp AthleteSettings

	return temp.Load(athlId, db, tx)
}

func AllAthleteSettings(db *sql.DB, tx *sql.Tx, filter map[string]any) ([]AthleteSettings, error) {
	var err error

	var rows *sql.Rows
	query, params := PrepareQuery("SELECT * FROM AthleteSettings", filter)
	if tx != nil {
		rows, err = tx.Query(query, params...)
	} else {
		rows, err = db.Query(query, params...)
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var listOfAthleteSettings []AthleteSettings
	for rows.Next() {
		listOfAthleteSettings = append(listOfAthleteSettings, AthleteSettings{})

		athleteSettingsToEdit := &listOfAthleteSettings[len(listOfAthleteSettings)-1]
		if err = rows.Scan(&athleteSettingsToEdit.AthleteId, &athleteSettingsToEdit.AutoUpdateActivityDescription, &athleteSettingsToEdit.IsAdmin); err != nil {
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return listOfAthleteSettings, nil
}
