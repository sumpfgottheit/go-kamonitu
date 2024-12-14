package setup

import (
	"errors"
	"fmt"
	"gopkg.in/ini.v1"
	"log/slog"
	"os"
)

const (
	EnvVarIniFilePath                        = "KAMONITU_INI_FILE"
	DefaultIniFilePath                       = "/etc/kamonitu.ini"
	IniKeyConfigDir                          = "config_dir"
	IniKeyVarDir                             = "var_dir"
	IniKeyIntervalSecondsBetweenMainLoopRuns = "interval_seconds_between_main_loop_runs"
)

type AppConfig struct {
	ConfigDir                          string `ini:"required"`
	VarDir                             string `ini:"required"`
	IntervalSecondsBetweenMainLoopRuns int
}

func (a *AppConfig) Validate() error {
	if a.ConfigDir == "" {
		slog.Error("ConfigDir muss gesetzt sein")
		return errors.New("ConfigDir muss gesetzt sein")
	}
	d, err := os.Stat(a.ConfigDir)
	if err != nil {
		slog.Error("ConfigDir kann nicht gelesen werden: ", "Error", err)
		return fmt.Errorf("ConfigDir kann nicht gelesen werden: %v", err)
	}
	if !d.IsDir() {
		slog.Error("ConfigDir ist kein Verzeichnis", "ConfigDir", a.ConfigDir)
		return fmt.Errorf("ConfigDir %v ist kein Verzeichnis", a.ConfigDir)
	}
	if a.VarDir == "" {
		slog.Error("VarDir muss gesetzt sein")
		return errors.New("VarDir muss gesetzt sein")
	}
	d, err = os.Stat(a.VarDir)
	if err != nil {
		slog.Error("ConfigDir kann nicht gelesen werden", "ConfigDir", err)
		return fmt.Errorf("ConfigDir kann nicht gelesen werden: %v", err)
	}
	if !d.IsDir() {
		slog.Error("ConfigDir ist kein Verzeichnis", "ConfigDir", a.ConfigDir)
		return fmt.Errorf("ConfigDir %v ist kein Verzeichnis", a.ConfigDir)
	}
	return nil
}

var appConfigHardCodedDefaults = AppConfig{
	ConfigDir:                          "/etc/kamonitu.d",
	VarDir:                             "/var/lib/kamonitu",
	IntervalSecondsBetweenMainLoopRuns: 60,
}

// Neues Struct AppCofig mit den Werten aus appConfigHardCodedDefaults
var Config = appConfigHardCodedDefaults

func validateIniFileSectionKamonitu(path string, file *ini.File) error {
	if file.Section("kamonitu") == nil {
		return fmt.Errorf("Section kamonitu muss vorhanden sein")
	}

	// Ensure all required Keys exist
	section := file.Section("kamonitu")
	required_keys := []string{IniKeyVarDir, IniKeyConfigDir}
	for _, keyName := range required_keys {
		if !section.HasKey(keyName) {
			return fmt.Errorf("%v: Section kamonitu muss Key %v enthalten", path, keyName)
		}
	}

	integerValidators := map[string]IntegerValidator{
		IniKeyIntervalSecondsBetweenMainLoopRuns: {
			required: false,
			lte:      3600,
			gte:      10,
		},
	}

	for key, validator := range integerValidators {
		if validator.required && !section.HasKey(key) {
			return fmt.Errorf("%v, Section [kamonitu] %v muss vorhanden sein", path, key)
		}
		if section.HasKey(key) {
			value, err := section.Key(key).Int()
			if err != nil {
				return fmt.Errorf("%v, Section [kamonitu]: Key %v ist kein Integer: %v", path, key, err)
			}
			if value < validator.gte || value > validator.lte {
				return fmt.Errorf("%v, Section [kamonitu]: Key  %v: muss zwischen %v und %v liegen", path, key, validator.gte, validator.lte)
			}
		}
	}

	return nil
}

func InitAppConfig() error {
	path, file, err := loadIniFile()
	if err != nil {
		return err
	}
	err = validateIniFileSectionKamonitu(path, file)
	if err != nil {
		return err
	}
	err = file.Section("kamonitu").MapTo(&Config)
	if err != nil {
		return err
	}
	err = Config.Validate()
	if err != nil {
		return err
	}
	slog.Debug("Config", "Config", Config)

	return nil
}
