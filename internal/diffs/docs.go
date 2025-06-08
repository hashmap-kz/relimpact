package diffs

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type DocInfo struct {
	Headings    []string
	Links       []string
	Images      []string
	SectionWord map[string]int // Section heading -> word count
}

type DocDiff struct {
	File              string
	HeadingsAdded     []string
	HeadingsRemoved   []string
	LinksAdded        []string
	LinksRemoved      []string
	ImagesAdded       []string
	ImagesRemoved     []string
	SectionWordChange []string // lines like "- Section `X`: 12 -> 34 words"
}

func FormatAllDocDiffs(diffs []DocDiff) string {
	if len(diffs) == 0 {
		return ""
	}

	var b bytes.Buffer
	b.WriteString("## Documentation Changes\n\n")

	for i := range diffs {
		b.WriteString(diffs[i].String())
	}

	return b.String()
}

func (d *DocDiff) String() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("### Doc File Changes: **`%s`**\n\n", d.File))

	if len(d.HeadingsAdded) > 0 {
		b.WriteString("#### Headings added:\n")
		for _, h := range d.HeadingsAdded {
			b.WriteString(fmt.Sprintf("- %s\n", h))
		}
		b.WriteString("\n")
	}
	if len(d.HeadingsRemoved) > 0 {
		b.WriteString("#### Headings removed:\n")
		for _, h := range d.HeadingsRemoved {
			b.WriteString(fmt.Sprintf("- %s\n", h))
		}
		b.WriteString("\n")
	}

	if len(d.LinksAdded) > 0 {
		b.WriteString("#### Links added:\n")
		for _, l := range d.LinksAdded {
			b.WriteString(fmt.Sprintf("- %s\n", l))
		}
		b.WriteString("\n")
	}
	if len(d.LinksRemoved) > 0 {
		b.WriteString("#### Links removed:\n")
		for _, l := range d.LinksRemoved {
			b.WriteString(fmt.Sprintf("- %s\n", l))
		}
		b.WriteString("\n")
	}

	if len(d.ImagesAdded) > 0 {
		b.WriteString("#### Images added:\n")
		for _, img := range d.ImagesAdded {
			b.WriteString(fmt.Sprintf("- %s\n", img))
		}
		b.WriteString("\n")
	}
	if len(d.ImagesRemoved) > 0 {
		b.WriteString("#### Images removed:\n")
		for _, img := range d.ImagesRemoved {
			b.WriteString(fmt.Sprintf("- %s\n", img))
		}
		b.WriteString("\n")
	}

	if len(d.SectionWordChange) > 0 {
		b.WriteString("#### Section Word Count Changes:\n")
		for _, line := range d.SectionWordChange {
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

func DiffDocs(oldDir, newDir string) []DocDiff {
	var diffs []DocDiff

	files := collectMarkdownFiles(oldDir, newDir)

	for _, file := range files {
		oldInfo := parseDoc(filepath.Join(oldDir, file))
		newInfo := parseDoc(filepath.Join(newDir, file))

		docDiff := computeDocDiff(file, oldInfo, newInfo)
		if docDiff != nil {
			diffs = append(diffs, *docDiff)
		}
	}

	return diffs
}

func computeDocDiff(file string, oldInfo, newInfo *DocInfo) *DocDiff {
	headingAdded, headingRemoved := diffSets(oldInfo.Headings, newInfo.Headings)
	linksAdded, linksRemoved := diffSets(oldInfo.Links, newInfo.Links)
	imagesAdded, imagesRemoved := diffSets(oldInfo.Images, newInfo.Images)
	sectionWordDiff := diffSectionWordCounts(oldInfo.SectionWord, newInfo.SectionWord)

	if len(headingAdded)+len(headingRemoved)+
		len(linksAdded)+len(linksRemoved)+
		len(imagesAdded)+len(imagesRemoved)+
		len(sectionWordDiff) == 0 {
		return nil
	}

	return &DocDiff{
		File:              file,
		HeadingsAdded:     headingAdded,
		HeadingsRemoved:   headingRemoved,
		LinksAdded:        linksAdded,
		LinksRemoved:      linksRemoved,
		ImagesAdded:       imagesAdded,
		ImagesRemoved:     imagesRemoved,
		SectionWordChange: sectionWordDiff,
	}
}

func collectMarkdownFiles(oldDir, newDir string) []string {
	seen := make(map[string]bool)
	var files []string

	walk := func(base string) {
		_ = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(d.Name(), ".md") {
				rel, err := filepath.Rel(base, path)
				if err == nil {
					if !seen[rel] {
						seen[rel] = true
						files = append(files, rel)
					}
				}
			}
			return nil
		})
	}

	walk(oldDir)
	walk(newDir)

	sort.Strings(files)
	return files
}

func parseDoc(path string) *DocInfo {
	content, err := os.ReadFile(path)
	if err != nil {
		return &DocInfo{SectionWord: make(map[string]int)}
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(content))

	info := &DocInfo{
		SectionWord: make(map[string]int),
	}

	currentHeading := "Document Root"

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n := n.(type) {
			case *ast.Heading:
				headingText := string(n.Text(content))
				info.Headings = append(info.Headings, headingText)
				currentHeading = headingText
			case *ast.Link:
				dest := string(n.Destination)
				info.Links = append(info.Links, dest)
			case *ast.Image:
				dest := string(n.Destination)
				info.Images = append(info.Images, dest)
			case *ast.Text:
				words := countWords(string(n.Text(content)))
				info.SectionWord[currentHeading] += words
			}
		}
		return ast.WalkContinue, nil
	})

	return info
}

func countWords(txt string) int {
	fields := strings.Fields(txt)
	return len(fields)
}

func diffSets(oldList, newList []string) (added, removed []string) {
	oldSet := make(map[string]bool)
	newSet := make(map[string]bool)

	for _, x := range oldList {
		oldSet[x] = true
	}
	for _, x := range newList {
		newSet[x] = true
	}

	for x := range newSet {
		if !oldSet[x] {
			added = append(added, x)
		}
	}
	for x := range oldSet {
		if !newSet[x] {
			removed = append(removed, x)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)

	return added, removed
}

func diffSectionWordCounts(oldCounts, newCounts map[string]int) []string {
	var lines []string
	seen := make(map[string]bool)

	for sec, oldCount := range oldCounts {
		newCount, exists := newCounts[sec]
		seen[sec] = true
		if !exists {
			lines = append(lines, fmt.Sprintf("- Section `%s`: REMOVED (%d words)", sec, oldCount))
		} else if oldCount != newCount {
			lines = append(lines, fmt.Sprintf("- Section `%s`: %d -> %d words", sec, oldCount, newCount))
		}
	}

	for sec, newCount := range newCounts {
		if !seen[sec] {
			lines = append(lines, fmt.Sprintf("- Section `%s`: ADDED (%d words)", sec, newCount))
		}
	}

	sort.Strings(lines)
	return lines
}
