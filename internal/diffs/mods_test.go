package diffs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeGoMod(t *testing.T, dir, content string) {
	t.Helper()
	path := filepath.Join(dir, "go.mod")
	err := os.WriteFile(path, []byte(content), 0o600)
	assert.NoError(t, err, "failed to write go.mod")
}

func TestDiffGoMod(t *testing.T) {
	tmpDirOld := t.TempDir()
	tmpDirNew := t.TempDir()

	oldGoMod := `
module example.com/mymodule

go 1.21

require (
    github.com/foo/bar v1.2.3
    github.com/old/pkg v0.9.1
)

replace github.com/foo/bar v1.2.3 => ../local/bar
`

	newGoMod := `
module example.com/mymodule

go 1.21

require (
    github.com/foo/bar v1.3.0
    github.com/new/dependency v1.0.0
)

replace github.com/foo/bar v1.3.0 => ../local/bar-new
`

	writeGoMod(t, tmpDirOld, oldGoMod)
	writeGoMod(t, tmpDirNew, newGoMod)

	diff := DiffGoMod(tmpDirOld, tmpDirNew)

	assert.ElementsMatch(t, diff.DependenciesAdded, []string{
		"github.com/new/dependency v1.0.0",
	})

	assert.ElementsMatch(t, diff.DependenciesRemoved, []string{
		"github.com/old/pkg v0.9.1",
	})

	assert.ElementsMatch(t, diff.DependenciesUpdated, []string{
		"github.com/foo/bar v1.2.3 -> v1.3.0",
	})
}
