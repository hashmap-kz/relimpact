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
	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		log.Fatal("Usage: relimpact --old <ref> --new <ref>")
	}

	fmt.Println(cmd.CreateChangelog(".", *oldRef, *newRef))
}
