package helper

import (
	"database/sql"

	"github.com/miki208/stravaadventuregame/internal/model"
)

func IsAthleteAdmin(athleteId int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var athleteSettings model.AthleteSettings

	found, err := athleteSettings.Load(athleteId, db, tx)
	if err != nil {
		return false, err
	}

	if !found {
		return false, nil
	}

	return athleteSettings.IsAdmin == 1, nil
}
