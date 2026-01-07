package database

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type ShortFileInfo struct {
	Size    int64
	ModTime time.Time
}

func (db DB) AllFileCount() (int, error) {
	row := db.d.QueryRow(`SELECT COUNT(*) FROM file`)
	var cnt int
	return cnt, row.Scan(&cnt)
}

// AllFileInfosAt returns a map with paths as keys, containing all
// stored paths below the given path with file info.
//
// db.Begin must have been called to create a transaction before calling
// AllFileInfosAt.
func (db DB) AllFileInfosAt(path string) (map[string]ShortFileInfo, error) {
	q := `SELECT path, size, modified FROM file WHERE path LIKE ?`
	rows, err := db.tx.Query(q, filepath.Join(path, "%"))
	if err != nil {
		return nil, fmt.Errorf("could not query database: %s", err)
	}
	infos := make(map[string]ShortFileInfo)
	for rows.Next() {
		var path string
		var unixModTime int64
		info := ShortFileInfo{}
		if err = rows.Scan(&path, &info.Size, &unixModTime); err != nil {
			return nil, fmt.Errorf("could not read from database: %s", err)
		}
		info.ModTime = time.Unix(unixModTime, 0)
		infos[path] = info
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("could not gather file infos: %s", rows.Err())
	}
	return infos, nil
}

// DeletePaths deletes all file entries for the given paths.
//
// db.Begin must have been called to create a transaction before calling
// DeletePaths.
func (db DB) DeletePaths(paths []string) error {
	// TODO: Use temporary table to improve performance and remove chunking.
	chunkSize := 1_000
	args := make([]any, 0, chunkSize)
	for i := 0; i < len(paths); i += chunkSize {
		args = args[:0]
		j := min(i+chunkSize, len(paths))
		for _, path := range paths[i:j] {
			args = append(args, path)
		}
		q := `DELETE FROM file WHERE path IN (?` + strings.Repeat(", ?", j-i-1) + `)`
		if _, err := db.tx.Exec(q, args...); err != nil {
			return err
		}
	}
	return nil
}
