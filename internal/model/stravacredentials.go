package model

import (
	"database/sql"
	"errors"

	"github.com/miki208/stravaadventuregame/internal/database"
)

type StravaCredential struct {
	AthleteId    int
	AccessToken  string
	RefreshToken string
	ExpiresAt    int
}

func (cred *StravaCredential) LoadByAthleteId(athlId int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var row *sql.Row
	query := "SELECT * FROM StravaCredentials WHERE athlete_id=?"

	if tx != nil {
		row = tx.QueryRow(query, athlId)
	} else {
		row = db.QueryRow(query, athlId)
	}

	err := row.Scan(&cred.AthleteId, &cred.AccessToken, &cred.RefreshToken, &cred.ExpiresAt)
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

func StravaCredentialForAthleteIdExists(athlId int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp StravaCredential

	exists, err := temp.LoadByAthleteId(athlId, db, tx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (cred *StravaCredential) Save(db *sql.DB, tx *sql.Tx) error {
	isExternalTx, tx, err := database.GetOrCreateSQLiteTransaction(db, tx)
	if err != nil {
		return err
	}

	exists, err := StravaCredentialForAthleteIdExists(cred.AthleteId, db, tx)
	if err != nil {
		return err
	}

	if exists {
		_, err := tx.Exec("UPDATE StravaCredentials SET access_token=?, refresh_token=?, expires_at=? WHERE athlete_id=?", cred.AccessToken, cred.RefreshToken, cred.ExpiresAt, cred.AthleteId)
		if err != nil {
			tx.Rollback()

			return err
		}
	} else {
		_, err := tx.Exec("INSERT INTO StravaCredentials VALUES(?, ?, ?, ?)", cred.AthleteId, cred.AccessToken, cred.RefreshToken, cred.ExpiresAt)
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
