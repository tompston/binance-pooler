package settings

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Env is a struct that holds the keys for the environment variables.
type Env struct {
	DefaultConfigPath string
	ConfigPathKey     string
	TestModeKey       string
	UseTestDbKey      string
	IsProductionKey   string
}

// Return true if the IsProductionKey environment variable is set to "true".
func (e *Env) IsProduction() bool {
	envVar := GetEnvVar(e.IsProductionKey)
	return envVar == "true"
}

// GetConfigPath returns the path to the config file.
func (e *Env) GetConfigPath(path ...string) string {
	if len(path) == 1 {
		return path[0]
	}

	envPath, ok := os.LookupEnv(e.ConfigPathKey)
	if ok && envPath != "" {
		return envPath
	}

	return e.DefaultConfigPath
}

// ShouldUseTestDb returns true if the UseTestDbKey environment variable is set to true.
func (e *Env) ShouldUseTestDb() bool {
	return os.Getenv(e.UseTestDbKey) == "true"
}

// GetEnvVar returns the value of the environment variable if it exists and
// is not an empty string. If the env var does not exist, the func
// returns an empty string.
func GetEnvVar(key string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}

	return ""
}

// Structure for the data that is stored in the toml config file
type TomlConfig struct {
	Api struct {
		Host string `toml:"host"`
		Port int    `toml:"port"`
	} `toml:"api"`
	MongoUri string `toml:"mongo_uri"`
}

// NewConfig loads a toml config file with the specified path.
func NewConfig(path string) (*TomlConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("err while reading toml config file: %v, %v", err, path)
	}

	var conf TomlConfig
	err = toml.Unmarshal(file, &conf)
	return &conf, err
}
