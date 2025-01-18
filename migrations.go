package main

import (
	"embed"
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	"io"
	"log/slog"
	"net/url"
)

//go:embed db/migrations/*.sql
var fs embed.FS

func migrateDatabase(dbpath string) error {
	u, _ := url.Parse("sqlite:" + dbpath)
	db := dbmate.New(u)
	db.FS = fs
	db.Log = io.Discard

	migrations, err := db.FindMigrations()
	if err != nil {
		slog.Error("Error finding migrations", "err", err)
		return err
	}

	for _, migration := range migrations {
		if migration.Applied {
			slog.Info("Migration already applied", "migration", migration.FileName)
		} else {
			slog.Info("Applied migration", "migration", migration.FileName)
		}
	}

	err = db.CreateAndMigrate()
	if err != nil {
		slog.Error("Error migrating database", "err", err)
		return err
	}
	return nil
}
