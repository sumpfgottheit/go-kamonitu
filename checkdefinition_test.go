package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestLoadCheckDefinitionDefaults(t *testing.T) {
	fileName := "defaults.ini"
	type TestStruct struct {
		name             string
		defaults         map[string]string
		expectedDefaults CheckDefinitionDefaults
		expectError      bool
	}
	tests := []TestStruct{
		{
			name:             "Empty file",
			defaults:         map[string]string{},
			expectedDefaults: checkDefinitionHardcodedDefaults,
			expectError:      false,
		},
		{
			name: "Defaults und Hardcoded Defaults 1:1",
			defaults: map[string]string{
				"interval_seconds_between_checks":        "60",
				"delay_seconds_before_first_check":       "0",
				"timeout_seconds":                        "60",
				"stop_checking_after_number_of_timeouts": "3",
			},
			expectedDefaults: checkDefinitionHardcodedDefaults,
			expectError:      false,
		},
		{
			name: "Defaults und Hardcoded Defaults 1:1 aber nur ein Paramater gesetzts",
			defaults: map[string]string{
				"interval_seconds_between_checks": "60",
			},
			expectedDefaults: checkDefinitionHardcodedDefaults,
			expectError:      false,
		},
		{
			name: "Überschreibe interval_seconds_between_checks",
			defaults: map[string]string{
				"interval_seconds_between_checks": "50",
			},
			expectedDefaults: CheckDefinitionDefaults{
				IntervalSecondsBetweenChecks:      50,
				DelaySecondsBeforeFirstCheck:      0,
				TimeoutSeconds:                    60,
				StopCheckingAfterNumberOfTimeouts: 3,
			},
			expectError: false,
		},
		{
			name: "Ungültige Wert",
			defaults: map[string]string{
				"interval_seconds_between_checks": "foo",
			},
			expectedDefaults: CheckDefinitionDefaults{},
			expectError:      true,
		},
		{
			name: "Ungültiger Key",
			defaults: map[string]string{
				"interval_seconds_between_checks": "30",
				"foo":                             "30",
			},
			expectedDefaults: CheckDefinitionDefaults{},
			expectError:      true,
		},
	}

	file := t.TempDir() + "/" + fileName
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup after each test case
			os.Remove(file)

			// Write the content of tt.defaults to the file in INI format
			f, err := os.Create(file)
			assert.NoError(t, err)
			defer f.Close()
			for key, value := range tt.defaults {
				_, err := f.WriteString(key + " = " + value + "\n")
				assert.NoError(t, err)
			}

			// Initialize the file store
			store := &CheckDefinitionFileStore{}
			err = store.LoadCheckDefinitionDefaults(file)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedDefaults, store.checkDefinitionDefaults)

		})
	}
}

func TestLoadDefinitionsFilesMtimeMap(t *testing.T) {
	type TestCase struct {
		name           string
		files          map[string]interface{} // File name mapped to modification time (integer) or nil (for non-regular files)
		expectedMtimes map[string]int64
		expectError    bool
	}

	tests := []TestCase{
		{
			name: "Valid INI files",
			files: map[string]interface{}{
				"file1.ini": int64(1631021765),
				"file2.ini": int64(1631051765),
			},
			expectedMtimes: map[string]int64{
				"file1.ini": 1631021765,
				"file2.ini": 1631051765,
			},
			expectError: false,
		},
		{
			name:           "Empty directory",
			files:          map[string]interface{}{},
			expectedMtimes: map[string]int64{},
			expectError:    false,
		},
		{
			name: "Directory with non-INI files",
			files: map[string]interface{}{
				"file1.txt": int64(1631051765),
				"file2.log": int64(1631054765),
				"valid.ini": int64(1631061765),
			},
			expectedMtimes: map[string]int64{
				"valid.ini": 1631061765,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create files in the temp directory based on test case
			for name, modTime := range tt.files {
				if modTime == nil {
					continue
				}
				path := tempDir + "/" + name
				file, err := os.Create(path)
				assert.NoError(t, err)
				file.Close()

				// Update modification time, if specified
				if mt, ok := modTime.(int64); ok {
					err := os.Chtimes(path, time.Unix(mt, 0), time.Unix(mt, 0))
					assert.NoError(t, err)
				}
			}

			store := CheckDefinitionFileStore{directory: tempDir}

			// Call the method and verify results
			err := store.LoadDefinitionsFilesMtimeMap()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMtimes, store.definitionsFilesMtimes)
			}
		})
	}
}

