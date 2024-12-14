package setup

import (
	"fmt"
	"gopkg.in/ini.v1"
	"log/slog"
)

const (
	IniKeyIntervalSecondsBetweenChecks      = "interval_seconds_between_checks"
	IniKeyDelaySecondsBeforeFirstCheck      = "delay_seconds_before_first_check"
	IniKeyTimeoutSeconds                    = "timeout_seconds"
	IniKeyStopCheckingAfterNumberOfTimeouts = "stop_checking_after_number_of_timeouts"
)

type CheckDefinitionDefaults struct {
	IntervalSecondsBetweenChecks      int `db:"interval_seconds_between_checks" ini:"required"`
	DelaySecondsBeforeFirstCheck      int `db:"delay_seconds_before_first_check" ini:"required"`
	TimeoutSeconds                    int `db:"timeout_seconds"`
	StopCheckingAfterNumberOfTimeouts int `db:"stop_checking_after_number_of_timeouts"`
}

var checkDefinitionHardcodedDefaults = CheckDefinitionDefaults{
	IntervalSecondsBetweenChecks:      60,
	DelaySecondsBeforeFirstCheck:      0,
	TimeoutSeconds:                    60,
	StopCheckingAfterNumberOfTimeouts: 3,
}

var checkDefinitionDefaultsFromIniFile = checkDefinitionHardcodedDefaults

var checkDefinitionIntegerValidators = map[string]IntegerValidator{
	IniKeyIntervalSecondsBetweenChecks: {
		required: false,
		lte:      3600,
		gte:      1,
	},
	IniKeyDelaySecondsBeforeFirstCheck: {
		required: false,
		lte:      3600,
		gte:      0,
	},
	IniKeyTimeoutSeconds: {
		required: false,
		lte:      300,
		gte:      1,
	},
	IniKeyStopCheckingAfterNumberOfTimeouts: {
		required: false,
		lte:      5,
		gte:      1,
	},
}

func validateIniFileSectionCheckDefaults(path string, file *ini.File) error {
	if file.Section("check_defaults") == nil {
		return fmt.Errorf("Section check_defaults muss vorhanden sein")
	}

	// Ensure all required Keys exist
	section := file.Section("check_defaults")

	for key, validator := range checkDefinitionIntegerValidators {
		if validator.required && !section.HasKey(key) {
			return fmt.Errorf("%v, Section [check_defaults] %v muss vorhanden sein", path, key)
		}
		if section.HasKey(key) {
			value, err := section.Key(key).Int()
			if err != nil {
				return fmt.Errorf("%v, Section [check_defaults]: Key %v ist kein Integer: %v", path, key, err)
			}
			if value < validator.gte || value > validator.lte {
				return fmt.Errorf("%v, Section [check_defaults]: Key  %v: muss zwischen %v und %v liegen", path, key, validator.gte, validator.lte)
			}
		}
	}

	return nil
}

func InitCheckDefinitionDefaults() error {
	path, file, err := loadIniFile()
	if err != nil {
		return err
	}

	err = validateIniFileSectionCheckDefaults(path, file)
	if err != nil {
		return err
	}
	err = file.Section("check_defaults").MapTo(&checkDefinitionDefaultsFromIniFile)
	if err != nil {
		return err
	}
	slog.Debug("CheckDefaults", "checkDefaults", checkDefinitionDefaultsFromIniFile)

	return nil

}
