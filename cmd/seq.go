package cmd

import (
	"strings"

	"github.com/hashmap-kz/relimpact/internal/diffs"
	"github.com/hashmap-kz/relimpact/internal/gitutils"
)

func CreateChangelogSequential(repoDir, oldRef, newRef string) string {
	// Checkout old/new worktrees
	tmpOld := gitutils.CheckoutWorktree(repoDir, oldRef)
	defer gitutils.CleanupWorktree(repoDir, tmpOld)

	tmpNew := gitutils.CheckoutWorktree(repoDir, newRef)
	defer gitutils.CleanupWorktree(repoDir, tmpNew)

	// Snapshot API
	oldAPI := diffs.SnapshotAPI(tmpOld)
	newAPI := diffs.SnapshotAPI(tmpNew)

	// Run diffs
	var sb strings.Builder

	// API diff
	apiDiffResult := diffs.DiffAPI(oldAPI, newAPI)
	sb.WriteString(apiDiffResult.String())
	sb.WriteString("\n")

	// Docs diff
	docsDiffs := diffs.DiffDocs(tmpOld, tmpNew)
	sb.WriteString(diffs.FormatAllDocDiffs(docsDiffs))
	sb.WriteString("\n")

	// go.mod diff
	modDiffs := diffs.DiffGoMod(tmpOld, tmpNew)
	sb.WriteString(modDiffs.String())
	sb.WriteString("\n")

	// Other files diff
	otherSection := diffs.DiffOther(repoDir, oldRef, newRef, includeExts)
	sb.WriteString(otherSection.String())
	sb.WriteString("\n")

	return sb.String()
}
