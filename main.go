package main

// Helper function to parse validation rules from a tag

import (
	"errors"
	"fmt"
	"github.com/lmittmann/tint"
	"gopkg.in/ini.v1"
	"log/slog"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

func setupLogging(loglevel slog.Level) {
	w := os.Stderr
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			AddSource:  true,
			Level:      loglevel,
			TimeFormat: time.RFC3339,
		}),
	))
	slog.Info("Initialized Logging.", "Level", loglevel.String())

}

const (
	EnvVarIniFilePath  = "KAMONITU_INI_FILE"
	DefaultIniFilePath = "/etc/kamonitu/kamonitu.ini"
)

type AppConfig struct {
	VarDir                             string `db:"var_dir" ini:"required" validation:"writeableDirectory"`
	ConfigDir                          string `db:"config_dir" ini:"required" validation:"readableDirectory"`
	IntervalSecondsBetweenMainLoopRuns int    `db:"interval_seconds_between_main_loop_runs" validation:"within(1,60)"`
	CheckDefinitionsDir                string `db:"check_definitions_dir" validation:"readableDirectory"`
}

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

var appConfigDefaults = AppConfig{
	VarDir:                             "/var/lib/kamonitu",
	ConfigDir:                          "/etc/kamonitu",
	IntervalSecondsBetweenMainLoopRuns: 60,
	CheckDefinitionsDir:                "/etc/kamonitu/check_definitions",
}

func makeAppConfig() (*AppConfig, error) {
	path := os.Getenv(EnvVarIniFilePath)
	if path != "" {
		slog.Info("Inifile Path from EnvVar.", "path", path)
	} else {
		path = DefaultIniFilePath
		slog.Info("Nutze Default Ini File Path.", "path", path)

	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		slog.Error("Config File nicht vorhanden", "path", path)
		return nil, fmt.Errorf("Config File %v nicht vorhanden", path)
	}

	appconfig, err := ParseIniFileToStruct(path, appConfigDefaults)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("Parsed ini file.", "config", appconfig)

	appconfig.CheckDefinitionsDir = appconfig.ConfigDir + "/check_definitions"
	if err = ValidateStruct(appconfig); err != nil {
		return nil, err
	}

	return appconfig, nil
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

// Helper function to check if a file has the .ini extension
func isIniFile(filename string) bool {
	return len(filename) > 4 && filename[len(filename)-4:] == ".ini"
}

func mtimeForFile(filename string) (int64, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file %q: %v", filename, err)
	}
	return fileInfo.ModTime().Unix(), nil
}

func main() {

	loglevel := slog.LevelInfo
	if os.Getenv("KAMONITU_DEBUG") == "1" {
		loglevel = slog.LevelDebug
	}
	setupLogging(loglevel)
	slog.Info("Starting Kamonitu.")

	checkDefinitionDefaults, err := ParseIniFileToStruct(
		"_working_dir/etc/kamonitu/check_defaults.ini",
		checkDefinitionHardcodedDefaults)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("Parsed ini file.", "defaults", checkDefinitionDefaults)

	var defaultCheckDefinition = CheckDefinition{}
	defaultCheckDefinition.CheckDefinitionDefaults = *checkDefinitionDefaults
	check_swap, err := ParseIniFileToStruct("_working_dir/etc/kamonitu/check_definitions/swap.ini", defaultCheckDefinition)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("Parsed ini file.", "swap", check_swap)

	_, err = makeAppConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("Parsed ini file.", "config", appConfigDefaults)

}

const (
	intWithinRegex = "intWithinRegex"
)

var validationRegexCacheMap map[string]*regexp.Regexp

