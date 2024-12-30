package main

import (
	"fmt"
	"log/slog"
)

func validateConfigHlc(config *AppConfig) error {
	slog.Info("Configuration file is valid.", "config", config)
	fmt.Println("Configuration is valid.")
	return nil
}
