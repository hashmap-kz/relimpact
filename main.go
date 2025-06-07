package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hashmap-kz/relimpact/internal/diffs"
	"github.com/hashmap-kz/relimpact/internal/git_utils"
)

func main() {
	oldRef := flag.String("old", "", "Old git ref")
	newRef := flag.String("new", "", "New git ref")
	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		log.Fatal("Usage: apidiff --old <ref> --new <ref>")
	}

	// API

	tmpOld := git_utils.CheckoutWorktree(*oldRef)
	defer git_utils.CleanupWorktree(tmpOld)

	tmpNew := git_utils.CheckoutWorktree(*newRef)
	defer git_utils.CleanupWorktree(tmpNew)

	oldAPI := diffs.SnapshotAPI(tmpOld)
	newAPI := diffs.SnapshotAPI(tmpNew)

	diffs.DiffAPI(oldAPI, newAPI)

	// docs

	docsDiffs := diffs.DiffDocs(tmpOld, tmpNew)
	if len(docsDiffs) > 0 {
		for _, section := range docsDiffs {
			fmt.Println(section)
		}
	}

	// others

	// TODO: configurable
	includeExts := []string{".sh", ".sql", ".json", ".yaml", ".yml", ".conf", ".ini", ".txt", ".csv"}
	otherSection := diffs.DiffOtherFiles(*oldRef, *newRef, includeExts)
	if otherSection != "" {
		fmt.Println(otherSection)
	}
}
