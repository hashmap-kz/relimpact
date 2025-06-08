package cmd

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateChangelog_IntegrationTempGit(t *testing.T) {
	tmpDir := t.TempDir()
	t.Log(tmpDir)

	// Init git repo
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "config", "user.email", "test@example.com")

	// Add go.mod
	runGo(t, tmpDir, "mod", "init", "mypkg")

	// Create v1 of package
	pkgDir := filepath.Join(tmpDir, "mypkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "foo.go"), []byte(`package mypkg

func Foo() {}
`), 0o600))

	// Create v1 of docs
	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "intro.md"), []byte(`# Intro

This is v1.
`), 0o600))

	// Create v1 of other file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(`key: value1`), 0o600))

	// Commit v1
	runGit(t, tmpDir, "add", "-A")
	runGit(t, tmpDir, "commit", "-m", "v1")
	runGit(t, tmpDir, "tag", "v1")

	// Modify package -> add new function
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "bar.go"), []byte(`package mypkg

func Bar() {}
`), 0o600))

	// Modify docs
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "intro.md"), []byte(`# Intro

This is v2.

# New Section
`), 0o600))

	// Modify other file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(`key: value2`), 0o600))

	// Commit new state
	runGit(t, tmpDir, "add", "-A")
	runGit(t, tmpDir, "commit", "-m", "add Bar and update docs and config")

	// CreateChangelog
	changelog := CreateChangelog(tmpDir, "v1", "HEAD")

	t.Logf("Changelog:\n%s", changelog)
	assert.Contains(t, changelog, "Bar()")
	assert.Contains(t, changelog, "New Section")
	assert.Contains(t, changelog, "config.yaml")
}

// TODO: testutils
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	err := cmd.Run()
	require.NoError(t, err, "git command failed: git %v", args)
}

// TODO: testutils
func runGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	err := cmd.Run()
	require.NoError(t, err, "go command failed: go %v", args)
}
