package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type FileFilter struct {
	CreatedSince, CreatedUntil *time.Time
	Prefix                     string
}

type PictureFilter struct {
	FileFilter
	Camera string
}

type VideoFilter struct {
	FileFilter
	MinDuration, MaxDuration *time.Duration
	Camera                   string
	MinYear, MaxYear         *int
}

type AudioFilter struct {
	FileFilter
	MinDuration, MaxDuration *time.Duration
	Author                   string
	MinYear, MaxYear         *int
}

type DocumentFilter struct {
	FileFilter
	TxtOnly bool
}

func (db DB) PrintPicturesPaths(filter PictureFilter) error {
	q := `SELECT f.path FROM file f ` +
		`INNER JOIN picture p ON p.file = f.id ` +
		`WHERE 1 = 1 ` // Ensure that "AND" can be used to add filters.
	var args []any
	q, args = addFileFilters(q, args, filter.FileFilter)
	if filter.Camera != "" {
		q += `AND p.camera = ? `
		args = append(args, filter.Camera)
	}
	q += `ORDER BY f.modified DESC `
	rows, err := db.d.Query(q, args...)
	if err != nil {
		return fmt.Errorf("could not query database: %s", err)
	}
	return printPaths(rows)
}

func (db DB) PrintVideosPaths(filter VideoFilter) error {
	q := `SELECT f.path FROM file f ` +
		`INNER JOIN video v ON v.file = f.id ` +
		`WHERE 1 = 1 ` // Ensure that "AND" can be used to add filters.
	var args []any
	q, args = addFileFilters(q, args, filter.FileFilter)
	if filter.MinDuration != nil {
		q += `AND v.seconds >= ? `
		args = append(args, int(filter.MinDuration.Seconds()))
	}
	if filter.MaxDuration != nil {
		q += `AND v.seconds <= ? `
		args = append(args, int(filter.MaxDuration.Seconds()))
	}
	if filter.Camera != "" {
		q += `AND v.camera = ? `
		args = append(args, filter.Camera)
	}
	if filter.MinYear != nil {
		q += `AND v.year >= ? `
		args = append(args, *filter.MinYear)
	}
	if filter.MaxYear != nil {
		q += `AND v.year <= ? `
		args = append(args, *filter.MaxYear)
	}
	q += `ORDER BY f.modified DESC `
	rows, err := db.d.Query(q, args...)
	if err != nil {
		return fmt.Errorf("could not query database: %s", err)
	}
	return printPaths(rows)
}

func (db DB) PrintAudiosPaths(filter AudioFilter) error {
	q := `SELECT f.path FROM file f ` +
		`INNER JOIN audio a ON a.file = f.id ` +
		`WHERE 1 = 1 ` // Ensure that "AND" can be used to add filters.
	var args []any
	q, args = addFileFilters(q, args, filter.FileFilter)
	if filter.MinDuration != nil {
		q += `AND a.seconds >= ? `
		args = append(args, int(filter.MinDuration.Seconds()))
	}
	if filter.MaxDuration != nil {
		q += `AND a.seconds <= ? `
		args = append(args, int(filter.MaxDuration.Seconds()))
	}
	if filter.Author != "" {
		q += `AND a.author = ? `
		args = append(args, filter.Author)
	}
	if filter.MinYear != nil {
		q += `AND a.year >= ? `
		args = append(args, *filter.MinYear)
	}
	if filter.MaxYear != nil {
		q += `AND a.year <= ? `
		args = append(args, *filter.MaxYear)
	}
	q += `ORDER BY f.modified DESC `
	rows, err := db.d.Query(q, args...)
	if err != nil {
		return fmt.Errorf("could not query database: %s", err)
	}
	return printPaths(rows)
}

func (db DB) PrintDocumentsPaths(filter DocumentFilter) error {
	q := `SELECT f.path FROM file f ` +
		`INNER JOIN document d ON d.file = f.id ` +
		`WHERE 1 = 1 ` // Ensure that "AND" can be used to add filters.
	var args []any
	q, args = addFileFilters(q, args, filter.FileFilter)
	if filter.TxtOnly {
		q += `AND f.mime LIKE ? `
		args = append(args, "text/%")
	}
	q += `ORDER BY f.modified DESC `
	rows, err := db.d.Query(q, args...)
	if err != nil {
		return fmt.Errorf("could not query database: %s", err)
	}
	return printPaths(rows)
}

func (db DB) PrintOthersPaths(filter FileFilter) error {
	q := `SELECT path FROM file f ` +
		`WHERE NOT EXISTS (SELECT 1 FROM picture p WHERE p.file = f.id) ` +
		`AND NOT EXISTS (SELECT 1 FROM video v WHERE v.file = f.id) ` +
		`AND NOT EXISTS (SELECT 1 FROM audio a WHERE a.file = f.id) ` +
		`AND NOT EXISTS (SELECT 1 FROM document d WHERE d.file = f.id) `
	var args []any
	q, args = addFileFilters(q, args, filter)
	q += `ORDER BY f.modified DESC `
	rows, err := db.d.Query(q, args...)
	if err != nil {
		return fmt.Errorf("could not query database: %s", err)
	}
	return printPaths(rows)
}

func (db DB) PrintAllPaths(filter FileFilter) error {
	q := `SELECT path FROM file f WHERE 1 = 1 `
	var args []any
	q, args = addFileFilters(q, args, filter)
	q += `ORDER BY f.modified DESC `
	rows, err := db.d.Query(q, args...)
	if err != nil {
		return fmt.Errorf("could not query database: %s", err)
	}
	return printPaths(rows)
}

func addFileFilters(q string, args []any, filter FileFilter) (string, []any) {
	if filter.CreatedSince != nil {
		q += `AND f.created_guess >= ? `
		args = append(args, filter.CreatedSince.Unix())
	}
	if filter.CreatedUntil != nil {
		q += `AND f.created_guess <= ? `
		args = append(args, filter.CreatedUntil.Unix())
	}
	if filter.Prefix != "" {
		q += `AND f.path LIKE ? `
		prefix := strings.ReplaceAll(filter.Prefix, `\`, `\\`)
		prefix = strings.ReplaceAll(prefix, `%`, `\%`)
		prefix = strings.ReplaceAll(prefix, `_`, `\_`)
		args = append(args, prefix+`%`)
	}
	return q, args
}

func printPaths(rows *sql.Rows) error {
	for rows.Next() {
		p := new(string)
		err := rows.Scan(&p)
		if err != nil {
			return fmt.Errorf("could not read from database: %s", err)
		}
		fmt.Println(*p)
	}
	return rows.Err()
}
