package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

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
			} else if strings.HasPrefix(validationRules, "oneOf(") && strings.HasSuffix(validationRules, ")") {
				values := strings.Split(validationRules[6:len(validationRules)-1], ",")
				if !slices.Contains(values, v) {
					return fmt.Errorf("field %v must be one of %v", field.Name, values)
				}
			} else {
				return fmt.Errorf("validation rule %v not supported for field %v", validationRules, field.Name)
			}
		}
	}
	return nil
}
