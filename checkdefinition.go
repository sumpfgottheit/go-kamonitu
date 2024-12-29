package main

import (
	"fmt"
	"log/slog"
	"os"
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

func (c *CheckDefinitionFileStore) LoadCheckDefinitionsFromDisk() error {
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
		checkDefinitionContent, err := ParseIniFileToStruct(filePath, ck)
		if err != nil {
			return err
		}
		slog.Info("Parsed ini file.", "file", filePath, "content", checkDefinitionContent)
		c.CheckDefinitions[fileName] = *checkDefinitionContent
	}
	return nil
}
