package common

import (
	"os"

	gap "github.com/muesli/go-app-paths"
	"gopkg.in/yaml.v2"
)

type Config struct {
	PruneHistoryAfterDays int `json:"pruneHistoryAfterDays" yaml:"prune-history-after-days,omitempty"`
	HistoryLength         int `json:"historyLength" yaml:"history-length,omitempty"`
}

func (c Config) String() string {
	output, err := yaml.Marshal(c)
	FailOn(err)

	return string(output)
}

func ReadConfig() Config {
	config := Config{
		PruneHistoryAfterDays: 30,
		HistoryLength:         15,
	}

	configPath := getConfigPath()
	if configBytes, err := os.ReadFile(configPath); err == nil {
		err = yaml.Unmarshal(configBytes, &config)
		if err != nil {
			Warn("Invalid config YAML, ignoring...")
		}
	}

	return config
}

func getConfigPath() string {
	scope := gap.NewScope(gap.User, "kv")

	configPath, err := scope.ConfigPath("config.yaml")
	FailOn(err)

	return configPath
}
