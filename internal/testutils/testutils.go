package testutils

import (
	"io"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func RunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	err := cmd.Run()
	require.NoError(t, err, "git command failed: git %v", args)
}

func RunGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	err := cmd.Run()
	require.NoError(t, err, "go command failed: go %v", args)
}
