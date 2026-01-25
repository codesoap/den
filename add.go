package den

import (
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codesoap/den/database"
	"github.com/codesoap/den/internal/mediainfo"
	"github.com/codesoap/den/internal/mimecat"

	"github.com/djherbis/times"
)

type addition struct {
	path      string
	d         fs.DirEntry
	mime      string
	mediainfo mediainfo.Info
}

type Progress struct{ Done, Total int }

// Add adds the given path to the tracked paths and indexes all
// non-hidden files in path. If a path is already tracked, an error will
// be returned.
//
// Progress updates will be written roughly once per second to the
// progress channel. The progress channel will be closed before the
// function returns.
func Add(path string, db database.DB, progress chan Progress) error {
	defer close(progress)
	path, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("could not normalize path: %s", err)
	}
	if err := db.TrackPath(path); err != nil {
		return fmt.Errorf("could not track path '%s': %s", path, err)
	}
	prog := Progress{}
	var lastProgressUpdate time.Time
	paths := make(map[string]fs.DirEntry)
	err = filepath.WalkDir(path,
		func(path string, d fs.DirEntry, err error) error {
			if time.Since(lastProgressUpdate) >= time.Second {
				progress <- prog
				lastProgressUpdate = time.Now()
			}
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
			paths[path] = d
			prog.Total++
			return nil
		})
	if err != nil {
		return fmt.Errorf("could not index files: %s", err)
	}

	entriesInTx := 0
	for path, d := range paths {
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
		if err = indexFile(db, path, d); err != nil {
			_ = db.Rollback()
			return err
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

func indexFile(db database.DB, path string, d fs.DirEntry) error {
	_, err := d.Info()
	if err != nil {
		return fmt.Errorf("could not get file info on '%s': %s", path, err)
	}
	m, err := determineMIME(path)
	if err != nil {
		return fmt.Errorf("could not determine mime type of '%s': %s", path, err)
	}
	a := addition{
		path: path,
		d:    d,
		mime: m,
	}
	cat := mimecat.MIMEToCategory(m)
	if cat == mimecat.Video || cat == mimecat.Audio || cat == mimecat.Picture {
		a.mediainfo, _ = mediainfo.MediaInfo(path)
	}
	switch {
	case cat == mimecat.Other:
		if err = addFile(a, db); err != nil {
			return fmt.Errorf("could not add other file '%s': %s", path, err)
		}
	case cat == mimecat.Picture:
		if err = addPicture(a, db); err != nil {
			return fmt.Errorf("could not add picture '%s': %s", path, err)
		}
	case cat == mimecat.Video && a.mediainfo.Type != mediainfo.TypeAudio:
		if err = addVideo(a, db); err != nil {
			return fmt.Errorf("could not add video '%s': %s", path, err)
		}
	case cat == mimecat.Video && a.mediainfo.Type == mediainfo.TypeAudio:
		fallthrough
	case cat == mimecat.Audio:
		if err = addAudio(a, db); err != nil {
			return fmt.Errorf("could not add audio '%s': %s", path, err)
		}
	case cat == mimecat.Document:
		if err = addDocument(a, db); err != nil {
			return fmt.Errorf("could not add document '%s': %s", path, err)
		}
	}
	return nil
}

func determineMIME(path string) (string, error) {
	var mimeType string
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("could not open file '%s': %s", path, err)
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	mimeType = http.DetectContentType(buf[:n])
	if idx := strings.IndexByte(mimeType, ';'); idx != -1 {
		mimeType = mimeType[:idx]
	}
	ext := filepath.Ext(path)
	if n == 0 || mimeType == "application/octet-stream" {
		secondOpinion := mime.TypeByExtension(ext)
		if secondOpinion != "" {
			mimeType = secondOpinion
			if idx := strings.IndexByte(mimeType, ';'); idx != -1 {
				mimeType = mimeType[:idx]
			}
		}
	} else if (mimeType == "text/plain" || mimeType == "text/xml") &&
		mime.TypeByExtension(ext) == "image/svg+xml" {
		mimeType = "image/svg+xml"
	} else if mimeType == "application/zip" {
		switch strings.ToLower(ext) {
		case ".odt":
			mimeType = "application/vnd.oasis.opendocument.text"
		case ".ods":
			mimeType = "application/vnd.oasis.opendocument.spreadsheet"
		case ".odp":
			mimeType = "application/vnd.oasis.opendocument.presentation "
		case ".docx":
			mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		case ".xlsx":
			mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		case ".pptx":
			mimeType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
		}
	}
	return mimeType, nil
}

func addFile(a addition, db database.DB) error {
	f, err := toFile(a)
	if err != nil {
		return err
	}
	return db.AddFile(f)
}

func addPicture(a addition, db database.DB) error {
	f, err := toFile(a)
	if err != nil {
		return err
	}
	pic := &database.Picture{
		File:   f,
		Camera: a.mediainfo.Camera,
	}
	return db.AddPicture(pic)
}

func addVideo(a addition, db database.DB) error {
	f, err := toFile(a)
	if err != nil {
		return err
	}
	vid := &database.Video{
		File:     f,
		Duration: a.mediainfo.Duration,
		Camera:   a.mediainfo.Camera,
		Year:     a.mediainfo.Year,
	}
	return db.AddVideo(vid)
}

func addAudio(a addition, db database.DB) error {
	f, err := toFile(a)
	if err != nil {
		return err
	}
	pic := &database.Audio{
		File:     f,
		Duration: a.mediainfo.Duration,
		Author:   a.mediainfo.Author,
		Year:     a.mediainfo.Year,
	}
	return db.AddAudio(pic)
}

func addDocument(a addition, db database.DB) error {
	f, err := toFile(a)
	if err != nil {
		return err
	}
	pic := &database.Document{
		File: f,
	}
	return db.AddDocument(pic)
}

func toFile(a addition) (*database.File, error) {
	info, err := a.d.Info()
	if err != nil {
		f := "could not get file info for '%s': %s"
		return &database.File{}, fmt.Errorf(f, a.path, err)
	}

	// Determining the time when a file was created seems to be
	// difficult. Just take the oldest one we can find:
	var created time.Time
	ts := times.Get(info)
	timeFound := false
	if ts.HasBirthTime() {
		t := ts.BirthTime()
		created = t
		timeFound = true
	}
	if ts.HasChangeTime() {
		t := ts.ChangeTime()
		if !timeFound || t.Before(created) {
			created = t
		}
		timeFound = true
	}
	if !timeFound || ts.ModTime().Before(created) {
		t := ts.ModTime()
		created = t
		timeFound = true
	}

	return &database.File{
		Path:         a.path,
		Size:         info.Size(),
		CreatedGuess: created,
		Modified:     info.ModTime(),
		MIME:         a.mime,
	}, nil
}