// ValidateStruct Verwendet Struct Tag "validation"
// - strings: readableDirectory, writeableDirectory
// - int: within(lower, upper)
func ValidateStruct(input any) error {

	if len(validationRegexCacheMap) == 0 {
		validationRegexCacheMap = make(map[string]*regexp.Regexp)
		validationRegexCacheMap[intWithinRegex] = regexp.MustCompile(`^within\(([-+]?\d+),([-+]?\d+)\)`)
	}
	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.New("input must be a pointer to a struct")
	}

	val = val.Elem()
	typ := val.Type()

	// Iterate over fields to validate based on "ini" tag
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		validationRules := field.Tag.Get("validation")
		if validationRules == "" {
			continue
		}
		slog.Debug("Validation Rules", "field", field.Name, "validationRules", validationRules)

		// Check integer and perform validations specified in the "validation" tag
		switch field.Type.Kind() {
		case reflect.Int:
			v := value.Int()

			regex := validationRegexCacheMap[intWithinRegex]
			matches := regex.FindStringSubmatch(validationRules)
			if len(matches) == 3 {
				lower, _ := strconv.Atoi(matches[1])
				upper, _ := strconv.Atoi(matches[2])
				if v < int64(lower) || v > int64(upper) {
					return fmt.Errorf("field %v muss innerhalb on %v und %v liegen, ist aber %v", field.Name, lower, upper, v)
				}
			} else {
				return fmt.Errorf("validation rule %v not supported for field %v", validationRules, field.Name)
			}
		case reflect.String:
			v := value.String()
			if validationRules == "readableDirectory" {
				info, err := os.Stat(v)
				if err != nil || !info.IsDir() {
					return fmt.Errorf("field %v must be a readable directory, but %v is not accessible or does not exist", field.Name, v)
				}
				// Check directory readability by attempting to open it
				dir, err := os.Open(v)
				if err != nil {
					dir.Close()
					return fmt.Errorf("field %v must be a readable directory, but %v is not readable", field.Name, v)
				}
				_, err = dir.Readdirnames(1) // Attempt to read a single entry
				if err != nil {
					dir.Close()
					return fmt.Errorf("field %v must be a readable directory, but %v is not readable", field.Name, v)
				}
				dir.Close()
			} else if validationRules == "writeableDirectory" {
				info, err := os.Stat(v)
				if err != nil || !info.IsDir() {
					return fmt.Errorf("field %v must be a writable directory, but %v is not accessible or does not exist", field.Name, v)
				}
				// Check directory writability by attempting to create and remove a temporary file
				tempFile := v + "/.tmp_write_test"
				file, err := os.Create(tempFile)
				if err != nil {
					return fmt.Errorf("field %v must be a writable directory, but %v is not writable", field.Name, v)
				}
				file.Close()
				err = os.Remove(tempFile)
				if err != nil {
					return fmt.Errorf("field %v must be a writable directory, but the temporary file in %v could not be removed", field.Name, v)
				}
			} else {
				return fmt.Errorf("validation rule %v not supported for field %v", validationRules, field.Name)
			}
		}
	}
	return nil
}

func ParseIniFileToStruct[T AppConfig | CheckDefinitionDefaults | CheckDefinition](filePath string, defaults T) (*T, error) {
	// Load the ini file
	iniFile, err := ini.Load(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ini file: %v", err)
	}

	section := iniFile.Section("")
	if section == nil {
		return nil, fmt.Errorf("Nur Key=Val ohne Sectionnamen erlaubt")
	}

	// Create default struct instance to populate
	configInstance := defaults
	configPointer := &configInstance

	// Process the 'CheckDefinitionDefaults' struct to validate and map
	typ := reflect.TypeOf(*configPointer)
	val := reflect.ValueOf(configPointer).Elem()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		iniKey := field.Tag.Get("db")    // The key in the ini file
		required := field.Tag.Get("ini") // Check if it's marked as "required"

		// Skip if no 'db' tag present
		if iniKey == "" {
			continue
		}

		// Validate required keys
		if required == "required" && !section.HasKey(iniKey) {
			return nil, fmt.Errorf("%v: missing required key: %v", filePath, iniKey)
		}

		// If the key exists, map it to the struct
		if section.HasKey(iniKey) {
			key := section.Key(iniKey)

			// Check field type and assign value properly
			switch field.Type.Kind() {
			case reflect.Int:
				intValue, err := key.Int()
				if err != nil {
					return nil, fmt.Errorf("%v: invalid integer for key: %v, value: %v", filePath, iniKey, key)
				}
				val.Field(i).SetInt(int64(intValue))
			case reflect.String:
				stringValue := key.String()
				val.Field(i).SetString(stringValue)
			default:
				return nil, fmt.Errorf("unsupported key type for key: %v", iniKey)
			}
		} else {
			// Fallback to hardcoded check_defaults.ini if not present (optional keys)
			defaultVal := val.Field(i).Interface()
			val.Field(i).Set(reflect.ValueOf(defaultVal))
		}
	}
	return configPointer, nil
}
