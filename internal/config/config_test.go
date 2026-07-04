package config

import (
	"log/slog"
	"os"
	"path"
	"testing"
	"time"

	"github.com/st3iny/nitrogen/test/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("parses valid config", func(t *testing.T) {
		configPath := path.Join(t.TempDir(), "config.yaml")
		err := os.WriteFile(configPath, testdata.ConfigYaml, 0644)
		require.NoError(t, err)

		expectedConfig := &Config{
			LogLevel:      slog.LevelDebug,
			DebounceDelay: 500 * time.Millisecond,
			Build: [][]string{
				{"make", "-f", "./test/Makefile", "build"},
			},
			Services: []ServiceConfig{
				{[]string{"./bin/svc1"}},
				{[]string{"./bin/svc2"}},
			},
			Watch: WatchConfig{
				Exclude: []string{
					".git*",
					".idea",
				},
			},
			GoModFile: "custom-go.mod",
		}

		cfg, err := Parse(configPath)
		require.NoError(t, err)
		assert.Equal(t, expectedConfig, cfg)
	})
	t.Run("applies default config", func(t *testing.T) {
		configPath := path.Join(t.TempDir(), "config.yaml")
		err := os.WriteFile(configPath, []byte("{}"), 0644)
		require.NoError(t, err)

		expectedConfig := &Config{
			DebounceDelay: 500 * time.Millisecond,
			GoModFile:     "go.mod",
		}

		cfg, err := Parse(configPath)
		require.NoError(t, err)
		assert.Equal(t, expectedConfig, cfg)
	})
	t.Run("fails if file does not exist", func(t *testing.T) {
		cfg, err := Parse("/does/not/exist")
		require.ErrorContains(t, err, "failed to read config file:")
		assert.Nil(t, cfg)
	})
	t.Run("fails if file contains invalid yaml", func(t *testing.T) {
		configPath := path.Join(t.TempDir(), "config.yaml")
		err := os.WriteFile(configPath, []byte("not yaml"), 0644)
		require.NoError(t, err)

		cfg, err := Parse(configPath)
		require.ErrorContains(t, err, "failed to parse config file:")
		assert.Nil(t, cfg)
	})
}
