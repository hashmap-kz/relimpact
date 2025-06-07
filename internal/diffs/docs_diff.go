package diffs

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type DocInfo struct {
	Headings []string
	Links    []string
	Images   []string
}

func DiffDocs(oldDir, newDir string) []string {
	var diffs []string

	files := collectMarkdownFiles(oldDir, newDir)

	for _, file := range files {
		oldInfo := parseDoc(filepath.Join(oldDir, file))
		newInfo := parseDoc(filepath.Join(newDir, file))

		section := diffDoc(file, oldInfo, newInfo)
		if section != "" {
			diffs = append(diffs, section)
		}
	}

	return diffs
}

func collectMarkdownFiles(oldDir, newDir string) []string {
	seen := make(map[string]bool)
	var files []string

	walk := func(base string) {
		filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(d.Name(), ".md") {
				rel, err := filepath.Rel(base, path)
				if err == nil && !seen[rel] {
					seen[rel] = true
					files = append(files, rel)
				}
			}
			return nil
		})
	}

	walk(filepath.Join(oldDir, "docs"))
	walk(filepath.Join(newDir, "docs"))
	walk(oldDir)
	walk(newDir)

	return files
}

func parseDoc(path string) DocInfo {
	content, err := os.ReadFile(path)
	if err != nil {
		return DocInfo{}
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(content))

	var info DocInfo

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := n.(type) {
		case *ast.Heading:
			txt := string(n.Text(content))
			info.Headings = append(info.Headings, txt)
		case *ast.Link:
			dest := string(n.Destination)
			info.Links = append(info.Links, dest)
		case *ast.Image:
			dest := string(n.Destination)
			info.Images = append(info.Images, dest)
		}

		return ast.WalkContinue, nil
	})

	return info
}

func diffDoc(file string, oldInfo, newInfo DocInfo) string {
	var b bytes.Buffer

	headingAdded, headingRemoved := diffSets(oldInfo.Headings, newInfo.Headings)
	linksAdded, linksRemoved := diffSets(oldInfo.Links, newInfo.Links)
	imagesAdded, imagesRemoved := diffSets(oldInfo.Images, newInfo.Images)

	if len(headingAdded)+len(headingRemoved)+len(linksAdded)+len(linksRemoved)+len(imagesAdded)+len(imagesRemoved) == 0 {
		return ""
	}

	b.WriteString(fmt.Sprintf("## Documentation Changes: `%s`\n\n", file))

	if len(headingAdded) > 0 {
		b.WriteString("### Headings added:\n")
		for _, h := range headingAdded {
			b.WriteString(fmt.Sprintf("- %s\n", h))
		}
		b.WriteString("\n")
	}
	if len(headingRemoved) > 0 {
		b.WriteString("### Headings removed:\n")
		for _, h := range headingRemoved {
			b.WriteString(fmt.Sprintf("- %s\n", h))
		}
		b.WriteString("\n")
	}

	if len(linksAdded) > 0 {
		b.WriteString("### Links added:\n")
		for _, l := range linksAdded {
			b.WriteString(fmt.Sprintf("- %s\n", l))
		}
		b.WriteString("\n")
	}
	if len(linksRemoved) > 0 {
		b.WriteString("### Links removed:\n")
		for _, l := range linksRemoved {
			b.WriteString(fmt.Sprintf("- %s\n", l))
		}
		b.WriteString("\n")
	}

	if len(imagesAdded) > 0 {
		b.WriteString("### Images added:\n")
		for _, img := range imagesAdded {
			b.WriteString(fmt.Sprintf("- %s\n", img))
		}
		b.WriteString("\n")
	}
	if len(imagesRemoved) > 0 {
		b.WriteString("### Images removed:\n")
		for _, img := range imagesRemoved {
			b.WriteString(fmt.Sprintf("- %s\n", img))
		}
		b.WriteString("\n")
	}

	return b.String()
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

	return added, removed
}
