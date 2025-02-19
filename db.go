// db.go
package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
)

var db *sqlx.DB

const (
	sqlite_connect_options = "?_foreign_keys=on&_busy_timeout=10000&_journal_mode=wal"
)

// Initialize the database connection
func initDB(path string) (*sqlx.DB, error) {
	var err error
	dbpath := path + sqlite_connect_options
	db, err = sqlx.Connect("sqlite3", dbpath)
	if err != nil {
		slog.Error("Failed to connect to database %s: %v", dbpath, err)
		return nil, err

	}
	slog.Info("Connected to database", "path", dbpath)
	return db, nil
}

// Close the database connection
func closeDB() {
	if db != nil {
		db.Close()
	}
}
