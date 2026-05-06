package cmd

import (
	"github.com/hashmap-kz/relimpact/internal/diffs"
	"github.com/hashmap-kz/relimpact/internal/gitutils"
)

func CreateChangelogSequential(repoDir, oldRef, newRef string) string {
	return CreateAPIReportSequential(repoDir, oldRef, newRef, ReportFormatMarkdown)
}

func CreateAPIReportSequential(repoDir, oldRef, newRef string, format ReportFormat) string {
	// Checkout old/new worktrees.
	tmpOld := gitutils.CheckoutWorktree(repoDir, oldRef)
	defer gitutils.CleanupWorktree(repoDir, tmpOld)

	tmpNew := gitutils.CheckoutWorktree(repoDir, newRef)
	defer gitutils.CleanupWorktree(repoDir, tmpNew)

	oldAPI := diffs.SnapshotAPI(tmpOld)
	newAPI := diffs.SnapshotAPI(tmpNew)

	apiDiff := diffs.DiffAPI(oldAPI, newAPI)
	return renderAPIReport(apiDiff, repoDir, oldRef, newRef, format)
}
