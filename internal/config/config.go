package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

// Config mirrors the YAML schema.
type Config struct {
	Sync []SyncRule `yaml:"sync"`
}

type SyncDirection string

func (d SyncDirection) String() string {
	return string(d)
}

const (
	Full          SyncDirection = "full"
	LocalToRemote SyncDirection = "local_to_remote"
	RemoteToLocal SyncDirection = "remote_to_local"
)

type SyncRule struct {
	Src              string          `yaml:"src"`
	Dst              string          `yaml:"dst"`
	Directions       []SyncDirection `yaml:"directions"`
	Ignore           []string        `yaml:"ignore"`
	Enabled          bool            `yaml:"enabled"`
	DebounceWindow   time.Duration   `yaml:"debounce_window"`
	RemotePollWindow time.Duration   `yaml:"remote_poll_window"`
}

// Load parses a YAML configuration file and returns a Config struct.
//
// It reads the file from the specified path, unmarshals the YAML content
// into a Config struct, and returns a pointer to the resulting Config.
//
// Parameters:
//   - path: A string representing the file path of the YAML configuration file to be loaded.
//
// Returns:
//   - *Config: A pointer to the parsed Config struct containing the configuration data.
//   - error: An error if any occurred during file reading or YAML unmarshaling. It returns nil if successful.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
