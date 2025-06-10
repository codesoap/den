package database

const schemaV1 = `
CREATE TABLE tracked_path(
	path TEXT NOT NULL,
	UNIQUE(path)
);

CREATE TABLE file(
	id            INTEGER PRIMARY KEY,
	path          TEXT NOT NULL,
	size          INTEGER NOT NULL,
	created_guess INTEGER NOT NULL,
	modified      INTEGER NOT NULL,
	mime          TEXT NOT NULL,
	UNIQUE(path)
);
CREATE INDEX file_path ON file(path);
CREATE INDEX file_created_guess ON file(created_guess);
CREATE INDEX file_modified ON file(modified);
CREATE INDEX file_mime ON file(mime);

CREATE TABLE picture(
	file   INTEGER PRIMARY KEY,
	camera TEXT,
	FOREIGN KEY(file) REFERENCES file(id) ON DELETE CASCADE
);
CREATE INDEX picture_camera ON picture(camera);

CREATE TABLE video(
	file    INTEGER PRIMARY KEY,
	seconds INTEGER,
	camera  TEXT,
	year    INTEGER,
	FOREIGN KEY(file) REFERENCES file(id) ON DELETE CASCADE
);
CREATE INDEX video_seconds ON video(seconds);
CREATE INDEX video_camera ON video(camera);
CREATE INDEX video_year ON video(year);

CREATE TABLE audio(
	file    INTEGER PRIMARY KEY,
	seconds INTEGER,
	author  TEXT,
	year    INTEGER,
	FOREIGN KEY(file) REFERENCES file(id) ON DELETE CASCADE
);
CREATE INDEX audio_seconds ON audio(seconds);
CREATE INDEX audio_author ON audio(author);
CREATE INDEX audio_year ON audio(year);

CREATE TABLE document(
	file INTEGER PRIMARY KEY,
	FOREIGN KEY(file) REFERENCES file(id) ON DELETE CASCADE
);

PRAGMA user_version = 1;
`
