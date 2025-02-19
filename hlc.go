package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/hashicorp/go-multierror"
	"log/slog"
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

func showConfigHlc(config *AppConfig) error {
	var content [][]string
	width := []int{0, 0, 0}

	m, order := structToMap(*config)
	fmt.Println()
	fmt.Printf("--> Applikationsconfigfile %s\n", AppConfigFilePath)
	contentAppConfig := make([][]string, len(order))
	for i, v := range order {
		width[0] = max(width[0], len(v))
		width[1] = max(width[1], len(m[v]))
		width[2] = max(width[2], len(appConfigSourceMap[v]))
		contentAppConfig[i] = []string{v, m[v], appConfigSourceMap[v]}
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
	for fileName, checkDefinition := range store.CheckDefinitions {
		m, order = structToMap(checkDefinition)
		for _, v := range order {
			width[0] = max(width[0], len(v))
			width[1] = max(width[1], len(m[v]))
			width[2] = max(width[2], len(store.CheckDefinitionSources[fileName][v]))
		}
	}

	sort2DSlice(contentAppConfig)
	PrintSimpleTableWithWidth([]string{"Key", "Value", "Source"}, contentAppConfig, width)
	fmt.Println()

	for fileName, checkDefinition := range store.CheckDefinitions {
		m, order = structToMap(checkDefinition)
		fmt.Printf("--> CheckDefinition %s\n", store.directory+"/"+fileName)
		content = make([][]string, len(order))
		for i, v := range order {
			content[i] = []string{v, m[v], store.CheckDefinitionSources[fileName][v]}
		}
		sort2DSlice(content)
		printSimpleTable([]string{"Key", "Value", "Source"}, content)
		fmt.Println()

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

func RunHlc(config *AppConfig) error {
	// Migrate Database
	err := migrateDatabase(config.DbFile())
	if err != nil {
		slog.Error("Error running database migrations", "err", err)
		return err
	}
	// Create CheckDefinitionStore
	store, err := makeCheckDefinitionFileStore(*config)
	if err != nil {
		return err
	}
	// Load CheckDefinitions
	err = store.LoadCheckDefinitionsFromDisk()
	slog.Info("Loaded Check Definitions", "count", len(store.CheckDefinitions))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, individualErr := range merr.Errors {
				slog.Warn("Error in check definition - Write it as failed check into database", "error", individualErr)
			}
		}
	}
	if len(store.CheckDefinitions) == 0 {
		slog.Warn("No Check Definitions found")
		return fmt.Errorf("no check definitions found")
	}

	// get Database Connection
	mydb, err := initDB(config.DbFile())
	if err != nil {
		slog.Error("Error initializing database", "err", err)
		return err
	}
	defer closeDB()

	store.db = mydb
	err = store.ensureCheckDefinitionsInDatabase()
	if err != nil {
		slog.Error("Error ensuring check definitions in database", "err", err)
		return err
	}

	return nil
}
