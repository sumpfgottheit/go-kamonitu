package main

// ReplaceKamonituResults deletes all existing kamonitu results with the given tag and inserts new results in the database.
// The results are marked as warnings.
func ReplaceKamonituResults(errors []string, tag string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing kamonitu results
	_, err = tx.Exec("DELETE FROM results WHERE filename = ? and tags = ?", "kamonitu", tag)
	if err != nil {
		return err
	}

	for _, myerror := range errors {
		_, err = tx.Exec("INSERT INTO results (filename, rc, name, text, tags) VALUES ('kamonitu', 1, 'Kamonitu Internal', ?, ?)", myerror, tag)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
