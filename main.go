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
			if cmd.Use == "completion" || cmd.Use == "wip" || cmd.Use == "version" || cmd.Use == "help" {
				return nil
			}
			var err error
			if os.Getenv("KAMONITU_DEBUG") == "1" {
				debug = true
			}
			if configFile == DefaultConfigFilePath && os.Getenv(EnvVarConfigFilePath) != "" {
				configFile = os.Getenv(EnvVarConfigFilePath)
			}
			AppConfigFilePath = configFile

			if debug {
				setupLogging(slog.LevelDebug, "")
			} else {
				setupLogging(slog.LevelInfo, "/var/log/kamonitu/")
			}

			appConfig, err = makeAppConfig(configFile)
			if err != nil {
				slog.Error("Failed to make AppConfig", "error", err)
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

	/* dump-config */
	showConfigCmd := &cobra.Command{
		Use:   "show-config",
		Short: "Shows the current configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showConfigHlc(appConfig)
		},
	}
	rootCmd.AddCommand(showConfigCmd)

	/* show-defaults */
	ShowDefaultsCmd := &cobra.Command{
		Use:   "show-defaults",
		Short: "Show the defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ShowDefaultsHlc(appConfig)
		},
	}
	rootCmd.AddCommand(ShowDefaultsCmd)

	/* describe-configfile */
	DescribeConfigsFilesCmd := &cobra.Command{
		Use:   "describe-configfiles",
		Short: "Beschreibung der Konfigurationsdateien",
		RunE: func(cmd *cobra.Command, args []string) error {
			return DescribeConfigFilesHlc(appConfig)
		},
	}
	rootCmd.AddCommand(DescribeConfigsFilesCmd)

	/* run */
	RunCmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"run"},
		Short:   "Starte Kamonitu",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunHlc(appConfig)
		},
	}
	rootCmd.AddCommand(RunCmd)

	wipCmd := &cobra.Command{
		Use:   "wip",
		Short: "WIP",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("WIP\n")

			definition, sources, err := loadSingleCheckDefinitionFromFile("_working_dir/etc/kamonitu/check_definitions/swap.ini")
			fmt.Println(definition)
			fmt.Println(sources)
			fmt.Println(err)

			return nil
		},
	}
	rootCmd.AddCommand(wipCmd)

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Command execution failed.", "error", err)
		fmt.Printf("Fehler: %v\n", err)
		fmt.Print("\033[31mFailed\033[0m\n") // Prints "Failed" in red
		os.Exit(1)
	}
}
