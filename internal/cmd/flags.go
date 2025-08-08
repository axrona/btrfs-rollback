// Package cmd defines the root command and global flags for the btrfs-rollback CLI tool.
// It sets up persistent flags such as --config and --dry-run,
// and loads configuration automatically when the command is invoked.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xeyossr/btrfs-rollback/internal/config"
)

var (
	cfg     config.Config
	cfgPath string
	dryRun  bool

	// RootCmd is the main entry point for the CLI
	RootCmd = &cobra.Command{
		Use:   "btrfs-rollback",
		Short: "",
		Long:  ``,
	}
)

// init initializes the CLI by setting up global flags and loading the config
func init() {
	// Setup persistent flags
	RootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "/etc/btrfs-rollback.toml", "Path to config file")
	RootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Simulate the operations without making changes")
}

// loadConfig reads the config file from the given path or falls back to default
func loadConfig() config.Config {
	if cfgPath != "" {
		absPath, err := filepath.Abs(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid config path: %v\n", err)
			os.Exit(1)
		}

		return config.ParseConfig(absPath)
	}

	return config.Config{}
}

// Config returns the loaded configuration
func Config() config.Config {
	return loadConfig()
}

// DryRun returns whether --dry-run flag was set
func DryRun() bool {
	return dryRun
}
