package database

import (
	"context"
	"database/sql"
	"fmt"
)

var errAlreadyExists = fmt.Errorf("an entry for this path already exists")

// BeginTx starts a transaction. It must be called before writing to the
// database. If a transaction is already in progress, BeginTx will block
// until a transaction can be started.
//
// Do not use when just reading from the database.
//
// It is the callers responsibility to call db.Commit or db.Rollback the
// end the transaction.
func (db *DB) BeginTx() error {
	db.txLock.Lock()
	tx, err := db.d.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		db.txLock.Unlock()
		return fmt.Errorf("could not start transaction: %s", err)
	}
	db.tx = tx
	return nil
}

func (db *DB) Commit() error {
	defer func() {
		db.txLock.Unlock()
		db.tx = nil
	}()
	if err := db.tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %s", err)
	}
	return nil
}

func (db *DB) Rollback() error {
	defer func() {
		db.txLock.Unlock()
		db.tx = nil
	}()
	if err := db.tx.Rollback(); err != nil {
		return fmt.Errorf("could not roll back transaction: %s", err)
	}
	return nil
}

// AddPicture adds a picture to the database. If the picture already
// exists, nothing will be done. db.BeginTx must have been called
// before.
func (db DB) AddPicture(pic *Picture) error {
	id, err := db.addFile(pic.File)
	if err == errAlreadyExists {
		return nil
	} else if err != nil {
		return err
	}
	q := `INSERT INTO picture (file, camera) VALUES (?, ?)`
	var camera *string
	if pic.Camera != "" {
		camera = &pic.Camera
	}
	if _, err = db.tx.Exec(q, id, camera); err != nil {
		f := "could not add picture for file with ID %d: %s"
		return fmt.Errorf(f, id, err)
	}
	return nil
}

// AddVideo adds a picture to the database. If the video already exists,
// nothing will be done. db.BeginTx must have been called before.
func (db DB) AddVideo(vid *Video) error {
	id, err := db.addFile(vid.File)
	if err == errAlreadyExists {
		return nil
	} else if err != nil {
		return err
	}
	q := `INSERT INTO video (file, seconds, camera, year) ` +
		`VALUES (?, ?, ?, ?)`
	var seconds *int
	var camera *string
	if vid.Duration != nil {
		s := int(vid.Duration.Seconds())
		seconds = &s
	}
	if vid.Camera != "" {
		camera = &vid.Camera
	}
	_, err = db.tx.Exec(q, id, seconds, camera, vid.Year)
	if err != nil {
		f := "could not add video for file with ID %d: %s"
		return fmt.Errorf(f, id, err)
	}
	return nil
}

// AddAudio adds an audio file to the database. If the audio already
// exists, nothing will be done. db.BeginTx must have been called
// before.
func (db DB) AddAudio(audio *Audio) error {
	id, err := db.addFile(audio.File)
	if err == errAlreadyExists {
		return nil
	} else if err != nil {
		return err
	}
	q := `INSERT INTO audio (file, seconds, author, year) ` +
		`VALUES (?, ?, ?, ?)`
	var seconds *int
	var author *string
	if audio.Duration != nil {
		s := int(audio.Duration.Seconds())
		seconds = &s
	}
	if audio.Author != "" {
		author = &audio.Author
	}
	_, err = db.tx.Exec(q, id, seconds, author, audio.Year)
	if err != nil {
		f := "could not add audio for file with ID %d: %s"
		return fmt.Errorf(f, id, err)
	}
	return nil
}

// AddDocument adds a document to the database. If the document already
// exists, nothing will be done. db.BeginTx must have been called
// before.
func (db DB) AddDocument(doc *Document) error {
	id, err := db.addFile(doc.File)
	if err == errAlreadyExists {
		return nil
	} else if err != nil {
		return err
	}
	q := `INSERT INTO document (file) VALUES (?)`
	if _, err = db.tx.Exec(q, id); err != nil {
		f := "could not add document for file with ID %d: %s"
		return fmt.Errorf(f, id, err)
	}
	return nil
}

// AddFile adds an unspecific file to the database. If the file already
// exists, nothing will be done. db.BeginTx must have been called
// before.
func (db DB) AddFile(file *File) error {
	_, err := db.addFile(file)
	if err == errAlreadyExists {
		return nil
	}
	return err
}

func (db DB) addFile(file *File) (int64, error) {
	q := `INSERT INTO file (path, size, created_guess, modified, mime) ` +
		`VALUES (?, ?, ?, ?, ?) ` +
		`ON CONFLICT (path) DO NOTHING`
	res, err := db.tx.Exec(q,
		file.Path,
		file.Size,
		file.CreatedGuess.Unix(),
		file.Modified.Unix(),
		file.MIME,
	)
	if err != nil {
		return 0, fmt.Errorf("could not create file entry: %s", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		f := "could not find ID of created file entry: %s"
		return id, fmt.Errorf(f, err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return id, fmt.Errorf("could not determine success: %s", err)
	} else if n == 0 {
		return id, errAlreadyExists
	}
	return id, nil
}
