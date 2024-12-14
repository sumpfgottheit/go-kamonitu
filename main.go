package main

import (
	"github.com/lmittmann/tint"
	"go-kamonitu/setup"
	"log/slog"
	"os"
	"time"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {

	loglevel := slog.LevelInfo
	if os.Getenv("KAMONITU_DEBUG") == "1" {
		loglevel = slog.LevelDebug
	}
	setupLogging(loglevel)

	err := setup.InitAppConfig()
	if err != nil {
		slog.Error("Konfiguration konnte nicht geladen werden", "Error", err)
		return
	}

	err = setup.InitCheckDefinitionDefaults()
	if err != nil {
		slog.Error("Check-Defaults konnten nicht geladen werden", "Error", err)
		return
	}
}

func setupLogging(loglevel slog.Level) {
	w := os.Stderr
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			AddSource:  true,
			Level:      loglevel,
			TimeFormat: time.RFC3339,
		}),
	))
	slog.Info("Initialized Logging.", "Level", loglevel.String())
}
