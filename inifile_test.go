package main

import (
	"os"
	"reflect"
	"testing"
)

func TestReadIniFile(t *testing.T) {
	tests := []struct {
		name      string
		fileData  string
		want      map[string]string
		expectErr bool
	}{
		{
			name: "valid ini file",
			fileData: `
key1 = value1
key2 = value2
; comment line
# another comment line
key3=value3
`,
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			expectErr: false,
		},
		{
			name: "empty ini file",
			fileData: `
`,
			want:      map[string]string{},
			expectErr: false,
		},
		{
			name: "line with empty key",
			fileData: `
= value1
key2 = value2
`,
			want:      nil,
			expectErr: true,
		},
		{
			name: "line with empty value",
			fileData: `
key1 =
key2 = value2
`,
			want:      nil,
			expectErr: true,
		},
		{
			name: "missing equals sign",
			fileData: `
key1 value1
key2 = value2
`,
			want:      nil,
			expectErr: true,
		},
		{
			name: "Keys with whitespace",
			fileData: `
key 1  = value1
key2 = value2
`,
			want:      nil,
			expectErr: true,
		},
		{
			name: "ini file with whitespace",
			fileData: `

key1 = value1

key2 = value2

`,
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpFile, err := os.CreateTemp("", "test-*.ini")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write test data to the temporary file
			if _, err := tmpFile.WriteString(tt.fileData); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}

			// Close the temporary file
			if err := tmpFile.Close(); err != nil {
				t.Fatalf("failed to close temp file: %v", err)
			}

			// Call the function being tested
			got, err := readIniFile(tmpFile.Name())

			// Validate outcomes
			if (err != nil) != tt.expectErr {
				t.Errorf("readIniFile() error = %v, expectErr = %v", err, tt.expectErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readIniFile() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestGetFieldNamesForStruct(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{
			name:  "empty struct",
			input: struct{}{},
			want:  []string{},
		},
		{
			name: "struct with fields",
			input: struct {
				Field1 string
				Field2 int
			}{},
			want: []string{"Field1", "Field2"},
		},
		{
			name: "struct with embedded anonymous field",
			input: struct {
				Field1 string
				EmbeddedStruct
			}{},
			want: []string{"Field1", "FieldA", "FieldB"},
		},
		{
			name: "pointer to a struct",
			input: &struct {
				Field1 string
				Field2 int
			}{},
			want: []string{"Field1", "Field2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFieldNamesForStruct(tt.input)
			if got == nil && len(tt.want) == 0 {
				got = []string{}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFieldNamesForStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

type EmbeddedStruct struct {
	FieldA int
	FieldB string
}

func TestParseStringMapToStruct(t *testing.T) {

	type User struct {
		Username string `db:"username"`
		Age      int    `db:"age"`
	}

	tests := []struct {
		name      string
		iniMap    map[string]string
		defaults  User
		want      User
		expectErr bool
	}{
		{
			name: "valid mapping to struct",
			iniMap: map[string]string{
				"username": "John Doe",
				"age":      "42",
			},
			defaults: User{},
			want: User{
				Username: "John Doe",
				Age:      42,
			},
			expectErr: false,
		},
		{
			name: "missing key in map",
			iniMap: map[string]string{
				"username": "John Doe",
			},
			defaults: User{},
			want: User{
				Username: "John Doe",
				Age:      0,
			},
			expectErr: false,
		},
		{
			name: "struct with unsupported field type",
			iniMap: map[string]string{
				"age": "4.2",
			},
			defaults:  User{},
			want:      User{},
			expectErr: true,
		},
		{
			name: "invalid int value in map",
			iniMap: map[string]string{
				"age": "thirty",
			},
			defaults:  User{},
			want:      User{},
			expectErr: true,
		},
		{
			name: "unused keys in map",
			iniMap: map[string]string{
				"username":  "john_doe",
				"unusedKey": "some_value",
			},
			defaults:  User{},
			want:      User{},
			expectErr: true,
		},
		{
			name:      "empty string map",
			iniMap:    map[string]string{},
			defaults:  User{},
			want:      User{},
			expectErr: false,
		},
		{
			name:   "empty string map with defaults",
			iniMap: map[string]string{},
			defaults: User{
				Username: "John Doe",
				Age:      42,
			},
			want: User{
				Username: "John Doe",
				Age:      42,
			},
			expectErr: false,
		},
		{
			name:   "empty string map with defaults - 2",
			iniMap: map[string]string{},
			defaults: User{
				Age: 44,
			},
			want: User{
				Username: "",
				Age:      44,
			},
			expectErr: false,
		},
		{
			name: "empty struct",
			iniMap: map[string]string{
				"username": "Arno Nym",
			},
			defaults: User{},
			want: User{
				Username: "Arno Nym",
				Age:      0,
			},
			expectErr: false,
		},
		{
			name: "valid mapping with tag mismatch",
			iniMap: map[string]string{
				"wrong_key": "unknown",
			},
			defaults:  User{},
			want:      User{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStringMapToStruct(tt.iniMap, tt.defaults)

			// Validate error expectation
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseStringMapToStruct() error = %v, expectErr = %v", err, tt.expectErr)
				return
			}

			// Validate returned struct
			if !tt.expectErr && !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("ParseStringMapToStruct() got = %v, want = %v", got, tt.want)
			}
		})
	}
}
