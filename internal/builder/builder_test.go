package builder

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	t.Run("runs two successive steps", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "test1")
		testFile2 := path.Join(t.TempDir(), "test2")
		steps := [][]string{
			{"touch", testFile1},
			{"touch", testFile2},
		}

		err := Build(steps)
		assert.NoError(t, err)

		assert.FileExists(t, testFile1)
		assert.FileExists(t, testFile2)
	})
	t.Run("stops execution if one step fails", func(t *testing.T) {
		testFile1 := path.Join(t.TempDir(), "test1")
		testFile2 := path.Join(t.TempDir(), "test2")
		steps := [][]string{
			{"touch", testFile1},
			{"false"},
			{"touch", testFile2},
		}

		err := Build(steps)
		assert.ErrorContains(t, err, "failed to execute build step 1: exit status 1")

		assert.FileExists(t, testFile1)
		assert.NoFileExists(t, testFile2)
	})
}
