package database

import "time"

type File struct {
	id           int64
	Path         string
	Size         int64
	CreatedGuess time.Time
	Modified     time.Time
	MIME         string
}

type Picture struct {
	*File
	Camera string
}

type Video struct {
	*File
	Duration *time.Duration
	Camera   string
	Year     *int
}

type Audio struct {
	*File
	Duration *time.Duration
	Author   string
	Year     *int
}

type Document struct{ *File }
