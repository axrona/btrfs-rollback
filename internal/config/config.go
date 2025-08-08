// Package config provides functionality for managing and parsing the configuration file
// for the btrfs-rollback tool. It defines the configuration structure and offers functions
// to load and parse the configuration settings from a TOML file.
//
// The configuration file is expected to be located at "/etc/btrfs-rollback.toml" and
// contains settings for:
//   - SubvolMain: The name of the main BTRFS subvolume, typically mounted as the Linux root.
//   - SubvolSnapshots: The name of the BTRFS subvolume containing snapshots.
//   - Mountpoint: The directory where the BTRFS root is mounted.
//
// If the configuration file cannot be found or parsed, default values are used to ensure
// the tool can function without custom settings.
package config

import (
	"github.com/BurntSushi/toml"
)

// Config represents the configuration structure for btrfs-rollback,
// containing essential settings such as the main subvolume, snapshot subvolume, and mount point.
type Config struct {
	SubvolMain      string  `toml:"subvol_main"`
	SubvolSnapshots string  `toml:"subvol_snapshots"`
	Mountpoint      string  `toml:"mountpoint"`
	Dev             *string `toml:"dev"`
}

// DefaultConfig holds the default values used by the btrfs-rollback tool
// when no custom configuration file is available.
var DefaultConfig = Config{
	SubvolMain:      "@",
	SubvolSnapshots: "@snapshots",
	Mountpoint:      "/btrfs",
}

// ParseConfig reads the configuration file located at configPath,
// parses its contents, and returns the corresponding Config struct.
// If the file cannot be read or parsed, DefaultConfig is returned.
func ParseConfig(path string) Config {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return DefaultConfig
	}

	return config
}
