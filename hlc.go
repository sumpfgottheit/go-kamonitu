package main

import (
	"fmt"
	"github.com/fatih/color"
	"log/slog"
	"reflect"
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

	return nil
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
