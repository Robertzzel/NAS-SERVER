package DatabaseService

import (
	"NAS-Server-Web/shared/configurations"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

type DatabaseService struct {
	*sql.DB
}

var instance *DatabaseService = nil

func NewDatabaseService() (DatabaseService, error) {
	if instance == nil {
		db, err := sql.Open("sqlite3", configurations.GetDatabasePath())
		if err != nil {
			return DatabaseService{}, err
		}

		dm := DatabaseService{db}
		if err = dm.migrateDatabase(); err != nil {
			return DatabaseService{}, err
		}

		instance = &dm
	}

	return DatabaseService{}, nil
}

func (db *DatabaseService) UsernameAndPasswordExists(username, password string) (bool, error) {
	var cnt int
	err := db.QueryRow(`select count(*) from User where Name = ? and Password = ? LIMIT 1`, username, hash(password)).Scan(&cnt)
	if err != nil {
		return false, errors.New("database problem")
	}
	return cnt != 0, nil
}

func (db *DatabaseService) GetUserAllocatedMemory(username string) (int, error) {
	var memory int
	err := db.QueryRow(`select AllocatedMemory from User where Name = ? LIMIT 1`, username).Scan(&memory)
	if err != nil {
		return 0, err
	}
	return memory, nil
}

func (db *DatabaseService) Close() {
	_ = db.DB.Close()
}

func (db *DatabaseService) AddUser(username, password string, memory int) error {
	_, err := db.Exec(`INSERT INTO User (Name, Password, AllocatedMemory) VALUES (?, ?, ?)`, username, hash(password), memory)
	return err
}

func (db *DatabaseService) migrateDatabase() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS User(
    	Id integer PRIMARY KEY,
		Name varchar(255) UNIQUE NOT NULL,
		Password varchar(255) NOT NULL,
    	AllocatedMemory integer NOT NULL
    )`)
	if err != nil {
		return err
	}
	return nil
}

func hash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}
