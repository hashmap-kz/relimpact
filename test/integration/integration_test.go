//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func relimpactBin() string {
	// TODO: also with envs
	return "./bin/relimpact"
}

func relimpactRepoURL() string {
	return "https://github.com/hashmap-kz/relimpact.git"
}

func TestExpected_Result1(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := runCmd("git", "clone", relimpactRepoURL(), tmpDir)
	require.NoError(t, err)

	md, err := runCmd(relimpactBin(), "--old", "269945d", "--new", "5a0c267", "--greedy", "--dir", tmpDir)
	require.NoError(t, err)

	readFile, err := os.ReadFile(filepath.Join("testdata", t.Name()+".md"))
	require.NoError(t, err)

	expected := string(readFile)
	require.Equal(t, expected, md)
}
