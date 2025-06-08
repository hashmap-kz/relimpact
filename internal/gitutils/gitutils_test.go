package gitutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashmap-kz/relimpact/internal/testutils"

	"github.com/stretchr/testify/require"
)

func TestCheckoutAndCleanupWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Init git repo
	testutils.RunGit(t, tmpDir, "init")
	testutils.RunGit(t, tmpDir, "config", "user.name", "Test User")
	testutils.RunGit(t, tmpDir, "config", "user.email", "test@example.com")

	// Write a file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("hello world"), 0o600))
	testutils.RunGit(t, tmpDir, "add", "file.txt")
	testutils.RunGit(t, tmpDir, "commit", "-m", "initial commit")
	testutils.RunGit(t, tmpDir, "tag", "v1")

	// Checkout worktree at v1
	oldCwd, err := os.Getwd()
	require.NoError(t, err)

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

	// Verify worktree dir exists and contains file.txt
	_, err = os.Stat(filepath.Join(worktreeDir, "file.txt"))
	require.NoError(t, err, "file.txt should exist in worktree")

	// Cleanup worktree
	CleanupWorktree(tmpDir, worktreeDir)

	// Verify worktree dir is gone
	_, err = os.Stat(worktreeDir)
	require.Error(t, err, "worktree dir should be removed")
	require.True(t, os.IsNotExist(err), "worktree dir should be removed")
}
