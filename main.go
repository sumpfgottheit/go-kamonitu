package main

// Helper function to parse validation rules from a tag

import (
	"fmt"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"os"
	"time"
)

func setupLogging(loglevel slog.Level, logDir string) {
	if logDir == "" {
		w := os.Stderr
		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				AddSource:  true,
				Level:      loglevel,
				TimeFormat: time.RFC3339,
			}),
		))
	} else {
		w := &lumberjack.Logger{
			Filename:   logDir + "kamonitu.log",
			MaxSize:    10, // megabytes
			MaxBackups: 5,
			MaxAge:     30,   // days
			Compress:   true, // disabled by default
		}
		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				AddSource:  false,
				Level:      loglevel,
				TimeFormat: time.RFC3339,
				NoColor:    true,
			}),
		))
	}
	slog.Info("Initialized Logging.", "Level", loglevel.String())

}

func main() {
	var debug bool
	var configFile string
	var appConfig *AppConfig

	rootCmd := &cobra.Command{
		Use:   "kamonitu",
		Short: "Kamonitu is a configuration validation tool.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
