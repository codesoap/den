package den

import (
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"time"

	"github.com/codesoap/den/database"
)

// Rescan updates the database to contain up to date information
// about files of all tracked paths.
func Rescan(db database.DB, progress chan Progress) error {
	defer close(progress)
	paths, err := db.TrackedPaths()
	if err != nil {
		return fmt.Errorf("could not query tracked paths: %s", err)
	}
	if err := db.BeginTx(); err != nil {
		return err
	}
	toReindex := make(map[string]fs.DirEntry)
	for i := range paths {
		todo, err := rescanPath(db, paths[i])
		if err != nil {
			_ = db.Rollback()
			return fmt.Errorf("could not rescan path '%s': %s", paths[i], err)
		}
		maps.Copy(toReindex, todo)
	}
	if err := db.Commit(); err != nil {
		// The transaction of checking the database and deleting old entries
		// is separated from the (re-)index transactions, so that no
		// transaction becomes too large. If something goes wrong during a
		// (re-)indexing transaction, that should cause no trouble.
		return fmt.Errorf("could not commit transaction: %s", err)
	}

	entriesInTx := 0
	var lastProgressUpdate time.Time
	prog := Progress{Total: len(toReindex)}
	for path, entry := range toReindex {
		if time.Since(lastProgressUpdate) >= time.Second {
			progress <- prog
			lastProgressUpdate = time.Now()
		}
		if entriesInTx == 0 {
			if err := db.BeginTx(); err != nil {
				return fmt.Errorf("could not start transaction: %s", err)
			}
		}
		entriesInTx++
		if err = indexFile(db, path, entry); err != nil {
			_ = db.Rollback()
			return fmt.Errorf("could not index file '%s': %s", path, err)
		}
		if entriesInTx == 1_000 {
			entriesInTx = 0
			if err2 := db.Commit(); err2 != nil {
				return fmt.Errorf("could not commit transaction: %s", err2)
			}
		}
		prog.Done++
	}
	if entriesInTx > 0 {
		if err := db.Commit(); err != nil {
			return fmt.Errorf("could not commit transaction: %s", err)
		}
	}
	prog.Done = prog.Total
	progress <- prog
	return nil
}

// rescanPath compares all stored files below the given path with the
// files found on the filesystem. Removed files are removed from the
// database and files that should be newly indexed or re-indexed are
// returned.
func rescanPath(db database.DB, path string) (map[string]fs.DirEntry, error) {
	// TODO: Use temporary SQLite table to use less memory:
	oldInfos, err := db.AllFileInfosAt(path)
	if err != nil {
		return nil, err
	}
	newInfos := make(map[string]fs.DirEntry)
	err = filepath.WalkDir(path,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			hidden, err := isHiddenFile(d.Name())
			if err != nil {
				f := "could not determine if '%s' is hidden: %s"
				return fmt.Errorf(f, path, err)
			} else if hidden {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return nil
			}
			// TODO: Use temporary SQLite table to use less memory:
			newInfos[path] = d
			return nil
		})
	if err != nil {
		return nil, err
	}

	var deletePaths []string
	toIndex := make(map[string]fs.DirEntry)
	for path, oldInfo := range oldInfos {
		if newInfo, ok := newInfos[path]; !ok {
			deletePaths = append(deletePaths, path)
		} else {
			info, err := newInfo.Info()
			if err != nil {
				f := "could not get info for file '%s': %s"
				return nil, fmt.Errorf(f, path, err)
			}
			if oldInfo.Size != info.Size() ||
				oldInfo.ModTime != info.ModTime().Truncate(time.Second) {
				// Deleting and adding anew reindexes the file:
				deletePaths = append(deletePaths, path)
				toIndex[path] = newInfo
			}
		}
	}
	for path, newInfo := range newInfos {
		if _, ok := oldInfos[path]; !ok {
			toIndex[path] = newInfo
		}
	}
	if err = db.DeletePaths(deletePaths); err != nil {
		return nil, fmt.Errorf("could not delete info on files: %s", err)
	}
	return toIndex, nil
}
