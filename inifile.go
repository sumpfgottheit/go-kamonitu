package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
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

// getFieldNamesForStruct retrieves all field names of a struct, including those from embedded structs.
func getFieldNamesForStruct(s interface{}) []string {
	var fieldNames []string
	typ := reflect.TypeOf(s)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			// Recursively get fields from embedded structs
			embeddedFields := getFieldNamesForStruct(reflect.New(field.Type).Interface())
			fieldNames = append(fieldNames, embeddedFields...)
		} else {
			fieldNames = append(fieldNames, field.Name)
		}
	}
	return fieldNames
}

// ParseStringMapToStruct maps a string map to a struct, using the "db" tag in struct fields
// to determine the mapping. It supports string and int types.
func ParseStringMapToStruct[T any](iniMap map[string]string, defaults T) (*T, error) {
	// Create a new instance of the struct type
	configInstance := defaults
	configPointer := &configInstance

	typ := reflect.TypeOf(configInstance)
	val := reflect.ValueOf(configPointer).Elem()

	// Iterate over each field in the struct
	for _, fieldName := range getFieldNamesForStruct(configInstance) {
		field, exists := typ.FieldByName(fieldName)
		if !exists {
			return nil, fmt.Errorf("field %v not found in struct", fieldName)
		}
		iniKey := field.Tag.Get("db")

		// Check if iniKey exists in the map
		if value, exists := iniMap[iniKey]; exists {
			switch field.Type.Kind() {
			case reflect.Int: // Handle integer fields
				intValue, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("invalid integer for key %v: value %v", iniKey, value)
				}
				val.FieldByName(fieldName).SetInt(int64(intValue))
			case reflect.String: // Handle string fields
				val.FieldByName(fieldName).SetString(value)
			default: // Unsupported types
				return nil, fmt.Errorf("unsupported type for field %v", field.Name)
			}
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
