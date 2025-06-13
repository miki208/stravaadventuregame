package model

import (
	"database/sql"
	"errors"

	"github.com/miki208/stravaadventuregame/internal/database"
)

type Athlete struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Sex       string `json:"sex"`
	isAdmin   int
}

func (athl *Athlete) IsAdmin() bool {
	return athl.isAdmin > 0
}

// LoadById returns true if there was a record in the database, otherwise returns false.
// To distinguish between no records and errors, when it returns false, check for errors.
func (athl *Athlete) LoadById(id int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var row *sql.Row
	query := "SELECT * FROM athlete WHERE id=?"

	if tx != nil {
		row = tx.QueryRow(query, id)
	} else {
		row = db.QueryRow(query, id)
	}

	err := row.Scan(&athl.Id, &athl.Username, &athl.FirstName, &athl.LastName, &athl.City, &athl.Country, &athl.Sex, &athl.isAdmin)
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

func AthleteExists(id int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp Athlete

	exists, err := temp.LoadById(id, db, tx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (athl *Athlete) Save(db *sql.DB, tx *sql.Tx) error {
	isExternalTx, tx, err := database.GetOrCreateSQLiteTransaction(db, tx)
	if err != nil {
		return err
	}

	exists, err := AthleteExists(athl.Id, db, tx)
	if err != nil {
		return err
	}

	if exists {
		_, err = tx.Exec("UPDATE athlete SET username=?, first_name=?, last_name=?, city=?, country=?, sex=? WHERE id=?", athl.Username, athl.FirstName, athl.LastName, athl.City, athl.Country, athl.Sex, athl.Id)
		if err != nil {
			tx.Rollback()

			return err
		}
	} else {
		_, err = tx.Exec("INSERT INTO athlete VALUES(?, ?, ?, ?, ?, ?, ?, ?)", athl.Id, athl.Username, athl.FirstName, athl.LastName, athl.City, athl.Country, athl.Sex, athl.isAdmin)
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

func (athl *Athlete) Delete(db *sql.DB, tx *sql.Tx) error {
	query := "DELETE FROM Athlete WHERE id=?"

	var err error
	if tx != nil {
		_, err = tx.Exec(query, athl.Id)
	} else {
		_, err = db.Exec(query, athl.Id)
	}

	if err != nil {
		if tx != nil {
			tx.Rollback()
		}

		return err
	}

	return nil
}

func IsAthleteAdmin(athlId int, db *sql.DB, tx *sql.Tx) (bool, error) {

	var athlete Athlete
	found, err := athlete.LoadById(athlId, db, tx)
	if err != nil {

		return false, err
	}

	if !found {
		return false, errors.New("athlete not found")
	}

	return athlete.IsAdmin(), nil
}
