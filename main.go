package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hashmap-kz/relimpact/cmd"
)

func main() {
	oldRef := flag.String("old", "", "Old git ref")
	newRef := flag.String("new", "", "New git ref")
	greedy := flag.Bool("greedy", false, "Maximum concurrency")
	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		log.Fatal("Usage: relimpact --old <ref> --new <ref>")
	}

	if *greedy {
		fmt.Println(cmd.CreateChangelog(".", *oldRef, *newRef))
	} else {
		fmt.Println(cmd.CreateChangelogSequential(".", *oldRef, *newRef))
	}
}
