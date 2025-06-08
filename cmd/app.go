package cmd

import (
	"strings"

	"github.com/hashmap-kz/relimpact/internal/diffs"
	"github.com/hashmap-kz/relimpact/internal/gitutils"
)

func CreateChangelog(repoDir, oldRef, newRef string) string {
	var sb strings.Builder

	// API
	tmpOld := gitutils.CheckoutWorktree(repoDir, oldRef)
	defer gitutils.CleanupWorktree(repoDir, tmpOld)

	tmpNew := gitutils.CheckoutWorktree(repoDir, newRef)
	defer gitutils.CleanupWorktree(repoDir, tmpNew)

	oldAPI := diffs.SnapshotAPI(tmpOld)
	newAPI := diffs.SnapshotAPI(tmpNew)

	apiDiffResult := diffs.DiffAPI(oldAPI, newAPI)
	sb.WriteString(apiDiffResult.String() + "\n---\n")

	// docs
	docsDiffs := diffs.DiffDocs(tmpOld, tmpNew)
	sb.WriteString(diffs.FormatAllDocDiffs(docsDiffs) + "\n---\n")

	// others
	// TODO: configurable
	includeExts := []string{".sh", ".sql", ".json", ".yaml", ".yml", ".conf", ".ini", ".txt", ".csv"}
	otherSection := diffs.DiffOtherFilesStruct(repoDir, oldRef, newRef, includeExts)
	sb.WriteString(otherSection.String() + "\n---\n")

	return sb.String()
}
