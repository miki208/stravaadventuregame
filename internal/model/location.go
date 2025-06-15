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

func (location *Location) Load(id int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var err error

	query, params := PrepareQuery("SELECT * FROM Location", map[string]any{"id": id})

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, params...)
	} else {
		row = db.QueryRow(query, params...)
	}

	err = row.Scan(&location.Id, &location.Lat, &location.Lon, &location.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

func LocationExists(id int, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp Location

	return temp.Load(id, db, tx)
}

func AllLocations(db *sql.DB, tx *sql.Tx, filter map[string]any) ([]Location, error) {
	var err error

	var rows *sql.Rows
	query, params := PrepareQuery("SELECT * FROM Location", filter)
	if tx != nil {
		rows, err = tx.Query(query, params...)
	} else {
		rows, err = db.Query(query, params...)
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var locatioons []Location
	for rows.Next() {
		locatioons = append(locatioons, Location{})

		locationToEdit := &locatioons[len(locatioons)-1]
		if err = rows.Scan(&locationToEdit.Id, &locationToEdit.Lat, &locationToEdit.Lon, &locationToEdit.Name); err != nil {
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return locatioons, nil
}
