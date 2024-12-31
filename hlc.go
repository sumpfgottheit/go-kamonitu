package main

import (
	"fmt"
	"github.com/fatih/color"
	"log/slog"
)

func validateConfigHlc(config *AppConfig) error {
	slog.Info("validate Configs")
	fmt.Printf("Validate Check Definition Configs aus '%s' ... \n", config.CheckDefinitionsDir)
	store := CheckDefinitionFileStore{directory: config.CheckDefinitionsDir}
	err := store.LoadCheckDefinitionsFromDisk()
	if err != nil {
		fmt.Println("Ungültige Konfiguration gefunden.")
		return err
	}
	fmt.Printf("Validate Check Definition Configs aus '%s' sind %v\n", config.CheckDefinitionsDir, color.GreenString("korrekt"))
	return nil
}
