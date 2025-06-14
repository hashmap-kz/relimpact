package diffs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestDiffSets(t *testing.T) {
	oldList := []string{"A", "B", "C"}
	newList := []string{"B", "C", "D"}

	added, removed := diffSets(oldList, newList)

	assert.ElementsMatch(t, []string{"D"}, added)
	assert.ElementsMatch(t, []string{"A"}, removed)
}

func TestDiffSectionWordCounts(t *testing.T) {
	oldCounts := map[string]int{
		"Intro": 10,
		"Old":   5,
	}
	newCounts := map[string]int{
		"Intro": 12, // changed
		"New":   8,  // added
	}

	lines := diffSectionWordCounts(oldCounts, newCounts)

	assert.Contains(t, lines, "- Section `Intro`: 10 -> 12 words")
	assert.Contains(t, lines, "- Section `Old`: REMOVED (5 words)")
	assert.Contains(t, lines, "- Section `New`: ADDED (8 words)")
}

func TestCountWords(t *testing.T) {
	text := "Hello world! This is a test."
	count := countWords(text)

	assert.Equal(t, 6, count)
}

func TestComputeDocDiff(t *testing.T) {
	oldInfo := &DocInfo{
		Headings:    []string{"Intro"},
		Links:       []string{"https://old.link"},
		Images:      []string{"old.png"},
		SectionWord: map[string]int{"Intro": 10},
	}

	newInfo := &DocInfo{
		Headings:    []string{"Intro", "New Section"},
		Links:       []string{"https://new.link"},
		Images:      []string{},
		SectionWord: map[string]int{"Intro": 12},
	}

	docDiff := computeDocDiff("docs/example.md", oldInfo, newInfo)
	assert.NotNil(t, docDiff)

	assert.ElementsMatch(t, []string{"New Section"}, docDiff.HeadingsAdded)
	assert.ElementsMatch(t, []string{}, docDiff.HeadingsRemoved)

	assert.ElementsMatch(t, []string{"https://new.link"}, docDiff.LinksAdded)
	assert.ElementsMatch(t, []string{"https://old.link"}, docDiff.LinksRemoved)

	assert.ElementsMatch(t, []string{}, docDiff.ImagesAdded)
	assert.ElementsMatch(t, []string{"old.png"}, docDiff.ImagesRemoved)

	assert.Contains(t, docDiff.SectionWordChange, "- Section `Intro`: 10 -> 12 words")
}

func TestFormatAllDocDiffs(t *testing.T) {
	docDiff1 := DocDiff{
		File:          "docs/one.md",
		HeadingsAdded: []string{"Section 1"},
	}

	docDiff2 := DocDiff{
		File:       "docs/two.md",
		LinksAdded: []string{"https://example.com"},
	}

	output := FormatAllDocDiffs([]DocDiff{docDiff1, docDiff2})

	assert.True(t, strings.HasPrefix(output, "\n---\n## Documentation Changes"))
	assert.Contains(t, output, "**`docs/one.md`**")
	assert.Contains(t, output, "**`docs/two.md`**")
	assert.Contains(t, output, "- Section 1")
	assert.Contains(t, output, "- https://example.com")
}

func TestDiffDocs_IntegrationTempDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create old and new dirs
	oldDir := filepath.Join(tmpDir, "old")
	newDir := filepath.Join(tmpDir, "new")
	require.NoError(t, os.Mkdir(oldDir, 0o755))
	require.NoError(t, os.Mkdir(newDir, 0o755))

	// Write old file
	require.NoError(t, os.WriteFile(filepath.Join(oldDir, "test.md"), []byte(`# Intro

This is intro.

[Link](https://old.link)

![Image](old.png)
`), 0o600))

	// Write new file
	require.NoError(t, os.WriteFile(filepath.Join(newDir, "test.md"), []byte(`# Intro

This is intro modified.

# New Section

[Link](https://new.link)
`), 0o600))

	// Run DiffDocs
	diffsResult := DiffDocs(oldDir, newDir)

	assert.Len(t, diffsResult, 1)

	diff := diffsResult[0]
	assert.Equal(t, "test.md", diff.File)
	assert.ElementsMatch(t, []string{"New Section"}, diff.HeadingsAdded)
	assert.ElementsMatch(t, []string{"https://new.link"}, diff.LinksAdded)
	assert.ElementsMatch(t, []string{"https://old.link"}, diff.LinksRemoved)
	assert.ElementsMatch(t, []string{"old.png"}, diff.ImagesRemoved)
	assert.NotEmpty(t, diff.SectionWordChange) // should have Intro 3 -> 4 or similar
}
