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
	configFilePath                     string
}

var appConfigDefaultMap = map[string]string{
	"var_dir":    "/var/lib/kamonitu",
	"config_dir": "/etc/kamonitu",
	"log_dir":    "/var/log/kamonitu",
	"log_level":  "warn",
	"interval_seconds_between_main_loop_runs": "60",
	"check_definitions_dir":                   "/etc/kamonitu/check_definitions",
}
var appConfigMap = make(map[string]string, len(appConfigDefaultMap))

var appConfigSourceMap = map[string]string{
	"var_dir":    "hardcoded",
	"config_dir": "hardcoded",
	"log_dir":    "hardcoded",
	"log_level":  "hardcoded",
	"interval_seconds_between_main_loop_runs": "hardcoded",
	"check_definitions_dir":                   "hardcoded",
}

func makeAppConfig(path string) (*AppConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		slog.Error("Config File nicht vorhanden", "path", path)
		return nil, fmt.Errorf("Config File %v nicht vorhanden", path)
	}

	iniFileMap, err := readIniFile(path)
	if err != nil {
		return nil, err
	}
	slog.Info("Parsed ini file.", "path", path)

	for key, value := range appConfigDefaultMap {
		appConfigMap[key] = value
	}
	for key, value := range iniFileMap {
		appConfigMap[key] = value
		appConfigSourceMap[key] = "ini"
	}

	appconfig, err := ParseStringMapToStruct(appConfigMap, AppConfig{})
	if err != nil {
		return nil, err
	}
	slog.Info("Parsed ini file to struct.", "path", path, "appconfig", appconfig)
	appconfig.configFilePath = path

	appconfig.CheckDefinitionsDir = appconfig.ConfigDir + "/check_definitions"
	if err = ValidateStruct(appconfig); err != nil {
		return nil, err
	}

	return appconfig, nil
}
