package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hashmap-kz/relimpact/internal/diffs"
	"github.com/hashmap-kz/relimpact/internal/gitutils"
)

// TODO: configurable
var includeExts = []string{".sh", ".sql", ".json", ".yaml", ".yml", ".conf", ".ini", ".txt", ".csv"}

func CreateChangelog(repoDir, oldRef, newRef string) string {
	//  1. Concurrent checkout old/new worktrees
	tmpOld, tmpNew := checkout(repoDir, oldRef, newRef)
	defer gitutils.CleanupWorktree(repoDir, tmpOld)
	defer gitutils.CleanupWorktree(repoDir, tmpNew)

	//  2. Concurrent SnapshotAPI old/new
	oldAPI, newAPI := snap(tmpOld, tmpNew)

	//  3. Concurrent make diffs
	return runDiffs(repoDir, oldRef, newRef, oldAPI, newAPI, tmpOld, tmpNew)
}

//nolint:gocritic
func checkout(repoDir, oldRef, newRef string) (string, string) {
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
		path := gitutils.CheckoutWorktree(repoDir, oldRef)
		worktreeCh <- worktreeResult{"old", path, nil}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				worktreeCh <- worktreeResult{"new", "", fmt.Errorf("checkout new failed: %v", r)}
			}
		}()
		path := gitutils.CheckoutWorktree(repoDir, newRef)
		worktreeCh <- worktreeResult{"new", path, nil}
	}()

	var tmpOld, tmpNew string
	for i := 0; i < 2; i++ {
		res := <-worktreeCh
		if res.err != nil {
			panic(res.err)
		}
		switch res.which {
		case "old":
			tmpOld = res.path
		case "new":
			tmpNew = res.path
		default:
			panic(fmt.Sprintf("unexpected worktree result: %v", res.which))
		}
	}
	return tmpOld, tmpNew
}

//nolint:gocritic
func snap(tmpOld, tmpNew string) (map[string]diffs.APIPackage, map[string]diffs.APIPackage) {
	var wgSnapshots sync.WaitGroup
	apiOldCh := make(chan map[string]diffs.APIPackage, 1)
	apiNewCh := make(chan map[string]diffs.APIPackage, 1)

	wgSnapshots.Add(2)
	go func() {
		defer wgSnapshots.Done()
		apiOldCh <- diffs.SnapshotAPI(tmpOld)
	}()
	go func() {
		defer wgSnapshots.Done()
		apiNewCh <- diffs.SnapshotAPI(tmpNew)
	}()

	wgSnapshots.Wait()
	close(apiOldCh)
	close(apiNewCh)

	oldAPI := <-apiOldCh
	newAPI := <-apiNewCh

	return oldAPI, newAPI
}

func runDiffs(
	repoDir, oldRef, newRef string,
	oldAPI, newAPI map[string]diffs.APIPackage,
	tmpOld, tmpNew string,
) string {
	var wgDiffs sync.WaitGroup
	apiDiffCh := make(chan string, 1)
	docsDiffCh := make(chan string, 1)
	otherDiffCh := make(chan string, 1)

	wgDiffs.Add(3)

	// API diff
	go func() {
		defer wgDiffs.Done()
		apiDiffResult := diffs.DiffAPI(oldAPI, newAPI)
		apiDiffCh <- apiDiffResult.String() + "\n"
	}()

	// Docs diff
	go func() {
		defer wgDiffs.Done()
		docsDiffs := diffs.DiffDocs(tmpOld, tmpNew)
		docsDiffCh <- diffs.FormatAllDocDiffs(docsDiffs) + "\n"
	}()

	// Other files diff
	go func() {
		defer wgDiffs.Done()
		otherSection := diffs.DiffOther(repoDir, oldRef, newRef, includeExts)
		otherDiffCh <- otherSection.String() + "\n"
	}()

	// Wait for all diffs to complete
	wgDiffs.Wait()
	close(apiDiffCh)
	close(docsDiffCh)
	close(otherDiffCh)

	//  Collect results
	var sb strings.Builder
	sb.WriteString(<-apiDiffCh)
	sb.WriteString(<-docsDiffCh)
	sb.WriteString(<-otherDiffCh)

	return sb.String()
}
