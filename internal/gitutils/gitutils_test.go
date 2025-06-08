package gitutils

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckoutAndCleanupWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Init git repo
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "config", "user.email", "test@example.com")

	// Write a file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("hello world"), 0o644))
	runGit(t, tmpDir, "add", "file.txt")
	runGit(t, tmpDir, "commit", "-m", "initial commit")
	runGit(t, tmpDir, "tag", "v1")

	// Checkout worktree at v1
	oldCwd, _ := os.Getwd()

	// restore working dir
	defer func(dir string) {
		err := os.Chdir(dir)
		if err != nil {
			t.Log("cannot restore chdir")
		}
	}(oldCwd)

	// git worktree must be run inside repo!
	require.NoError(t, os.Chdir(tmpDir))

	worktreeDir := CheckoutWorktree(tmpDir, "v1")
	t.Logf("Worktree dir: %s", worktreeDir)

	// Verify worktree dir exists and contains file.txt
	_, err := os.Stat(filepath.Join(worktreeDir, "file.txt"))
	require.NoError(t, err, "file.txt should exist in worktree")

	// Cleanup worktree
	CleanupWorktree(tmpDir, worktreeDir)

	// Verify worktree dir is gone
	_, err = os.Stat(worktreeDir)
	require.Error(t, err, "worktree dir should be removed")
	require.True(t, os.IsNotExist(err), "worktree dir should be removed")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	require.NoError(t, err, "git command failed: git %v", args)
}
