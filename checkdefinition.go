package main

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/jmoiron/sqlx"
	"log/slog"
	"os"
	"path/filepath"
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
	db                     *sqlx.DB
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

func loadSingleCheckDefinitionFromFile(path string) (checkDefinitionContent *CheckDefinition, sources map[string]string, e error) {
	iniFileMap, err := readIniFile(path)
	if err != nil {
		return nil, nil, err
	}

	/**** Merken der Sourcen für Ausgabe von show-config ****/
	// checkDefinitionDefaultsSourceMap hat bereits Hardcoded und default.ini file korrekt gesetzt
	sources = make(map[string]string, 10)
	for key, value := range checkDefinitionDefaultsSourceMap {
		sources[key] = value
	}
	// alles Keys die im checkdefinitionfile sind, überschreiben diese in der source map
	for key, _ := range iniFileMap {
		sources[key] = filepath.Base(path)
	}
	/**** Sourcen für ausgabe von show-config gemerkt ****/

	// Kopiere die Default Key/Values, falls diese nicht im checkDefinitionFile gesetzt sind
	for key, value := range checkDefinitionDefaultsMap {
		if _, ok := iniFileMap[key]; !ok {
			iniFileMap[key] = value
		}
	}
	slog.Info("Parsed ini file.", "path", path, "iniFileMap", iniFileMap)

	checkDefinitionContent, err = ParseStringMapToStruct(iniFileMap, CheckDefinition{})
	if err != nil {
		slog.Error("Could not parse ini file to Struct", "file", path, "err", err)
		return nil, nil, err
	}

	err = ValidateStruct(checkDefinitionContent)
	if err != nil {
		slog.Error("error validating struct", "file", path, "err", err)
		return nil, nil, err

	}
	slog.Info("Parsed ini file.", "file", path, "content", checkDefinitionContent)

	return checkDefinitionContent, sources, nil
}

// LoadCheckDefinitionsFromDisk loads check definitions from .ini files in the directory and parses their contents into structs.
// Fills the slice checkDefinitions
func (c *CheckDefinitionFileStore) LoadCheckDefinitionsFromDisk() error {
	var errors *multierror.Error
	slog.Info("Loading check files from disk.", "directory", c.directory)
	files, err := os.ReadDir(c.directory)
	if err != nil {
		return fmt.Errorf("failed to read directory %q: %v", c.directory, err)
	}
	c.CheckDefinitionSources = make(map[string]map[string]string, len(files))
	c.CheckDefinitions = make(map[string]CheckDefinition, len(files))

	for _, file := range files {

		if file.IsDir() || !isIniFile(file.Name()) {
			continue
		}

		path := c.directory + "/" + file.Name()

		checkDefinition, sources, err := loadSingleCheckDefinitionFromFile(path)
		if err != nil {
			myerr := fmt.Errorf("failed to load check definition file %q: %v", path, err)
			slog.Error(myerr.Error())
			errors = multierror.Append(errors, myerr)
			continue
		}

		c.CheckDefinitionSources[file.Name()] = sources
		c.CheckDefinitions[file.Name()] = *checkDefinition
	}

	return errors.ErrorOrNil()
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

func (c *CheckDefinitionFileStore) ensureCheckDefinitionsInDatabase() error {
	/*
	 * Remove CheckDefinitions that no longer exist
	 */
	filenames := getKeys(c.CheckDefinitions)
	slog.Info("Remove CheckDefinitions that no longer exist", "filenames", filenames)
	query, args, err := sqlx.In("delete from check_definitions where filename not in (?);", filenames)
	if err != nil {
		slog.Error("Error building query 'delete from chech_definitions where filename not in (?)'", "query", query, "args", args, "err", err)
		return err
	}
	// sqlx.In returns queries with the `?` bindvar, we can rebind it for our backend
	query = c.db.Rebind(query)
	rows, err := c.db.Query(query, args...)
	if err != nil {
		slog.Error("Error executing query 'delete from chech_definitions where filename not in (?)'", "query", query, "args", args, "err", err)
		return err
	}
	rows.Close()

	/*
	 * Upsert existing CheckDefinitions
	 */
	for filename, cd := range c.CheckDefinitions {
		sql := `insert into 
    				check_definitions(filename, check_command, interval_seconds_between_checks, delay_seconds_before_first_check, timeout_seconds, stop_checking_after_number_of_timeouts) 
					values(?,?,?,?,?,?)
				on conflict(filename) do 
					update 
					    set check_command=?, 
					    interval_seconds_between_checks=?, 
					    delay_seconds_before_first_check=?, 
					    timeout_seconds=?, 
					    stop_checking_after_number_of_timeouts=?`
		_, err := c.db.Exec(sql, filename, cd.CheckCommand, cd.IntervalSecondsBetweenChecks, cd.DelaySecondsBeforeFirstCheck, cd.TimeoutSeconds, cd.StopCheckingAfterNumberOfTimeouts, cd.CheckCommand, cd.IntervalSecondsBetweenChecks, cd.DelaySecondsBeforeFirstCheck, cd.TimeoutSeconds, cd.StopCheckingAfterNumberOfTimeouts)
		if err != nil {
			slog.Error("Error executing query 'insert into check_definitions'", "sql", sql, "err", err)
			return err
		}
	}
	slog.Info("CheckDefinitions in database updated.")

	return nil
}

//func (c *CheckDefinitionFileStore) checkDefinitionsToRun() []CheckDefinition {
//	checkDefinitionsToRun = []CheckDefinition{}
//	filenames := []string{}
//	err := c.db.Select(&filenames, "select filename from check_definitions where unixepoch() - interval_seconds_between_checks >= last_run_timestamp order by filename")
//	if err != nil {
//		slog.Error("Error executing query 'select filename from check_definitions where unixepoch() - interval_seconds_between_checks >= last_run_timestamp order by filename'", "err", err)
//	}
//	for filename := range filenames {
//		ch
//	}
//
//	return checkDefinitionsToRun
//}
