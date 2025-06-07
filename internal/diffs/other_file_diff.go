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

// DiffOtherFiles generates a Markdown section with Other Files Changes.
// includeExts -> which extensions to include (e.g. .sh, .sql, .yaml, .json, .conf, etc)
func DiffOtherFiles(oldRef, newRef string, includeExts []string) string {
	changes := collectOtherFileChanges(oldRef, newRef, includeExts)
	if len(changes) == 0 {
		return ""
	}

	var b bytes.Buffer
	b.WriteString("### Other Files Changes\n\n")

	// Sorted extensions
	var exts []string
	for ext := range changes {
		exts = append(exts, ext)
	}
	sort.Strings(exts)

	for _, ext := range exts {
		actions := changes[ext]
		b.WriteString(fmt.Sprintf("#### %s\n\n", ext))

		// Sorted actions
		actionOrder := []string{"Added", "Modified", "Removed", "Other"}
		for _, action := range actionOrder {
			files, ok := actions[action]
			if !ok || len(files) == 0 {
				continue
			}
			sort.Strings(files)
			b.WriteString(fmt.Sprintf("- %s:\n", action))
			for _, f := range files {
				b.WriteString(fmt.Sprintf("  - %s\n", f))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

func collectOtherFileChanges(oldRef, newRef string, includeExts []string) map[string]map[string][]string {
	changes := make(map[string]map[string][]string)

	cmd := exec.Command("git", "diff", "--name-status", oldRef, newRef)
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
