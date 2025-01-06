package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// sort2DSlice sorts a 2D slice of strings lexicographically based on the first element of each inner slice.
func sort2DSlice(content [][]string) {
	sort.Slice(content, func(i, j int) bool {
		// edge cases
		if len(content[i]) == 0 && len(content[j]) == 0 {
			return false // two empty slices - so one is not less than other i.e. false
		}
		if len(content[i]) == 0 || len(content[j]) == 0 {
			return len(content[i]) == 0 // empty slice listed "first" (change to != 0 to put them last)
		}

		// both slices len() > 0, so can test this now:
		return content[i][0] < content[j][0]
	})
}

// PrintSimpleTableWithWidth prints a table with a specified header, content, and column widths.
// Each column is formatted to fit the width defined in the columnWidths slice.
func PrintSimpleTableWithWidth(header []string, content [][]string, columnWidths []int) {
	// Print the header
	for i, col := range header {
		fmt.Printf("%-*s ", columnWidths[i], col)
	}
	fmt.Println()

	// Print an underline for the header
	for _, width := range columnWidths {
		fmt.Printf("%s ", strings.Repeat("-", width))
	}
	fmt.Println()

	// Print each row of content
	for _, row := range content {
		for i, cell := range row {
			fmt.Printf("%-*s ", columnWidths[i], cell)
		}
		fmt.Println()
	}
}

// printSimpleTable prints a simple table with headers and rows of content aligned column-wise.
func printSimpleTable(header []string, content [][]string) {
	// Calculate column widths based on header and content
	columnWidths := make([]int, len(header))
	for i, col := range header {
		columnWidths[i] = len(col)
	}
	for _, row := range content {
		for i, cell := range row {
			if len(cell) > columnWidths[i] {
				columnWidths[i] = len(cell)
			}
		}
	}

	// Print the header
	for i, col := range header {
		fmt.Printf("%-*s ", columnWidths[i], col)
	}
	fmt.Println()

	// Print an underline for the header
	for _, width := range columnWidths {
		fmt.Printf("%s ", strings.Repeat("-", width))
	}
	fmt.Println()

	// Print each row of content
	for _, row := range content {
		for i, cell := range row {
			fmt.Printf("%-*s ", columnWidths[i], cell)
		}
		fmt.Println()
	}
}

// structToMap converts a struct to a map with keys as snake_case field names and values as string representations of the fields.
// Returns the map and a slice of ordered field names in the same snake_case format.
func structToMap(s interface{}) (result map[string]string, orderedFieldName []string) {
	result = make(map[string]string)

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	fieldNames := getFieldNamesForStruct(s)

	typ := val.Type()
	for _, fieldName := range fieldNames {
		field, _ := typ.FieldByName(fieldName)
		ini_key := field.Tag.Get("db")
		if ini_key == "" {
			continue
		}
		iniMapFieldName := camelCaseToSnakeCase(fieldName)
		fieldValue := val.FieldByName(field.Name)
		// Convert field value to string
		result[iniMapFieldName] = fmt.Sprintf("%v", fieldValue.Interface())
	}

	orderedFieldName = make([]string, len(fieldNames))
	for i, v := range fieldNames {
		orderedFieldName[i] = camelCaseToSnakeCase(v)
	}
	return result, orderedFieldName
}

// camelCaseToSnakeCase converts a string from camelCase to snake_case format.
// Handles uppercase letters by inserting an underscore before each and converts all letters to lowercase.
func camelCaseToSnakeCase(s string) string {
	var result string

	for i, v := range s {
		if i > 0 && v >= 'A' && v <= 'Z' {
			result += "_"
		}

		result += string(v)
	}

	return strings.ToLower(result)

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

// sortedKeys returns a sorted slice of all keys from the given map.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

// getStructTags extracts specified struct tag values for each field in the provided struct and returns them as a nested map.
func getStructTags(input interface{}, tags []string) map[string]map[string]string {

	// Get the type of the input
	t := reflect.TypeOf(input)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil // Not a struct
	}

	// Iterate over the fields
	result := make(map[string]map[string]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldTags := make(map[string]string)

		// Extract tags from each field
		for _, key := range tags { // Add more tag keys if needed
			tagValue := field.Tag.Get(key)
			if tagValue != "" {
				fieldTags[key] = tagValue
			}
		}

		// Add to the result map
		result[field.Name] = fieldTags
	}

	return result
}
