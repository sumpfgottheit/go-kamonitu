package main

import (
	"fmt"
	"github.com/fatih/color"
	"log/slog"
	"reflect"
	"sort"
	"strings"
)

func validateConfigHlc(config *AppConfig) error {
	slog.Info("validate Configs")
	fmt.Printf("Validate Check Definition Configs aus '%s' ... \n", config.CheckDefinitionsDir)
	store, err := makeCheckDefinitionFileStore(*config)
	if err != nil {
		return err
	}
	err = store.LoadCheckDefinitionsFromDisk()
	if err != nil {
		return err
	}
	fmt.Printf("Validate Check Definition Configs aus '%s' sind %v\n", config.CheckDefinitionsDir, color.GreenString("korrekt"))
	return nil
}

func DumpConfigHlc(config *AppConfig) error {

	m, order := structToMap(*config)
	for _, v := range order {
		fmt.Printf("%v=%v [%s]\n", v, m[v], appConfigSourceMap[v])
	}

	store, err := makeCheckDefinitionFileStore(*config)
	if err != nil {
		slog.Error("Failed to make CheckDefinitionFileStore", "error", err)
		return err
	}
	err = store.LoadCheckDefinitionsFromDisk()
	if err != nil {
		slog.Error("Failed to load CheckDefinitions from Disk", "error", err)
		return err
	}
	return nil
}

func ShowDefaultsHlc(config *AppConfig) error {

	_, err := makeCheckDefinitionFileStore(*config)
	if err != nil {
		slog.Error("Failed to make CheckDefinitionFileStore", "error", err)
		return err
	}

	// Retrieve keys from appConfigDefaultMap (assuming it's defined elsewhere)
	keys := append(sortedKeys(appConfigDefaultMap), sortedKeys(checkDefinitionsDefaultMapFromFile)...)
	longest_key := 0
	for _, key := range keys {
		if len(key) > longest_key {
			longest_key = len(key)
		}
	}

	fmt.Println()
	fmt.Println("--> Hardcoded Defaults für das Configfile in /etc/kamonitu/kamonitu.ini [KAMONITU_CONFIG_FILE]")
	keys = sortedKeys(appConfigDefaultMap)
	for _, key := range keys {
		fmt.Printf("%-*s = %v\n", longest_key, key, appConfigDefaultMap[key])
	}

	fmt.Println()
	fmt.Println("--> Effective Defaults für die Check Definitionen")
	keys = sortedKeys(checkDefinitionDefaultsMap)
	for _, key := range keys {
		fmt.Printf("%-*s = %v\n", longest_key, key, checkDefinitionDefaultsMap[key])
	}

	fmt.Println()
	fmt.Println("--> Defaults für die Check Definitionen aus /etc/kamonitu/check_defaults.ini")
	keys = sortedKeys(checkDefinitionsDefaultMapFromFile)
	for _, key := range keys {
		fmt.Printf("%-*s = %v\n", longest_key, key, checkDefinitionsDefaultMapFromFile[key])
	}

	fmt.Println()
	fmt.Println("--> Hardcoded Defaults für die Check Definitionen")
	keys = sortedKeys(hardCodedcheckDefinitionDefaultsMap)
	for _, key := range keys {
		fmt.Printf("%-*s = %v\n", longest_key, key, hardCodedcheckDefinitionDefaultsMap[key])
	}

	fmt.Println()

	return nil
}

func DescribeConfigFilesHlc(appConfig *AppConfig) error {
	fmt.Printf("Das Applikationsconfigfile ist definiert via:\n")
	fmt.Printf("* Parameter -f / --config-file\n")
	fmt.Printf("* Umgebungsvaraible KAMONITO_CONFIG_FILE\n")
	fmt.Printf("* /etc/kamonitu/kamonitu.ini\n")
	fmt.Printf("\n")
	fmt.Println("Required \"yes\" bedeutet, dass dieser Parameter im Configfile gesetzt werden muss.")
	fmt.Println("Bei \"no\" wird der default Wert verwendet. Dieser kann mit dem Befehl 'kamonitu show-defaults' abgerufen werden.")
	fmt.Println()

	tags := getStructTags(appConfig, []string{"db", "ini", "validation"})
	content := make([][]string, 0)
	for _, v := range tags {
		if len(v) == 0 {
			continue
		}
		ini := "yes"
		if v["ini"] == "" {
			ini = "no"
		}
		if v["ini"] == "not_allowed" {
			continue
		}
		content = append(content, []string{v["db"], ini, v["validation"]})
	}
	sort2DSlice(content)
	printSimpleTable([]string{"Key", "Required", "Validation"}, content)
	fmt.Println()

	fmt.Println("Die Check Definitionen werden in ini Dateien im Verzeichnis $config_dir/check_definition/*.ini gespeichert.")
	fmt.Println("Defaultwerte für die Checks können in der Datei $config_dir/check_defaults.ini definiert werden.")
	fmt.Println("Für Werte, die weder in den Check Definitionen, noch in der Defaultdatei definiert werden, wird der hardcoded Defaultwer verwendet.")
	fmt.Println("Mittels 'kamonitu show-defaults' werden die aktuellen Defaultwerte angezeigt.")
	fmt.Println()

	tags = getStructTags(CheckDefinition{}, []string{"db", "ini", "validation"})
	content = make([][]string, 0)
	for _, v := range tags {
		if len(v) == 0 {
			continue
		}
		ini := "yes"
		if v["ini"] == "" {
			ini = "no"
		}
		if v["ini"] == "not_allowed" {
			continue
		}
		content = append(content, []string{v["db"], ini, v["validation"]})
	}
	sort2DSlice(content)
	printSimpleTable([]string{"Key", "Required", "Validation"}, content)
	fmt.Println()

	return nil
}

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

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

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
