package main

// Helper function to parse validation rules from a tag

import (
	"fmt"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

func main() {
	var debug bool
	var configFile string
	var appConfig *AppConfig

	// RootCmd setups logging and reads and validates the AppConfig file and sets the variable appConfig
	rootCmd := &cobra.Command{
		Use:   "kamonitu",
		Short: "Kamonitu is a configuration validation tool.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Use == "completion" {
				return nil
			}
			var err error
			if os.Getenv("KAMONITU_DEBUG") == "1" {
				debug = true
			}
			if configFile == DefaultConfigFilePath && os.Getenv(EnvVarConfigFilePath) != "" {
				configFile = os.Getenv(EnvVarConfigFilePath)
			}

			if debug {
				setupLogging(slog.LevelDebug, "")
			} else {
				setupLogging(slog.LevelInfo, "/var/log/kamonitu/")
			}

			appConfig, err = makeAppConfig(configFile)
			if err != nil {
				return err
			}
			slog.Debug("root.PersistentPreRunE successfull")
			return nil
		},
	}

	/*
		Globale Flags - Handling der Environment Variablen in rootCmd.PersistentPreRunE
	*/
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode (can also be set via KAMONITU_DEBUG environment variable)")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config-file", "f", DefaultConfigFilePath, "Path to config file (default can also be set via KAMONITU_CONFIG_FILE environment variable)")

	/*
		Subcommands
	*/

	/* validate-config */
	validateConfigCmd := &cobra.Command{
		Use:   "validate-config",
		Short: "Validates the configuration file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateConfigHlc(appConfig)
		},
	}
	rootCmd.AddCommand(validateConfigCmd)

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Command execution failed.", "error", err)
		fmt.Printf("Fehler: %v\n", err)
		fmt.Print("\033[31mFailed\033[0m\n") // Prints "Failed" in red
		os.Exit(1)
	}
}