func TestLoadDefinitionsFilesMtimeMapInvalidDirectory(t *testing.T) {
	tempDir := t.TempDir()
	store := CheckDefinitionFileStore{directory: tempDir + "/foo/bar"}
	err := store.LoadDefinitionsFilesMtimeMap()
	assert.Error(t, err)
}

func TestMakeCheckDefinitionFileStore(t *testing.T) {
	d := t.TempDir()
	defaults := map[string]string{
		"interval_seconds_between_checks": "30",
	}
	defaultFile := d + "/defaults.ini"

	// Write the content of tt.defaults to the file in INI format
	f, err := os.Create(defaultFile)
	assert.NoError(t, err)
	defer f.Close()
	for key, value := range defaults {
		_, err := f.WriteString(key + " = " + value + "\n")
		assert.NoError(t, err)
	}

	checkdir := d + "/checks"
	err = os.Mkdir(checkdir, 0755)
	assert.NoError(t, err)

	checks := map[string]map[string]string{
		"swap.ini": {
			"check_command":                    "/usr/bin/check_swap -w 10% -c 5%",
			"delay_seconds_before_first_check": "10",
		},
		"cpu.ini": {
			"check_command": "/usr/bin/check_cpu -w 50% -c 75%",
		},
	}
	for name, check := range checks {
		path := checkdir + "/" + name
		f, err := os.Create(path)
		assert.NoError(t, err)
		defer f.Close()
		for key, value := range check {
			_, err := f.WriteString(key + " = " + value + "\n")
			assert.NoError(t, err)
		}
	}

	store, err := makeCheckDefinitionFileStore(checkdir, defaultFile)
	assert.NoError(t, err)
	err = store.LoadCheckDefinitionsFromDisk()
	assert.NoError(t, err)

	assert.Equal(t, len(store.CheckDefinitions), 2)
	assert.Equal(t, store.CheckDefinitions["swap.ini"].CheckCommand, "/usr/bin/check_swap -w 10% -c 5%")
	assert.Equal(t, store.CheckDefinitions["swap.ini"].DelaySecondsBeforeFirstCheck, 10)
	assert.Equal(t, store.CheckDefinitions["swap.ini"].IntervalSecondsBetweenChecks, 30)
	assert.Equal(t, store.CheckDefinitions["swap.ini"].TimeoutSeconds, 60)

	assert.Equal(t, store.CheckDefinitions["cpu.ini"].CheckCommand, "/usr/bin/check_cpu -w 50% -c 75%")
	assert.Equal(t, store.CheckDefinitions["cpu.ini"].DelaySecondsBeforeFirstCheck, 0)
	assert.Equal(t, store.CheckDefinitions["cpu.ini"].IntervalSecondsBetweenChecks, 30)
	assert.Equal(t, store.CheckDefinitions["cpu.ini"].TimeoutSeconds, 60)

	checks["swap.ini"]["timeout_seconds"] = "nix"
	for name, check := range checks {
		path := checkdir + "/" + name
		f, err := os.Create(path)
		assert.NoError(t, err)
		defer f.Close()
		for key, value := range check {
			_, err := f.WriteString(key + " = " + value + "\n")
			assert.NoError(t, err)
		}
	}
	store, err = makeCheckDefinitionFileStore(checkdir, defaultFile)
	assert.NoError(t, err)
	err = store.LoadCheckDefinitionsFromDisk()
	assert.Error(t, err)
}
