package cmd

import (
	"context"

	"github.com/hashmap-kz/relimpact/internal/diffs"
	"github.com/hashmap-kz/relimpact/internal/gitutils"
)

func CreateAPIReportSequential(ctx context.Context, repoDir, oldRef, newRef string, format ReportFormat) string {
	// Checkout old/new worktrees.
	tmpOld := gitutils.CheckoutWorktree(ctx, repoDir, oldRef)
	defer gitutils.CleanupWorktree(repoDir, tmpOld)

	tmpNew := gitutils.CheckoutWorktree(ctx, repoDir, newRef)
	defer gitutils.CleanupWorktree(repoDir, tmpNew)

	oldAPI := diffs.SnapshotAPI(ctx, tmpOld)
	newAPI := diffs.SnapshotAPI(ctx, tmpNew)

	apiDiff := diffs.DiffAPI(oldAPI, newAPI)
	return renderAPIReport(apiDiff, repoDir, oldRef, newRef, format)
}
