package model

import (
	"database/sql"
	"errors"
)

type Location struct {
	Id   int
	Lat  float64
	Lon  float64
	Name string
}

func GetLocations(db *sql.DB, tx *sql.Tx) ([]*Location, error) {
	query := "SELECT * FROM Location"

	var result []*Location

	var err error
	var rows *sql.Rows
	if tx != nil {
		rows, err = tx.Query(query)
	} else {
		rows, err = db.Query(query)
	}

	if err != nil {
		if tx != nil {
			tx.Rollback()
		}

		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		result = append(result, &Location{})

		location := result[len(result)-1]
		if err = rows.Scan(&location.Id, &location.Lat, &location.Lon, &location.Name); err != nil {
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

func (location *Location) LoadById(id int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var row *sql.Row
	query := "SELECT * FROM Location WHERE id=?"

	if tx != nil {
		row = tx.QueryRow(query, id)
	} else {
		row = db.QueryRow(query, id)
	}

	err := row.Scan(&location.Id, &location.Lat, &location.Lon, &location.Name)
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

func LocationExists(id int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp Location

	exists, err := temp.LoadById(id, db, tx)
	if err != nil {
		return false, err
	}

	return exists, nil
}
