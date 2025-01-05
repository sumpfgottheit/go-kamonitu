package main

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	checkDefinitionDefaultsFileName = "check_defaults.ini"
)

var hardCodedcheckDefinitionDefaultsMap = map[string]string{
	"interval_seconds_between_checks":        "120",
	"delay_seconds_before_first_check":       "0",
	"timeout_seconds":                        "60",
	"stop_checking_after_number_of_timeouts": "3",
}
var checkDefinitionsDefaultMapFromFile map[string]string
var checkDefinitionDefaultsMap map[string]string

var checkDefinitionDefaultsSourceMap = map[string]string{
	"interval_seconds_between_checks":        "hardcoded",
	"delay_seconds_before_first_check":       "hardcoded",
	"timeout_seconds":                        "hardcoded",
	"stop_checking_after_number_of_timeouts": "hardcoded",
}

type CheckDefinition struct {
	CheckCommand                      string `db:"check_command" ini:"required"`
	IntervalSecondsBetweenChecks      int    `db:"interval_seconds_between_checks" validation:"within(5,3600)"`
	DelaySecondsBeforeFirstCheck      int    `db:"delay_seconds_before_first_check" validation:"within(0,600)"`
	TimeoutSeconds                    int    `db:"timeout_seconds" validation:"within(1,120)"`
	StopCheckingAfterNumberOfTimeouts int    `db:"stop_checking_after_number_of_timeouts" validation:"within(1,10)"`
}

type CheckDefinitionFileStore struct {
	directory              string
	CheckDefinitions       map[string]CheckDefinition
	CheckDefinitionSources map[string]map[string]string
}

// LoadCheckDefinitionDefaults loads default check definition settings from a specified INI file and updates the current configuration.
func (c *CheckDefinitionFileStore) LoadCheckDefinitionDefaults(checkDefaultsFile string) error {

	checkDefinitionDefaultsMap = make(map[string]string, len(hardCodedcheckDefinitionDefaultsMap))
	for key, value := range hardCodedcheckDefinitionDefaultsMap {
		checkDefinitionDefaultsMap[key] = value
	}

	slog.Info("Loading check definition defaults from disk.", "checkDefaultsFile", checkDefaultsFile)
	_, err := os.Stat(checkDefaultsFile)
	if err != nil {
		slog.Warn("Check definition defaults file not found.", "file", checkDefaultsFile)
		return nil
	}

	iniFileMap, err := readIniFile(checkDefaultsFile)
	if err != nil {
		slog.Error("Check definition defaults file could not be read.", "file", checkDefaultsFile)
		return err
	}
	slog.Info("Parsed ini file.", "path", checkDefaultsFile, "iniFileMap", iniFileMap)

	checkDefinitionsDefaultMapFromFile = make(map[string]string)
	for key, value := range iniFileMap {
		checkDefinitionsDefaultMapFromFile[key] = value
		checkDefinitionDefaultsSourceMap[key] = "default-ini"
	}

	for key, _ := range hardCodedcheckDefinitionDefaultsMap {
		if newKey, ok := checkDefinitionsDefaultMapFromFile[key]; ok {
			checkDefinitionDefaultsMap[key] = newKey
		}
	}
	return nil
}

// LoadCheckDefinitionsFromDisk loads check definitions from .ini files in the directory and parses their contents into structs.
// Fills the slice checkDefinitions
func (c *CheckDefinitionFileStore) LoadCheckDefinitionsFromDisk() error {

	slog.Info("Loading check files from disk.", "directory", c.directory)
	files, err := os.ReadDir(c.directory)
	if err != nil {
		return fmt.Errorf("failed to read directory %q: %v", c.directory, err)
	}
	c.CheckDefinitionSources = make(map[string]map[string]string, len(files))
	c.CheckDefinitions = make(map[string]CheckDefinition, len(files))

	var iniFileMap map[string]string
	for _, file := range files {

		if file.IsDir() || !isIniFile(file.Name()) {
			continue
		}

		path := c.directory + "/" + file.Name()
		iniFileMap, err = readIniFile(path)
		if err != nil {
			slog.Error("iniFile could not be read", "file", path)
			return err
		}

		c.CheckDefinitionSources[file.Name()] = make(map[string]string, 10)
		for key, value := range checkDefinitionDefaultsSourceMap {
			c.CheckDefinitionSources[file.Name()][key] = value
		}
		for key, _ := range iniFileMap {
			c.CheckDefinitionSources[file.Name()][key] = "check-definition-ini"
		}
		for key, value := range checkDefinitionDefaultsMap {
			if _, ok := iniFileMap[key]; !ok {
				iniFileMap[key] = value
			}
		}

		var checkDefinitionContent *CheckDefinition
		checkDefinitionContent, err = ParseStringMapToStruct(iniFileMap, CheckDefinition{})
		if err != nil {
			slog.Error("Could not parse ini file to Struct", "file", path)
			return err
		}

		err = ValidateStruct(checkDefinitionContent)
		if err != nil {
			return err
		}
		slog.Info("Parsed ini file.", "file", path, "content", checkDefinitionContent)
		c.CheckDefinitions[file.Name()] = *checkDefinitionContent
	}

	return nil
}

// makeCheckDefinitionFileStore initializes and returns a CheckDefinitionFileStore with defaults loaded from a file or hardcoded values.
func makeCheckDefinitionFileStore(config AppConfig) (*CheckDefinitionFileStore, error) {
	slog.Info("make CheckDefinitionFileStore")
	store := CheckDefinitionFileStore{
		directory: config.CheckDefinitionsDir,
	}
	slog.Info("Load check definition defaults from disk.")
	err := store.LoadCheckDefinitionDefaults(config.ConfigDir + "/" + checkDefinitionDefaultsFileName)
	if err != nil {
		return nil, err
	}
	slog.Info("Return new CheckDefinitionFileStore")
	return &store, nil
}
