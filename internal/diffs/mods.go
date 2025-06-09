package diffs

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
)

type GoModDiff struct {
	DependenciesAdded   []string
	DependenciesRemoved []string
	DependenciesUpdated []string
}

func (d *GoModDiff) String() string {
	var b strings.Builder
	b.WriteString("\n---\n## go.mod Changes\n\n")
	b.WriteString("<details>\n<summary>Click to expand</summary>\n\n")

	if len(d.DependenciesAdded) > 0 {
		b.WriteString("### Dependencies added\n")
		for _, line := range d.DependenciesAdded {
			b.WriteString("- " + line + "\n")
		}
		b.WriteString("\n")
	}

	if len(d.DependenciesRemoved) > 0 {
		b.WriteString("### Dependencies removed\n")
		for _, line := range d.DependenciesRemoved {
			b.WriteString("- " + line + "\n")
		}
		b.WriteString("\n")
	}

	if len(d.DependenciesUpdated) > 0 {
		b.WriteString("### Dependencies updated\n")
		for _, line := range d.DependenciesUpdated {
			b.WriteString("- " + line + "\n")
		}
		b.WriteString("\n")
	}

	if len(d.DependenciesAdded)+len(d.DependenciesRemoved)+len(d.DependenciesUpdated) == 0 {
		b.WriteString("_No changes detected._\n\n")
	}

	b.WriteString("</details>\n\n")
	return b.String()
}

func DiffGoMod(oldDir, newDir string) GoModDiff {
	oldMod := parseGoMod(oldDir + "/go.mod")
	newMod := parseGoMod(newDir + "/go.mod")

	oldDeps := make(map[string]string)
	for _, r := range oldMod.Require {
		oldDeps[r.Mod.Path] = r.Mod.Version
	}
	newDeps := make(map[string]string)
	for _, r := range newMod.Require {
		newDeps[r.Mod.Path] = r.Mod.Version
	}

	diff := GoModDiff{}

	for path, version := range newDeps {
		if _, exists := oldDeps[path]; !exists {
			diff.DependenciesAdded = append(diff.DependenciesAdded, fmt.Sprintf("%s %s", path, version))
		}
	}

	for path, version := range oldDeps {
		if _, exists := newDeps[path]; !exists {
			diff.DependenciesRemoved = append(diff.DependenciesRemoved, fmt.Sprintf("%s %s", path, version))
		}
	}

	for path, oldVer := range oldDeps {
		if newVer, exists := newDeps[path]; exists && newVer != oldVer {
			diff.DependenciesUpdated = append(diff.DependenciesUpdated, fmt.Sprintf("%s %s -> %s", path, oldVer, newVer))
		}
	}

	oldReplace := make(map[string]string)
	for _, r := range oldMod.Replace {
		key := r.Old.Path
		if r.Old.Version != "" {
			key += "@" + r.Old.Version
		}
		val := r.New.Path
		if r.New.Version != "" {
			val += "@" + r.New.Version
		}
		oldReplace[key] = val
	}

	newReplace := make(map[string]string)
	for _, r := range newMod.Replace {
		key := r.Old.Path
		if r.Old.Version != "" {
			key += "@" + r.Old.Version
		}
		val := r.New.Path
		if r.New.Version != "" {
			val += "@" + r.New.Version
		}
		newReplace[key] = val
	}

	sort.Strings(diff.DependenciesAdded)
	sort.Strings(diff.DependenciesRemoved)
	sort.Strings(diff.DependenciesUpdated)
	return diff
}

func parseGoMod(path string) *modfile.File {
	data, err := os.ReadFile(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: could not read %s: %v\n", path, err)
		return &modfile.File{}
	}

	f, err := modfile.Parse(path, data, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: could not parse %s: %v\n", path, err)
		return &modfile.File{}
	}

	return f
}
