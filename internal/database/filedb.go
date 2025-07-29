package database

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"sync"
)

type FileDatabase struct {
	lock        sync.Mutex
	dataDirPath string
}

func CreateFileDatabase(dataDirPath string) *FileDatabase {
	slog.Info("Initializing file database...", "dataDirPath", dataDirPath)

	return &FileDatabase{dataDirPath: dataDirPath}
}

func (db *FileDatabase) Read(level1Name, level2Name string, data any) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.readNoLock(level1Name, level2Name, data)
}

func (db *FileDatabase) readNoLock(level1Name, level2Name string, data any) error {
	fullDirPath := db.dataDirPath + level1Name + "/"

	file, err := os.Open(fullDirPath + level2Name + ".json")
	if err != nil {
		return err
	}

	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &data)
	if err != nil {
		return err
	}

	return nil
}

func (db *FileDatabase) Write(level1Name, level2Name string, data any) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.writeNoLock(level1Name, level2Name, data)
}

func (db *FileDatabase) writeNoLock(level1Name, level2Name string, data any) error {
	fullDirPath := db.dataDirPath + level1Name + "/"

	content, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	err = os.MkdirAll(fullDirPath, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(fullDirPath + level2Name + ".json")
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func (db *FileDatabase) Exists(level1Name, level2Name string) (bool, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.existsNoLock(level1Name, level2Name)
}

func (db *FileDatabase) existsNoLock(level1Name, level2Name string) (bool, error) {
	fullDirPath := db.dataDirPath + level1Name + "/"

	_, err := os.Stat(fullDirPath + level2Name + ".json")
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (db *FileDatabase) Delete(level1Name, level2Name string) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.deleteNoLock(level1Name, level2Name)
}

func (db *FileDatabase) deleteNoLock(level1Name, level2Name string) error {
	fullDirPath := db.dataDirPath + level1Name + "/"

	exists, err := db.existsNoLock(level1Name, level2Name)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	err = os.Remove(fullDirPath + level2Name + ".json")
	if err != nil {
		return err
	}

	return nil
}

func (db *FileDatabase) DeleteAll(level1Name string) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.deleteAllNoLock(level1Name)
}

func (db *FileDatabase) deleteAllNoLock(level1Name string) error {
	fullDirPath := db.dataDirPath + level1Name + "/"

	return os.RemoveAll(fullDirPath)
}
