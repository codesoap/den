package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type queryable interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// TrackPath adds a path to be tracked. It will return an error if the
// path is already tracked.
func (db *DB) TrackPath(path string) error {
	path = strings.TrimSuffix(path, string(filepath.Separator))
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("could not find info: %s", err)
	}
	tx, err := db.d.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %s", err)
	}
	paths, err := trackedPaths(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not find previous paths: %s", err)
	}
	for _, oldPath := range paths {
		if oldPath == path {
			tx.Rollback()
			return fmt.Errorf("path is already tracked")
		} else if strings.HasPrefix(path, oldPath) {
			tx.Rollback()
			f := "path is already contained in tracked path '%s'"
			return fmt.Errorf(f, oldPath)
		} else if strings.HasPrefix(oldPath, path) {
			tx.Rollback()
			f := "path contains already tracked path '%s'"
			return fmt.Errorf(f, oldPath)
		}
	}
	_, err = tx.Exec(`INSERT INTO tracked_path (path) VALUES (?)`, path)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not track path: %s", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %s", err)
	}
	return nil
}

// TrackedPaths returns a list of all tracked paths.
func (db *DB) TrackedPaths() ([]string, error) {
	return trackedPaths(db.d)
}

// UntrackPath removes the given path from the tracked paths. It also
// removes all file entries that existed only because of this path.
func (db *DB) UntrackPath(path string) error {
	tx, err := db.d.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %s", err)
	}

	res, err := tx.Exec(`DELETE FROM tracked_path WHERE path = ?`, path)
	if err != nil {
		tx.Rollback()
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not determine success: %s", err)
	} else if n != 1 {
		tx.Rollback()
		return fmt.Errorf("untracked %d instead of 1 paths", n)
	}

	paths, err := trackedPaths(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not find previous paths: %s", err)
	}
	if len(paths) == 0 {
		if _, err := tx.Exec(`DELETE FROM file`); err != nil {
			tx.Rollback()
			return fmt.Errorf("could not delete old entries: %s", err)
		}
	} else {
		q := `DELETE FROM file WHERE ` +
			`path NOT GLOB ?` +
			strings.Repeat(" AND path NOT GLOB ?", len(paths)-1)
		args := make([]any, 0)
		for _, path := range paths {
			args = append(args, filepath.Join(path, "*"))
		}
		if _, err := tx.Exec(q, args...); err != nil {
			tx.Rollback()
			return fmt.Errorf("could not delete old entries: %s", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %s", err)
	}
	return nil
}

func trackedPaths(q queryable) ([]string, error) {
	rows, err := q.Query(`SELECT path FROM tracked_path`)
	if err != nil {
		return nil, fmt.Errorf("could not query tracked paths: %s", err)
	}
	paths := make([]string, 0)
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("could not read tracked path: %s", err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}
