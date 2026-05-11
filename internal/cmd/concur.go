package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashmap-kz/relimpact/internal/diffs"
	"github.com/hashmap-kz/relimpact/internal/gitutils"
	"github.com/hashmap-kz/relimpact/internal/loggr"
)

func CreateAPIReportConcurrently(ctx context.Context, repoDir, oldRef, newRef string, format ReportFormat) string {
	// 1. Concurrent checkout old/new worktrees.
	tmpOld, tmpNew := checkout(ctx, repoDir, oldRef, newRef)
	defer gitutils.CleanupWorktree(repoDir, tmpOld)
	defer gitutils.CleanupWorktree(repoDir, tmpNew)

	// 2. Concurrent API snapshots.
	oldAPI, newAPI := snap(ctx, tmpOld, tmpNew)

	// 3. Render API-only report.
	apiDiff := diffs.DiffAPI(oldAPI, newAPI)
	return renderAPIReport(apiDiff, repoDir, oldRef, newRef, format)
}

//nolint:gocritic
func checkout(ctx context.Context, repoDir, oldRef, newRef string) (string, string) {
	type worktreeResult struct {
		which string
		path  string
		err   error
	}

	worktreeCh := make(chan worktreeResult, 2)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				worktreeCh <- worktreeResult{"old", "", fmt.Errorf("checkout old failed: %v", r)}
			}
		}()
		path := gitutils.CheckoutWorktree(ctx, repoDir, oldRef)
		worktreeCh <- worktreeResult{"old", path, nil}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				worktreeCh <- worktreeResult{"new", "", fmt.Errorf("checkout new failed: %v", r)}
			}
		}()
		path := gitutils.CheckoutWorktree(ctx, repoDir, newRef)
		worktreeCh <- worktreeResult{"new", path, nil}
	}()

	var tmpOld, tmpNew string
	for i := 0; i < 2; i++ {
		res := <-worktreeCh
		if res.err != nil {
			loggr.Fatalf("worktree checkout error: %v", res.err)
		}
		switch res.which {
		case "old":
			tmpOld = res.path
		case "new":
			tmpNew = res.path
		default:
			loggr.Fatalf("unexpected worktree result: %v", res.which)
		}
	}
	return tmpOld, tmpNew
}

//nolint:gocritic
func snap(ctx context.Context, tmpOld, tmpNew string) (map[string]diffs.APIPackage, map[string]diffs.APIPackage) {
	var wgSnapshots sync.WaitGroup
	apiOldCh := make(chan map[string]diffs.APIPackage, 1)
	apiNewCh := make(chan map[string]diffs.APIPackage, 1)

	wgSnapshots.Add(2)
	go func() {
		defer wgSnapshots.Done()
		apiOldCh <- diffs.SnapshotAPI(ctx, tmpOld)
	}()
	go func() {
		defer wgSnapshots.Done()
		apiNewCh <- diffs.SnapshotAPI(ctx, tmpNew)
	}()

	wgSnapshots.Wait()
	close(apiOldCh)
	close(apiNewCh)

	oldAPI := <-apiOldCh
	newAPI := <-apiNewCh

	return oldAPI, newAPI
}
