package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hashmap-kz/relimpact/cmd"
	"github.com/hashmap-kz/relimpact/internal/loggr"
)

func main() {
	oldRef := flag.String("old", "", "Old git ref")
	newRef := flag.String("new", "", "New git ref")
	greedy := flag.Bool("greedy", false, "Maximum concurrency")
	dir := flag.String("dir", ".", "Git directory")
	format := flag.String("format", "markdown", "Report format: markdown or html")
	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: relimpact --old <ref> --new <ref> [--format markdown|html]\n")
		os.Exit(1)
	}

	reportFormat := cmd.ReportFormat(strings.ToLower(strings.TrimSpace(*format)))
	switch reportFormat {
	case cmd.ReportFormatMarkdown, cmd.ReportFormatHTML:
	default:
		_, _ = fmt.Fprintf(os.Stderr, "unsupported format %q: use markdown or html\n", *format)
		os.Exit(1)
	}

	loggr.Init(loggr.LevelTrace, "relimpact")

	if *greedy {
		fmt.Print(cmd.CreateAPIReport(*dir, *oldRef, *newRef, reportFormat))
	} else {
		fmt.Print(cmd.CreateAPIReportSequential(*dir, *oldRef, *newRef, reportFormat))
	}
}
