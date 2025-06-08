package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hashmap-kz/relimpact/internal/diffs"
	"github.com/hashmap-kz/relimpact/internal/gitutils"
)

func main() {
	oldRef := flag.String("old", "", "Old git ref")
	newRef := flag.String("new", "", "New git ref")
	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		log.Fatal("Usage: apidiff --old <ref> --new <ref>")
	}

	// API
	tmpOld := gitutils.CheckoutWorktree(*oldRef)
	defer gitutils.CleanupWorktree(tmpOld)

	tmpNew := gitutils.CheckoutWorktree(*newRef)
	defer gitutils.CleanupWorktree(tmpNew)

	oldAPI := diffs.SnapshotAPI(tmpOld)
	newAPI := diffs.SnapshotAPI(tmpNew)

	apiDiffResult := diffs.DiffAPI(oldAPI, newAPI)
	fmt.Println(apiDiffResult.String())

	// docs
	docsDiffs := diffs.DiffDocs(tmpOld, tmpNew)
	fmt.Println(diffs.FormatAllDocDiffs(docsDiffs))

	// others
	// TODO: configurable
	includeExts := []string{".sh", ".sql", ".json", ".yaml", ".yml", ".conf", ".ini", ".txt", ".csv"}
	otherSection := diffs.DiffOtherFilesStruct(*oldRef, *newRef, includeExts)
	fmt.Println(otherSection.String())
}
