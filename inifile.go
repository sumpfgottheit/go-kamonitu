package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"reflect"
)

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
