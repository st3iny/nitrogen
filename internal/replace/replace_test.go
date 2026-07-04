package replace

import (
	"os"
	"path"
	"testing"

	"github.com/st3iny/nitrogen/test/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("parses mod file", func(t *testing.T) {
		modFile := path.Join(t.TempDir(), "go.mod")
		err := os.WriteFile(modFile, []byte(testdata.GoModWithReplaces), 0644)
		require.NoError(t, err)

		replaces, err := Parse(modFile)
		require.NoError(t, err)
		require.Len(t, replaces, 2)
		assert.Equal(t, "my-dep1", replaces[0].Package)
		assert.Equal(t, "../../my-dep1", replaces[0].LocalPath)
		assert.Equal(t, "my-dep2", replaces[1].Package)
		assert.Equal(t, "../../my-dep2", replaces[1].LocalPath)
	})
	t.Run("parses mod file without replaces", func(t *testing.T) {
		modFile := path.Join(t.TempDir(), "go.mod")
		err := os.WriteFile(modFile, []byte(testdata.GoMod), 0644)
		require.NoError(t, err)

		replaces, err := Parse(modFile)
		require.NoError(t, err)
		assert.Len(t, replaces, 0)
	})
	t.Run("fails if file does not exist", func(t *testing.T) {
		replaces, err := Parse("/does/not/exist")
		require.ErrorContains(t, err, "failed to read mod file:")
		assert.Nil(t, replaces)
	})
	t.Run("fails if mod file is invalid", func(t *testing.T) {
		modFile := path.Join(t.TempDir(), "go.mod")
		err := os.WriteFile(modFile, []byte("not a go mod file"), 0644)
		require.NoError(t, err)

		replaces, err := Parse(modFile)
		require.ErrorContains(t, err, "failed to parse mod file:")
		assert.Nil(t, replaces)
	})
}
