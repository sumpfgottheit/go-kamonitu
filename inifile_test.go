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
