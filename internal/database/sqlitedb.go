package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func CreateSQLiteDatabase(dbFilePath string) *sql.DB {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbFilePath))
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

func GetOrCreateSQLiteTransaction(db *sql.DB, tx *sql.Tx) (bool, *sql.Tx, error) {
	isExternalTx := tx != nil

	var err error
	if !isExternalTx {
		tx, err = db.Begin()

		if err != nil {
			return false, nil, err
		}
	}

	return isExternalTx, tx, nil
}

func CommitOrRollbackSQLiteTransaction(tx *sql.Tx) error {
	err := tx.Commit()
	if err != nil {
		tx.Rollback()

		return err
	}

	return nil
}
