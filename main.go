package main

import (
	"fmt"
	"github.com/lmittmann/tint"
	"gopkg.in/ini.v1"
	"log/slog"
	"os"
	"reflect"
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

type AppConfig struct {
	VarDir                             string `db:"var_dir" ini:"required"`
	ConfigDir                          string `db:"config_dir" ini:"required"`
	IntervalSecondsBetweenMainLoopRuns int    `db:"interval_seconds_between_main_loop_runs" ini:"optional"`
}

var appConfigDefaults = AppConfig{
	VarDir:                             "/var/lib/kamonitu",
	ConfigDir:                          "/etc/kamonitu",
	IntervalSecondsBetweenMainLoopRuns: 60,
}

func ParseIniFileToStruct(filePath string, defaults map[string]string) (*AppConfig, error) {
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
	appConfig := &AppConfig{}

	// Process the 'CheckDefinitionDefaults' struct to validate and map
	typ := reflect.TypeOf(*appConfig)
	val := reflect.ValueOf(appConfig).Elem()

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

		// Wenn das Inifile einen Wert nicht hat, dann kann dieser aus Defaults geholt werden
		defaultValue, exists := defaults[iniKey]
		if !section.HasKey(iniKey) && exists {
			section.Key(iniKey).SetValue(defaultValue)
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
	return appConfig, nil
}

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {

	loglevel := slog.LevelInfo
	if os.Getenv("KAMONITU_DEBUG") == "1" {
		loglevel = slog.LevelDebug
	}
	setupLogging(loglevel)

	appConfigDefaults_map, e := StructToMap(appConfigDefaults)
	if e != nil {
		slog.Error(e.Error())
		os.Exit(1)
	}

	slog.Info("Converted struct to map.", "defaults", appConfigDefaults_map)
	config, e := ParseIniFileToStruct("_working_dir/etc/kamonitu.ini", appConfigDefaults_map)
	if e != nil {
		slog.Error(e.Error())
		os.Exit(1)
	}
	slog.Info("Parsed ini file.", "a", config)
}

// StructToMap converts a struct to a map[string]string
func StructToMap(input interface{}) (map[string]string, error) {
	// Validate that input is a struct
	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct")
	}

	// Prepare the resulting map
	result := make(map[string]string)
	t := reflect.TypeOf(input)

	// Iterate over each field of the struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get the "db" tag or the field name
		key := field.Tag.Get("db")
		if key == "" {
			key = field.Name // Use the field name if "db" tag is not present
		}

		// Convert the field value to a string
		switch value.Kind() {
		case reflect.String:
			result[key] = value.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[key] = strconv.FormatInt(value.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			result[key] = strconv.FormatUint(value.Uint(), 10)
		case reflect.Bool:
			result[key] = strconv.FormatBool(value.Bool())
		case reflect.Float32, reflect.Float64:
			result[key] = strconv.FormatFloat(value.Float(), 'f', -1, 64)
		default:
			return nil, fmt.Errorf("unsupported field type: %v for field %v", value.Kind(), field.Name)
		}
	}

	return result, nil
}
