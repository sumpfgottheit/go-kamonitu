package main

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	checkDefinitionDefaultsFileName = "check_defaults.ini"
)

type CheckDefinitionDefaults struct {
	IntervalSecondsBetweenChecks      int `db:"interval_seconds_between_checks"`
	DelaySecondsBeforeFirstCheck      int `db:"delay_seconds_before_first_check"`
	TimeoutSeconds                    int `db:"timeout_seconds"`
	StopCheckingAfterNumberOfTimeouts int `db:"stop_checking_after_number_of_timeouts"`
}

type CheckDefinition struct {
	CheckDefinitionDefaults
	CheckCommand        string `db:"check_command" ini:"required"`
	DefinitionFilePath  string `db:"definition_file_path"`
	DefinitionFileMtime int64  `db:"definition_file_mtime"`
}

var checkDefinitionHardcodedDefaults = CheckDefinitionDefaults{
	IntervalSecondsBetweenChecks:      60,
	DelaySecondsBeforeFirstCheck:      0,
	TimeoutSeconds:                    60,
	StopCheckingAfterNumberOfTimeouts: 3,
}

type CheckDefinitionFileStore struct {
	directory               string
	definitionsFilesMtimes  map[string]int64
	checkDefinitionDefaults CheckDefinitionDefaults
	CheckDefinitions        map[string]CheckDefinition
}

// LoadCheckDefinitionDefaults loads default check definitions from a specified file.
// Falls back to hardcoded defaults if the file is missing or invalid.
// Validates and parses the file content into the appropriate struct.
// Returns an error if reading, parsing, or validation fails.
func (c *CheckDefinitionFileStore) LoadCheckDefinitionDefaults(checkDefaultsFile string) error {
	slog.Info("Loading check definition defaults from disk.", "checkDefaultsFile", checkDefaultsFile)
	_, err := os.Stat(checkDefaultsFile)
	if err != nil {
		slog.Warn("Check definition defaults file not found.", "file", checkDefaultsFile)
		c.checkDefinitionDefaults = checkDefinitionHardcodedDefaults
		return nil
	}

	iniFileMap, err := readIniFile(checkDefaultsFile)
	if err != nil {
		slog.Error("Check definition defaults file could not be read.", "file", checkDefaultsFile)
		return err
	}

	checkDefinitionDefaultsContent, err := ParseStringMapToStruct(iniFileMap, checkDefinitionHardcodedDefaults)
	if err != nil {
		slog.Error("Check definition defaults file could not be parsed.", "file", checkDefaultsFile)
		return err
	}

	err = ValidateStruct(checkDefinitionDefaultsContent)
	if err != nil {
		slog.Error("Check definition defaults file could not be validated.", "file", checkDefaultsFile)
		return err
	}
	slog.Info("Loaded check definition defaults from disk.", "file", checkDefaultsFile, "content", checkDefinitionDefaultsContent)
	c.checkDefinitionDefaults = *checkDefinitionDefaultsContent
	return nil
}

// LoadDefinitionsFilesMtimeMap scans the directory for .ini files, retrieves their modification times, and updates the mtimes map.
// Returns an error if directory reading or file stat operation fails.
func (c *CheckDefinitionFileStore) LoadDefinitionsFilesMtimeMap() error {
	slog.Info("Loading check files with mtimes from disk.", "directory", c.directory)
	files, err := os.ReadDir(c.directory)
	if err != nil {
		return fmt.Errorf("failed to read directory %q: %v", c.directory, err)
	}
	c.definitionsFilesMtimes = make(map[string]int64, len(files))
	for _, file := range files {
		if file.IsDir() || !isIniFile(file.Name()) {
			continue
		}

		filePath := c.directory + "/" + file.Name()
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("failed to stat file %q: %v", filePath, err)
		}
		slog.Info("Found check definition file. Fetch Mtime", "file", filePath)

		c.definitionsFilesMtimes[file.Name()] = fileInfo.ModTime().Unix()
	}
	slog.Info("Loaded check files with mtimes from disk.", "length", len(c.definitionsFilesMtimes))
	return nil
}

// LoadCheckDefinitionsFromDisk loads check definitions from .ini files in the directory and parses their contents into structs.
// Fills the slice checkDefinitions
func (c *CheckDefinitionFileStore) LoadCheckDefinitionsFromDisk() error {
	slog.Info("Loading check definitions from disk.", "directory", c.directory)
	err := c.LoadDefinitionsFilesMtimeMap()
	if err != nil {
		return err
	}
	c.CheckDefinitions = make(map[string]CheckDefinition, len(c.definitionsFilesMtimes))
	for fileName, mtime := range c.definitionsFilesMtimes {
		filePath := c.directory + "/" + fileName
		slog.Info("Found check definition file.", "file", filePath)

		ck := CheckDefinition{
			DefinitionFilePath:      filePath,
			DefinitionFileMtime:     mtime,
			CheckDefinitionDefaults: c.checkDefinitionDefaults,
		}
		iniFileMap, err := readIniFile(filePath)
		if err != nil {
			return err
		}

		checkDefinitionContent, err := ParseStringMapToStruct(iniFileMap, ck)
		if err != nil {
			return err
		}

		err = ValidateStruct(checkDefinitionContent)
		if err != nil {
			return err
		}
		slog.Info("Parsed ini file.", "file", filePath, "content", checkDefinitionContent)
		c.CheckDefinitions[fileName] = *checkDefinitionContent
	}
	return nil
}

// makeCheckDefinitionFileStore initializes and returns a CheckDefinitionFileStore with defaults loaded from a file or hardcoded values.
func makeCheckDefinitionFileStore(checkDefinitionsDir string, checkDefinitionsDefaultFile string) (*CheckDefinitionFileStore, error) {
	slog.Info("make CheckDefinitionFileStore")
	store := CheckDefinitionFileStore{
		directory:               checkDefinitionsDir,
		definitionsFilesMtimes:  make(map[string]int64),
		checkDefinitionDefaults: checkDefinitionHardcodedDefaults,
	}
	slog.Info("Load check definition defaults from disk.")
	err := store.LoadCheckDefinitionDefaults(checkDefinitionsDefaultFile)
	if err != nil {
		return nil, err
	}
	slog.Info("Return new CheckDefinitionFileStore")
	return &store, nil
}
