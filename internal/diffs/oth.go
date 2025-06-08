package diffs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type OtherFileDiff struct {
	Ext      string
	Added    []string
	Modified []string
	Removed  []string
	Other    []string
}

type OtherFilesDiffSummary struct {
	Diffs []OtherFileDiff
}

func (s *OtherFilesDiffSummary) String() string {
	if len(s.Diffs) == 0 {
		return ""
	}

	var b bytes.Buffer
	b.WriteString("\n---\n## Other Files Changes\n\n")

	for _, d := range s.Diffs {
		b.WriteString(fmt.Sprintf("### `%s`\n\n", d.Ext))

		writeSection := func(label string, files []string) {
			if len(files) == 0 {
				return
			}
			b.WriteString(fmt.Sprintf("- %s:\n", label))
			sort.Strings(files)
			for _, f := range files {
				b.WriteString(fmt.Sprintf("  - %s\n", f))
			}
			b.WriteString("\n")
		}

		writeSection("Added", d.Added)
		writeSection("Modified", d.Modified)
		writeSection("Removed", d.Removed)
		writeSection("Other", d.Other)
	}

	return b.String()
}

func DiffOther(workDir, oldRef, newRef string, includeExts []string) *OtherFilesDiffSummary {
	changes := collectOtherFileChanges(workDir, oldRef, newRef, includeExts)

	var summary OtherFilesDiffSummary

	// Sorted extensions
	exts := make([]string, 0, len(changes))
	for ext := range changes {
		exts = append(exts, ext)
	}
	sort.Strings(exts)

	for _, ext := range exts {
		actions := changes[ext]
		diff := OtherFileDiff{Ext: ext}

		if files, ok := actions["Added"]; ok {
			diff.Added = append(diff.Added, files...)
		}
		if files, ok := actions["Modified"]; ok {
			diff.Modified = append(diff.Modified, files...)
		}
		if files, ok := actions["Removed"]; ok {
			diff.Removed = append(diff.Removed, files...)
		}
		if files, ok := actions["Other"]; ok {
			diff.Other = append(diff.Other, files...)
		}

		summary.Diffs = append(summary.Diffs, diff)
	}

	return &summary
}

func collectOtherFileChanges(workDir, oldRef, newRef string, includeExts []string) map[string]map[string][]string {
	changes := make(map[string]map[string][]string)

	cmd := exec.Command("git", "diff", "--name-status", oldRef, newRef)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git diff failed: %v\n", err)
		return changes
	}

	lines := strings.Split(string(out), "\n")
	includeSet := make(map[string]bool)
	for _, ext := range includeExts {
		includeSet[ext] = true
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		status, path := parts[0], parts[1]
		ext := filepath.Ext(path)
		if ext == "" {
			ext = "(no extension)"
		}
		if !includeSet[ext] {
			continue
		}
		if _, ok := changes[ext]; !ok {
			changes[ext] = make(map[string][]string)
		}

		var action string
		switch status {
		case "A":
			action = "Added"
		case "M":
			action = "Modified"
		case "D":
			action = "Removed"
		default:
			action = "Other"
		}

		changes[ext][action] = append(changes[ext][action], path)
	}

	return changes
}
