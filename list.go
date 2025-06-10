package den

import "github.com/codesoap/den/database"

func List(db database.DB) ([]string, error) {
	return db.TrackedPaths()
}
