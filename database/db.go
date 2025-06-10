package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents a database connection. It is not safe for asynchronous
// write accesses. Use the BeginTx, Commit, Rollback and Add* Methods in
// series only.
type DB struct {
	d      *sql.DB
	txLock *sync.Mutex
	tx     *sql.Tx
}

// NewDB creates a new database by initializing or updating the schema
// and returns the ready to use database.
func NewDB(dbFile string) (DB, error) {
	rawDB, err := sql.Open("sqlite3", dbFile)
	db := DB{
		d:      rawDB,
		txLock: &sync.Mutex{},
	}
	if err != nil {
		return db, fmt.Errorf("could not open database: %s", err)
	}
	if _, err = db.d.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return db, fmt.Errorf("could not set pragma for foreign keys: %s", err)
	}
	if err = db.updateSchema(); err != nil {
		return db, fmt.Errorf("could not update schema: %s", err)
	}
	return db, nil
}

func (db DB) updateSchema() error {
	version, err := db.schemaVersion()
	if err != nil {
		return fmt.Errorf("could not find database version: %s", err)
	}
	tx, err := db.d.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction for schema updates: %s", err)
	}
	switch version {
	case 0:
		if _, err = tx.Exec(schemaV1); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not update database schema to version 1: %s", err)
		}
		fallthrough
	case 1:
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("could not commit schema update transaction: %s", err)
		}
		return nil
	}
	_ = tx.Rollback()
	return fmt.Errorf("unknown database version '%d'", version)
}

func (db DB) schemaVersion() (int, error) {
	rows, err := db.d.Query(`PRAGMA user_version`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	version := 0
	versionFound := false
	for rows.Next() {
		if versionFound {
			return 0, fmt.Errorf("found multiple database schema versions")
		}
		if err := rows.Scan(&version); err != nil {
			return 0, err
		}
		versionFound = true
	}
	if !versionFound {
		return 0, fmt.Errorf("found no database schema version")
	}
	return version, nil
}
