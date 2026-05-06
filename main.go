package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hashmap-kz/relimpact/internal/x/fmtx"

	"github.com/hashmap-kz/relimpact/cmd"
	"github.com/hashmap-kz/relimpact/internal/loggr"
)

func main() {
	oldRef := flag.String("old", "", "Old git ref")
	newRef := flag.String("new", "", "New git ref")
	greedy := flag.Bool("greedy", false, "Maximum concurrency")
	dir := flag.String("dir", ".", "Git directory")
	format := flag.String("format", "markdown", "Report format: markdown or html")
	outputFile := flag.String("output", "", "Specify output file")
	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		fmtx.Fprintf(os.Stderr, "Usage: relimpact --old <ref> --new <ref> [--format markdown|html]\n")
		os.Exit(1)
	}

	reportFormat := cmd.ReportFormat(strings.ToLower(strings.TrimSpace(*format)))
	switch reportFormat {
	case cmd.ReportFormatMarkdown, cmd.ReportFormatHTML:
		// ok
	default:
		fmtx.Fprintf(os.Stderr, "unsupported format %q: use markdown or html\n", *format)
		os.Exit(1)
	}

	// TODO: log level (envs, CLI)
	loggr.Init(loggr.LevelTrace, "relimpact")

	var report string
	if *greedy {
		report = cmd.CreateAPIReport(*dir, *oldRef, *newRef, reportFormat)
	} else {
		report = cmd.CreateAPIReportSequential(*dir, *oldRef, *newRef, reportFormat)
	}

	if *outputFile != "" {
		err := os.WriteFile(*outputFile, []byte(report), 0o750)
		if err != nil {
			fmtx.Fprintf(os.Stderr, "error writing file: %v", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(report)
	}
}
