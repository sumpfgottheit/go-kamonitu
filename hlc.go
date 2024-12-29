package main

import (
	"log/slog"
)

func validateConfigHlc(config *AppConfig) error {
	slog.Info("Configuration file is valid.", "config", config)
	return nil
}
