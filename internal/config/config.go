package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

type Config struct {
	LogLevel    slog.Level      `yaml:"logLevel"`
	ReloadDelay time.Duration   `yaml:"reloadDelay"`
	Build       [][]string      `yaml:"build"`
	Services    []ServiceConfig `yaml:"services"`
	Watch       WatchConfig     `yaml:"watch"`
}

func Parse(path string) (*Config, error) {
	slog.Info("loading config", "path", path)

	configYaml, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(configYaml, &cfg)
	return &cfg, err
}

type ServiceConfig struct {
	Command []string `yaml:"command"`
}

type WatchConfig struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}
