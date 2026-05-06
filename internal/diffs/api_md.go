package diffs

import (
	"fmt"
	"strings"
)

func (d *APIDiff) String() string {
	return d.Markdown(ReportMetadata{})
}

func (d *APIDiff) Markdown(meta ReportMetadata) string {
	data := d.buildReportData(meta)

	var b strings.Builder
	writeMarkdownHeader(&b, data)
	writeMarkdownContents(&b, data)

	if len(data.Breaking) > 0 {
		writeMarkdownPageBlock(&b, "Breaking changes", data.Breaking)
	}
	if len(data.Added) > 0 {
		writeMarkdownPageBlock(&b, "New API", data.Added)
	}
	if len(data.Breaking) == 0 && len(data.Added) == 0 {
		b.WriteString("\nNo public API changes detected.\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func writeMarkdownHeader(b *strings.Builder, data reportData) {
	b.WriteString("# API compatibility report\n\n")
	fmt.Fprintf(b, "`%s` → `%s`\n\n", data.Meta.OldRef, data.Meta.NewRef)

	if data.Summary.Breaking > 0 {
		b.WriteString("> ⚠️ **Breaking API changes detected.** Review before release.\n\n")
	} else {
		b.WriteString("> ✅ **No breaking API changes detected.**\n\n")
	}

	b.WriteString("| Breaking | Changed | Removed | Added | Packages |\n")
	b.WriteString("|---:|---:|---:|---:|---:|\n")
	fmt.Fprintf(
		b,
		"| %d | %d | %d | %d | %d |\n\n",
		data.Summary.Breaking,
		data.Summary.Changed,
		data.Summary.Removed,
		data.Summary.Added,
		data.Summary.ChangedPackages,
	)
}

func writeMarkdownContents(b *strings.Builder, data reportData) {
	if len(data.Sidebar.Breaking)+len(data.Sidebar.Added) <= 3 {
		return
	}

	b.WriteString("## Contents\n\n")
	if len(data.Sidebar.Breaking) > 0 {
		b.WriteString("### Breaking changes\n\n")
		for _, entry := range data.Sidebar.Breaking {
			fmt.Fprintf(b, "- `%s` — %d\n", entry.Package, entry.Count)
		}
		b.WriteByte('\n')
	}
	if len(data.Sidebar.Added) > 0 {
		b.WriteString("### New API\n\n")
		for _, entry := range data.Sidebar.Added {
			fmt.Fprintf(b, "- `%s` — %d\n", entry.Package, entry.Count)
		}
		b.WriteByte('\n')
	}
}

func writeMarkdownPageBlock(b *strings.Builder, title string, sections []pkgSectionData) {
	fmt.Fprintf(b, "## %s\n\n", title)
	for _, section := range sections {
		writeMarkdownPackageSection(b, section)
	}
}

func writeMarkdownPackageSection(b *strings.Builder, section pkgSectionData) {
	fmt.Fprintf(b, "### `%s`\n\n", section.Package)
	for _, status := range section.StatusSections {
		writeMarkdownStatusSection(b, status)
	}
}

func writeMarkdownStatusSection(b *strings.Builder, section statusSectionData) {
	fmt.Fprintf(b, "#### %s\n\n", section.Title)

	for _, card := range section.ChangeCards {
		writeMarkdownChangeCard(b, card)
	}
	for _, group := range section.KindGroups {
		writeMarkdownKindGroup(b, group)
	}
	for _, block := range section.TypeDefBlocks {
		writeMarkdownTypeDefBlock(b, block)
	}
	for _, diff := range section.StructFieldDiffs {
		writeMarkdownStructFieldDiff(b, diff)
	}
}

func writeMarkdownChangeCard(b *strings.Builder, card changeCardData) {
	fmt.Fprintf(b, "**%s**\n\n", card.Name)
	b.WriteString("```diff\n")
	writePrefixedMarkdownBlock(b, "-", card.OldSignature)
	writePrefixedMarkdownBlock(b, "+", card.NewSignature)
	b.WriteString("```\n\n")
}

func writeMarkdownKindGroup(b *strings.Builder, group kindGroupData) {
	fmt.Fprintf(b, "**%s**\n\n", group.KindLabel)
	b.WriteString("```diff\n")
	for _, row := range group.Rows {
		line := row.Code
		if row.Type != "" {
			line += " " + row.Type
		}
		fmt.Fprintf(b, "%s %s\n", row.Prefix, line)
	}
	b.WriteString("```\n\n")
}

func writeMarkdownTypeDefBlock(b *strings.Builder, block typeDefBlockData) {
	fmt.Fprintf(b, "**%s**\n\n", block.KindLabel)
	b.WriteString("```diff\n")
	b.WriteString(strings.TrimRight(block.Text, "\n"))
	b.WriteString("\n```\n\n")
}

func writeMarkdownStructFieldDiff(b *strings.Builder, diff structFieldDiffData) {
	fmt.Fprintf(b, "**%s**\n\n", diff.Title)
	b.WriteString("```diff\n")
	b.WriteString(strings.TrimRight(diff.Text, "\n"))
	b.WriteString("\n```\n\n")
}

func writePrefixedMarkdownBlock(b *strings.Builder, prefix, text string) {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for _, line := range lines {
		fmt.Fprintf(b, "%s %s\n", prefix, line)
	}
}
