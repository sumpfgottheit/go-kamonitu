package main

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	EnvVarConfigFilePath  = "KAMONITU_CONFIG_FILE"
	DefaultConfigFilePath = "/etc/kamonitu/kamonitu.ini"
)

type AppConfig struct {
	VarDir                             string `db:"var_dir" ini:"required" validation:"writeableDirectory"`
	ConfigDir                          string `db:"config_dir" ini:"required" validation:"readableDirectory"`
	LogDir                             string `db:"log_dir" ini:"required" validation:"writeableDirectory"`
	LogLevel                           string `db:"log_level" validation:"oneOf(debug,info,warn,error)"`
	IntervalSecondsBetweenMainLoopRuns int    `db:"interval_seconds_between_main_loop_runs" validation:"within(1,60)"`
	CheckDefinitionsDir                string `db:"check_definitions_dir" validation:"readableDirectory"`
}

var appConfigDefaults = AppConfig{
	LogLevel:                           "warning",
	IntervalSecondsBetweenMainLoopRuns: 60,
	CheckDefinitionsDir:                "/etc/kamonitu/check_definitions",
}

func makeAppConfig(path string) (*AppConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		slog.Error("Config File nicht vorhanden", "path", path)
		return nil, fmt.Errorf("Config File %v nicht vorhanden", path)
	}

	appconfig, err := ParseIniFileToStruct(path, appConfigDefaults)
	if err != nil {
		return nil, err
	}
	slog.Info("Parsed ini file.", "config", appconfig)

	appconfig.CheckDefinitionsDir = appconfig.ConfigDir + "/check_definitions"
	if err = ValidateStruct(appconfig); err != nil {
		return nil, err
	}

	return appconfig, nil
}
