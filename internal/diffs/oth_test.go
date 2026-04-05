package diffs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashmap-kz/relimpact/internal/testutils"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestOtherFilesDiffSummary_String(t *testing.T) {
	summary := &OtherFilesDiffSummary{
		Diffs: []OtherFileDiff{
			{
				Ext:      ".sh",
				Added:    []string{"scripts/setup.sh"},
				Modified: []string{"scripts/deploy.sh"},
				Removed:  []string{},
				Other:    []string{"scripts/misc.sh"},
			},
			{
				Ext:      ".sql",
				Added:    []string{"migrations/001_init.sql"},
				Modified: []string{},
				Removed:  []string{"migrations/000_old.sql"},
				Other:    []string{},
			},
		},
	}

	out := summary.String()

	readFile := testutils.ReadTestData(t, t.Name()+".md")
	require.Equal(t, out, string(readFile))
}

func TestDiffOtherFilesStruct_IntegrationTempGit(t *testing.T) {
	tmpDir := t.TempDir()

	// Init git repo
	testutils.RunGit(t, tmpDir, "init")
	testutils.RunGit(t, tmpDir, "config", "user.name", "Test User")
	testutils.RunGit(t, tmpDir, "config", "user.email", "test@example.com")

	// Write initial file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "script.sh"), []byte("echo hello\n"), 0o600))
	testutils.RunGit(t, tmpDir, "add", "script.sh")
	testutils.RunGit(t, tmpDir, "commit", "-m", "initial commit")
	// After first commit
	testutils.RunGit(t, tmpDir, "tag", "oldref")
	oldRef := "oldref"

	// Modify file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "script.sh"), []byte("echo hello world\n"), 0o600))
	// Add new file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(`{"key": "value"}`), 0o600))
	testutils.RunGit(t, tmpDir, "add", "-A")
	testutils.RunGit(t, tmpDir, "commit", "-m", "update files")
	newRef := "HEAD"

	// Run DiffOther
	summary := DiffOther(tmpDir, oldRef, newRef, []string{".sh", ".json"})

	assert.NotEmpty(t, summary.Diffs)

	foundSh := false
	foundJSON := false

	for _, d := range summary.Diffs {
		if d.Ext == ".sh" {
			foundSh = true
			assert.ElementsMatch(t, []string{"script.sh"}, d.Modified)
		}
		if d.Ext == ".json" {
			foundJSON = true
			assert.ElementsMatch(t, []string{"config.json"}, d.Added)
		}
	}

	assert.True(t, foundSh)
	assert.True(t, foundJSON)
}
