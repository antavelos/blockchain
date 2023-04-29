package db

import (
	"encoding/json"
	"os"
	"sync"
)

type DB struct {
	Filename string
}

func NewDB(filename string) *DB {
	return &DB{Filename: filename}
}

func (db *DB) Load() ([]byte, error) {
	err := createIfNotExists(db.Filename)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(db.Filename)
}

func (db *DB) Save(data any) error {
	err := createIfNotExists(db.Filename)
	if err != nil {
		return err
	}

	dataBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil
	}

	return os.WriteFile(db.Filename, dataBytes, os.ModePerm)
}

func (db *DB) WithLock(processData func([]byte) (any, error)) error {
	m := sync.Mutex{}
	m.Lock()
	defer m.Unlock()

	data, err := db.Load()
	if err != nil {
		return err
	}

	processed, err := processData(data)
	if err != nil {
		return err
	}

	return db.Save(processed)
}

func createIfNotExists(filename string) error {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		file, err := os.OpenFile(filename, os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		file.Close()
	}

	return nil
}
