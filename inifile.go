package main

import (
	"bufio"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"reflect"
	"strings"
)

// readIniFile reads an INI file from the given file path and returns a map of key-value pairs.
// The function skips empty lines and lines starting with ';' or '#' as comments - after Trimming the line.
// Returns an error if the file cannot be opened or contains parsing issues.
func readIniFile(filePath string) (map[string]string, error) {
	// Open the ini file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ini file: %v", err)
	}
	defer file.Close()

	// Define a map to store key-value pairs
	config := make(map[string]string)

	// Read file line by line and parse it
	reader := bufio.NewScanner(file)
	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())
		if len(line) == 0 || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			// Skip empty lines and commented lines
			continue
		}
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			if strings.Contains(key, " ") {
				return nil, fmt.Errorf("invalid line in ini file: %v. Key contains spaces", line)
			}
			value := strings.TrimSpace(parts[1])
			if key == "" {
				return nil, fmt.Errorf("invalid line in ini file: %v. Key is empty", line)
			}
			if value == "" {
				return nil, fmt.Errorf("invalid line in ini file: %v. Value is empty", line)
			}
			config[key] = value
		} else {
			return nil, fmt.Errorf("invalid line in ini file: %v. No '=' found", line)
		}
	}

	if err = reader.Err(); err != nil {
		return nil, fmt.Errorf("error reading ini file: %v", err)
	}
	return config, nil
}

func collectFieldNames(t reflect.Type, m map[string]struct{}) {

	// Return if not struct or pointer to struct.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	// Iterate through fields collecting names in map.
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		m[sf.Name] = struct{}{}

		// Recurse into anonymous fields.
		if sf.Anonymous {
			collectFieldNames(sf.Type, m)
		}
	}
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

	m := make(map[string]struct{})
	collectFieldNames(typ, m)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		iniKey := field.Tag.Get("db")    // The key in the ini file
		required := field.Tag.Get("ini") // Check if it's marked as "required"

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

// Helper function to check if a file has the .ini extension
func isIniFile(filename string) bool {
	return len(filename) > 4 && filename[len(filename)-4:] == ".ini"
}

// Helper function to check if a file is readable
func isFileReadable(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()
	return true
}

func mtimeForFile(filename string) (int64, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file %q: %v", filename, err)
	}
	return fileInfo.ModTime().Unix(), nil
}
