package cmd

import (
	"time"

	"github.com/hashmap-kz/relimpact/internal/diffs"
)

type ReportFormat string

const (
	ReportFormatMarkdown ReportFormat = "markdown"
	ReportFormatHTML     ReportFormat = "html"
)

func renderAPIReport(apiDiff *diffs.APIDiff, repoDir, oldRef, newRef string, format ReportFormat) string {
	meta := diffs.ReportMetadata{
		Repo:   repoDir,
		OldRef: oldRef,
		NewRef: newRef,
		Now:    time.Now(),
	}
	if format == ReportFormatHTML {
		return apiDiff.HTML(meta)
	}
	return apiDiff.Markdown(meta)
}
