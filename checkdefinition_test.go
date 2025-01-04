package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

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
	config := AppConfig{
		CheckDefinitionsDir: checkdir,
		ConfigDir:           d,
	}
	store, err := makeCheckDefinitionFileStore(config)
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
	store, err = makeCheckDefinitionFileStore(config)
	assert.NoError(t, err)
	err = store.LoadCheckDefinitionsFromDisk()
	assert.Error(t, err)
}
