package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

var defaultConfig = Config{
	DebounceDelay: 500 * time.Millisecond,
	GoModFile:     "go.mod",
}

type Config struct {
	LogLevel      slog.Level      `yaml:"logLevel"`
	DebounceDelay time.Duration   `yaml:"debounceDelay"`
	Build         [][]string      `yaml:"build"`
	Services      []ServiceConfig `yaml:"services"`
	Watch         WatchConfig     `yaml:"watch"`
	GoModFile     string          `yaml:"goModFile"`
}

func Parse(path string) (*Config, error) {
	slog.Info("loading config", "path", path)

	configYaml, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := defaultConfig
	err = yaml.Unmarshal(configYaml, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

type ServiceConfig struct {
	Command []string `yaml:"command"`
}

type WatchConfig struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}
