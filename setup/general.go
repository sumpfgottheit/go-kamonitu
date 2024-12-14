package setup

import (
	"fmt"
	"gopkg.in/ini.v1"
	"log/slog"
	"os"
)

type IntegerValidator struct {
	required bool
	lte      int
	gte      int
}

// use a single instance of validate, it caches struct info

func configFilePath() (string, error) {
	path := os.Getenv(EnvVarIniFilePath)
	if path == "" {
		slog.Info("Kein Ini File Path aus Environment erhalten", "EnvVar", EnvVarIniFilePath, "DefaultIniFilePath", DefaultIniFilePath)
		path = DefaultIniFilePath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		slog.Error("Config File nicht vorhanden", "path", path)
		return "", fmt.Errorf("Config File %v nicht vorhanden", path)
	}

	slog.Debug("Kamonitu Ini File Path.", "path", path)
	return path, nil
}

func loadIniFile() (string, *ini.File, error) {
	path, err := configFilePath()
	if err != nil {
		return "  ", nil, err
	}
	file, err := ini.Load(path)
	if err != nil {
		return "", nil, err
	}
	slog.Debug("Kamonitu Ini File geladen.", "path", path)
	file.NameMapper = ini.TitleUnderscore
	return path, file, nil
}
