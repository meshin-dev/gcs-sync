package cmd

import (
	"fmt"
	"gcs_sync/internal/config"
	"gcs_sync/internal/logging"
	"gcs_sync/internal/watcher"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	cfgPath  string
	logLevel string
	rootCmd  = &cobra.Command{
		Use:   "gcs-sync",
		Short: "Bi-directional Google Cloud Storage synchronizer",
		RunE:  run,
	}
)

// init initializes the command-line flags for the root command.
// It sets up two persistent flags:
//   - config: Specifies the path to the YAML configuration file.
//   - log-level: Sets the logging level for the application.
func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c",
		"/app/settings/config.yaml", "path to YAML configuration")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l",
		"info", "log level (trace|debug|info|warn|error)")
}

// run is the main execution function for the gcs-sync command.
// It initializes the logger, loads the configuration, sets up the application
// using the fx dependency injection framework, and runs the application.
//
// Parameters:
//   - _ *cobra.Command: Unused parameter representing the Cobra command.
//   - _ []string: Unused parameter representing command-line arguments.
//
// Returns:
//   - error: An error if any step in the process fails, nil otherwise.
func run(_ *cobra.Command, _ []string) error {
	// Configure global logger
	logging.Init(logLevel)

	// Load config early so startup fails fast if YAML is invalid
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Build Fx app
	app := fx.New(
		fx.Supply(cfg),
		fx.Supply(logging.L()),
		fx.Invoke(watcher.StartAll),
	)

	// Blocks until SIGINT / SIGTERM
	app.Run()

	return nil
}

// Execute lets main.go launch the CLI.
func Execute() error { return rootCmd.Execute() }
