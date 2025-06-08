package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashmap-kz/relimpact/cmd"
)

func main() {
	oldRef := flag.String("old", "", "Old git ref")
	newRef := flag.String("new", "", "New git ref")
	greedy := flag.Bool("greedy", false, "Maximum concurrency")
	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: relimpact --old <ref> --new <ref>")
		os.Exit(1)
	}

	if *greedy {
		fmt.Println(cmd.CreateChangelog(".", *oldRef, *newRef))
	} else {
		fmt.Println(cmd.CreateChangelogSequential(".", *oldRef, *newRef))
	}
}
