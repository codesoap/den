package den

import "github.com/codesoap/den/database"

func Delete(path string, db database.DB) error {
	return db.UntrackPath(path)
}
