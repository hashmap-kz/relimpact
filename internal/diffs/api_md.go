package diffs

import (
	"fmt"
	"sort"
	"strings"
)

func (d *APIDiff) String() string {
	var sb strings.Builder
	sb.WriteString("## API Changes\n")

	type summaryRow struct {
		Name    string
		Added   int
		Removed int
	}

	summary := []summaryRow{
		{"Packages", len(d.PackagesAdded), len(d.PackagesRemoved)},
		{"Funcs", len(d.FuncsAdded), len(d.FuncsRemoved)},
		{"Vars", len(d.VarsAdded), len(d.VarsRemoved)},
		{"Consts", len(d.ConstsAdded), len(d.ConstsRemoved)},
		{"Types", len(d.TypesAdded), len(d.TypesRemoved)},
		{"Fields", len(d.FieldsAdded), len(d.FieldsRemoved)},
		{"Methods", len(d.MethodsAdded), len(d.MethodsRemoved)},
	}

	var totalAdded, totalRemoved int
	for _, s := range summary {
		totalAdded += s.Added
		totalRemoved += s.Removed
	}

	// TOC
	sb.WriteString("\n- [Summary](#summary)\n")
	sb.WriteString("- [Breaking Changes](#breaking-changes)\n")
	if len(d.PackagesAdded) > 0 {
		sb.WriteString("- [Packages Added](#packages-added)\n")
	}
	if len(d.PackagesRemoved) > 0 {
		sb.WriteString("- [Packages Removed](#packages-removed)\n")
	}
	sb.WriteString("- [Package Changes](#package-changes)\n")

	// Summary table
	sb.WriteString("\n### Summary\n\n")
	sb.WriteString("| Kind     | Added | Removed |\n")
	sb.WriteString("|----------|------:|--------:|\n")
	for _, s := range summary {
		hfprintf(&sb, "| %-8s | %5d | %7d |\n", s.Name, s.Added, s.Removed)
	}
	hfprintf(&sb, "| %-8s | %5d | %7d |\n", "Total", totalAdded, totalRemoved)

	// Breaking Changes section
	sb.WriteString("\n### Breaking Changes\n\n")
	if totalRemoved == 0 {
		sb.WriteString("_No breaking changes detected._\n")
	} else {
		for _, s := range summary {
			if s.Removed > 0 {
				hfprintf(&sb, "- %s Removed: **%d**\n", s.Name, s.Removed)
			}
		}
	}

	// Packages added/removed
	writeSectionSimple := func(prefix string, packages []string) {
		if len(packages) == 0 {
			return
		}
		hfprintf(&sb, "\n### %s\n\n", prefix)
		sorted := append([]string{}, packages...)
		sort.Strings(sorted)
		for _, pkg := range sorted {
			hfprintf(&sb, "- `%s`\n", pkg)
		}
	}

	writeSectionSimple("Packages Added", d.PackagesAdded)
	writeSectionSimple("Packages Removed", d.PackagesRemoved)

	type changeKind string
	const (
		added   changeKind = "Added"
		removed changeKind = "Removed"
	)

	groupByPkgLabel := func(items []DiffItem, kind changeKind) map[string]map[string][]string {
		group := make(map[string]map[string][]string)
		for _, res := range items {
			if _, ok := group[res.Path]; !ok {
				group[res.Path] = make(map[string][]string)
			}
			key := fmt.Sprintf("%s %s", kind, res.Label)
			group[res.Path][key] = append(group[res.Path][key], res.Signature)
		}
		return group
	}

	grouped := make(map[string]map[string][]string)
	mergeGroup := func(m map[string]map[string][]string) {
		for pkg, labels := range m {
			if _, ok := grouped[pkg]; !ok {
				grouped[pkg] = make(map[string][]string)
			}
			for label, xs := range labels {
				grouped[pkg][label] = append(grouped[pkg][label], xs...)
			}
		}
	}

	mergeGroup(groupByPkgLabel(d.FuncsAdded, added))
	mergeGroup(groupByPkgLabel(d.FuncsRemoved, removed))
	mergeGroup(groupByPkgLabel(d.VarsAdded, added))
	mergeGroup(groupByPkgLabel(d.VarsRemoved, removed))
	mergeGroup(groupByPkgLabel(d.ConstsAdded, added))
	mergeGroup(groupByPkgLabel(d.ConstsRemoved, removed))
	mergeGroup(groupByPkgLabel(d.TypesAdded, added))
	mergeGroup(groupByPkgLabel(d.TypesRemoved, removed))
	mergeGroup(groupByPkgLabel(d.FieldsAdded, added))
	mergeGroup(groupByPkgLabel(d.FieldsRemoved, removed))
	mergeGroup(groupByPkgLabel(d.MethodsAdded, added))
	mergeGroup(groupByPkgLabel(d.MethodsRemoved, removed))

	if len(grouped) > 0 {
		sb.WriteString("\n### Package Changes\n")
		pkgs := make([]string, 0, len(grouped))
		for pkg := range grouped {
			pkgs = append(pkgs, pkg)
		}
		sort.Strings(pkgs)

		for _, pkg := range pkgs {
			hfprintf(&sb, "\n#### Package `%s`\n\n", pkg)
			sb.WriteString("<details>\n<summary>Click to expand</summary>\n\n")

			labels := make([]string, 0, len(grouped[pkg]))
			for label := range grouped[pkg] {
				labels = append(labels, label)
			}
			sort.Strings(labels)

			for _, label := range labels {
				hfprintf(&sb, "- %s:\n", label)
				xs := grouped[pkg][label]
				sort.Strings(xs)
				for _, x := range xs {
					hfprintf(&sb, "    - %s\n", x)
				}
			}

			sb.WriteString("\n</details>\n")
		}
	}

	return sb.String()
}
