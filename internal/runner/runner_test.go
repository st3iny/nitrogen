package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/st3iny/nitrogen/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_RunServices(t *testing.T) {
	t.Run("runs services and exits on cancel", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "file1")
		testFile2 := path.Join(t.TempDir(), "file2")

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		runner := Runner{
			Services: []config.ServiceConfig{
				{[]string{"touch", testFile1}},
				{[]string{"touch", testFile2}},
			},
		}

		var wg sync.WaitGroup
		wg.Go(func() { runner.RunServices(ctx) })

		time.Sleep(1 * time.Second)
		cancel()
		wg.Wait()

		assert.FileExists(t, testFile1)
		assert.FileExists(t, testFile2)
	})
	t.Run("restarts services on exit", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "file1")
		testFile2 := path.Join(t.TempDir(), "file2")

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		runner := Runner{
			Services: []config.ServiceConfig{
				{[]string{"sh", "-c", fmt.Sprintf("echo line >> %s; exit 0", testFile1)}},
				{[]string{"sh", "-c", fmt.Sprintf("echo line >> %s; exit 1", testFile2)}},
			},
			RestartDelay: 600 * time.Millisecond,
		}

		var wg sync.WaitGroup
		wg.Go(func() { runner.RunServices(ctx) })

		time.Sleep(1 * time.Second)
		cancel()
		wg.Wait()

		assert.FileExists(t, testFile1)
		assert.FileExists(t, testFile2)

		file1Data, err := os.ReadFile(testFile1)
		require.NoError(t, err)

		file2Data, err := os.ReadFile(testFile2)
		require.NoError(t, err)

		assert.Greater(t, bytes.Count(file1Data, []byte("line\n")), 1)
		assert.Greater(t, bytes.Count(file2Data, []byte("line\n")), 1)
	})
}
