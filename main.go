package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type GoPackage struct {
	ImportPath string
	Name       string
	Export     []string
}

func main() {
	oldRef := flag.String("old", "", "Old git ref (required)")
	newRef := flag.String("new", "", "New git ref (required)")
	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		fmt.Println("Usage: relimpact --old <old-ref> --new <new-ref>")
		os.Exit(1)
	}

	tmpOld := checkoutWorktree(*oldRef)
	defer cleanupWorktree(tmpOld)

	tmpNew := checkoutWorktree(*newRef)
	defer cleanupWorktree(tmpNew)

	oldAPI := listGoAPI(tmpOld)
	newAPI := listGoAPI(tmpNew)

	docsDiff := diffDocs(tmpOld, tmpNew)

	reportMarkdown(oldAPI, newAPI, docsDiff)
}

func checkoutWorktree(ref string) string {
	tmpDir, err := os.MkdirTemp("", "relimpact-"+ref)
	if err != nil {
		panic(err)
	}
	run("git", "worktree", "add", "--detach", tmpDir, ref)
	return tmpDir
}

func cleanupWorktree(path string) {
	run("git", "worktree", "remove", "--force", path)
}

func listGoAPI(dir string) map[string]GoPackage {
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "go list failed in %s: %v\n", dir, err)
		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(out))
	apis := make(map[string]GoPackage)
	for dec.More() {
		var pkg struct {
			ImportPath string
			Name       string
			Exports    []string
		}
		if err := dec.Decode(&pkg); err != nil {
			panic(err)
		}
		apis[pkg.ImportPath] = GoPackage{
			ImportPath: pkg.ImportPath,
			Name:       pkg.Name,
			Export:     pkg.Exports,
		}
	}
	return apis
}

func diffDocs(oldDir, newDir string) []string {
	var diffLines []string
	cmd := exec.Command("git", "diff", "--no-index", "--", filepath.Join(oldDir, "docs"), filepath.Join(newDir, "docs"))
	out, _ := cmd.CombinedOutput()
	if len(out) > 0 {
		diffLines = append(diffLines, "## Docs Changes\n```diff\n"+string(out)+"\n```\n")
	}

	// README diff
	cmd = exec.Command("git", "diff", "--no-index", "--", filepath.Join(oldDir, "README.md"), filepath.Join(newDir, "README.md"))
	out, _ = cmd.CombinedOutput()
	if len(out) > 0 {
		diffLines = append(diffLines, "## README.md Changes\n```diff\n"+string(out)+"\n```\n")
	}
	return diffLines
}

func reportMarkdown(oldAPI, newAPI map[string]GoPackage, docsDiff []string) {
	fmt.Println("# Release Impact Report")
	fmt.Println()

	// API Diff
	fmt.Println("## API Changes")
	for path, newPkg := range newAPI {
		oldPkg, ok := oldAPI[path]
		if !ok {
			fmt.Printf("- Package added: `%s`\n", path)
			continue
		}
		oldSet := make(map[string]bool)
		for _, e := range oldPkg.Export {
			oldSet[e] = true
		}
		for _, e := range newPkg.Export {
			if !oldSet[e] {
				fmt.Printf("- Export added in `%s`: `%s`\n", path, e)
			}
		}
	}

	for path, oldPkg := range oldAPI {
		if _, ok := newAPI[path]; !ok {
			fmt.Printf("- Package removed: `%s`\n", path)
		}
		newSet := make(map[string]bool)
		if newPkg, exists := newAPI[path]; exists {
			for _, e := range newPkg.Export {
				newSet[e] = true
			}
		}
		for _, e := range oldPkg.Export {
			if !newSet[e] {
				fmt.Printf("- Export removed from `%s`: `%s`\n", path, e)
			}
		}
	}

	// Docs diff
	for _, d := range docsDiff {
		fmt.Println(d)
	}
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
