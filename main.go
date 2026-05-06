package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hashmap-kz/relimpact/cmd"
	"github.com/hashmap-kz/relimpact/internal/loggr"
	"github.com/hashmap-kz/relimpact/internal/x/fmtx"
)

func main() {
	oldRef := flag.String("old", "", "old git ref, tag, branch, or commit")
	newRef := flag.String("new", "", "new git ref, tag, branch, or commit")
	dir := flag.String("dir", ".", "path to git repository")
	format := flag.String("format", "markdown", "report format: markdown or html")
	outputFile := flag.String("output", "", "write report to file instead of stdout")
	greedy := flag.Bool("greedy", false, "use maximum concurrency")

	flag.Usage = usage

	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		usage()
		os.Exit(2)
	}

	reportFormat := cmd.ReportFormat(strings.ToLower(strings.TrimSpace(*format)))
	switch reportFormat {
	case cmd.ReportFormatMarkdown, cmd.ReportFormatHTML:
		// ok
	default:
		fmtx.Fprintf(os.Stderr, "relimpact: unsupported format %q: use markdown or html\n\n", *format)
		usage()
		os.Exit(2)
	}

	// TODO: make log level configurable via env/CLI.
	loggr.Init(loggr.LevelTrace, "relimpact")

	var report string
	if *greedy {
		report = cmd.CreateAPIReport(*dir, *oldRef, *newRef, reportFormat)
	} else {
		report = cmd.CreateAPIReportSequential(*dir, *oldRef, *newRef, reportFormat)
	}

	if *outputFile == "" {
		fmt.Print(report)
		return
	}

	if err := os.WriteFile(*outputFile, []byte(report), 0o644); err != nil {
		fmtx.Fprintf(os.Stderr, "relimpact: write %s: %v\n", *outputFile, err)
		os.Exit(1)
	}
}

func usage() {
	fmtx.Fprintf(os.Stderr, `relimpact compares exported Go API between two git refs.

Usage:
  relimpact --old <ref> --new <ref> [flags]

Required:
  --old string       old git ref, tag, branch, or commit
  --new string       new git ref, tag, branch, or commit

Flags:
  --dir string       path to git repository (default ".")
  --format string    report format: markdown or html (default "markdown")
  --output string    write report to file instead of stdout
  --greedy           use maximum concurrency
  -h, --help         show help

Examples:
  relimpact --old v1.0.0 --new HEAD

  relimpact --old v1.0.0 --new HEAD \
    --format html \
    --output api-report.html

  relimpact --dir /path/to/repo \
    --old main \
    --new feature/api-change \
    --output api-report.md

`)
}
