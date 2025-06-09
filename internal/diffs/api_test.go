package diffs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashmap-kz/relimpact/internal/testutils"

	"github.com/hashmap-kz/relimpact/internal/gitutils"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestDiffList(t *testing.T) {
	oldList := []string{"A", "B", "C"}
	newList := []string{"B", "C", "D"}

	added, removed := diffList("Test", "pkg/mypkg", oldList, newList)

	assert.Len(t, added, 1)
	assert.Equal(t, "D", added[0].X)

	assert.Len(t, removed, 1)
	assert.Equal(t, "A", removed[0].X)
}

func TestDiffAPI(t *testing.T) {
	oldAPI := map[string]APIPackage{
		"pkg/mypkg": {
			Funcs:  []string{"Foo()"},
			Vars:   []string{"X int"},
			Consts: []string{"Y string"},
			Types: map[string]APIType{
				"MyStruct": {
					Kind:    "struct",
					Fields:  []string{"A int"},
					Methods: []string{"Bar()"},
				},
			},
		},
	}

	newAPI := map[string]APIPackage{
		"pkg/mypkg": {
			Funcs:  []string{"Foo()", "NewFoo()"},
			Vars:   []string{"X int"},
			Consts: []string{}, // removed Y
			Types: map[string]APIType{
				"MyStruct": {
					Kind:    "struct",
					Fields:  []string{"A int", "B string"},
					Methods: []string{"Bar()", "Baz()"},
				},
			},
		},
	}

	apiDiff := DiffAPI(oldAPI, newAPI)

	assert.Len(t, apiDiff.FuncsAdded, 1)
	assert.Equal(t, "NewFoo()", apiDiff.FuncsAdded[0].X)

	assert.Len(t, apiDiff.ConstsRemoved, 1)
	assert.Equal(t, "Y string", apiDiff.ConstsRemoved[0].X)

	assert.Len(t, apiDiff.FieldsAdded, 1)
	assert.Equal(t, "B string", apiDiff.FieldsAdded[0].X)

	assert.Len(t, apiDiff.MethodsAdded, 1)
	assert.Equal(t, "Baz()", apiDiff.MethodsAdded[0].X)
}

func TestAPIDiff_IntegrationTempGit(t *testing.T) {
	tmpDir := t.TempDir()
	// t.Log(tmpDir)

	// Init git repo
	testutils.RunGit(t, tmpDir, "init")
	testutils.RunGit(t, tmpDir, "config", "user.name", "Test User")
	testutils.RunGit(t, tmpDir, "config", "user.email", "test@example.com")
	testutils.RunGo(t, tmpDir, "mod", "init", "mypkg")

	// Create v1 of package
	pkgDir := filepath.Join(tmpDir, "mypkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "foo.go"), []byte(`package mypkg

func Foo() {}
`), 0o600))

	testutils.RunGit(t, tmpDir, "add", "-A")
	testutils.RunGit(t, tmpDir, "commit", "-m", "v1")
	testutils.RunGit(t, tmpDir, "tag", "v1")

	// Modify package -> add new function
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "bar.go"), []byte(`package mypkg

func Bar() {}
`), 0o600))

	testutils.RunGit(t, tmpDir, "add", "-A")
	testutils.RunGit(t, tmpDir, "commit", "-m", "add Bar")

	// Checkout old worktree
	oldWorktree := gitutils.CheckoutWorktree(tmpDir, "v1")
	defer gitutils.CleanupWorktree(tmpDir, oldWorktree)

	// Snapshot API
	oldAPI := SnapshotAPI(filepath.Join(oldWorktree, "mypkg"))
	newAPI := SnapshotAPI(filepath.Join(tmpDir, "mypkg"))

	// Diff
	apiDiff := DiffAPI(oldAPI, newAPI)

	// Assertions
	require.NotEmpty(t, apiDiff.FuncsAdded)

	foundBar := false
	for _, f := range apiDiff.FuncsAdded {
		if f.X == "Bar()" {
			foundBar = true
			break
		}
	}
	require.True(t, foundBar, "expected Bar() to be reported as added")

	// print diff for debug
	// t.Logf("API Diff:\n%s", apiDiff.String())
}

func TestAPIDiff_String(t *testing.T) {
	d := &APIDiff{
		PackagesAdded:   []string{"pkg/foo"},
		PackagesRemoved: []string{"pkg/bar"},
		FuncsAdded: []APIDiffRes{
			{Path: "pkg/foo", Label: "Funcs", X: "NewFoo() -> error"},
		},
		FuncsRemoved: []APIDiffRes{
			{Path: "pkg/bar", Label: "Funcs", X: "OldBar() -> error"},
		},
		TypesRemoved: []APIDiffRes{
			{Path: "pkg/bar", Label: "Types", X: "OldType"},
		},
	}

	out := d.String()

	// Check top-level markers
	assert.Contains(t, out, "## API Changes")
	assert.Contains(t, out, "- [Summary](#summary)")
	assert.Contains(t, out, "- [Breaking Changes](#breaking-changes)")
	assert.Contains(t, out, "- [Packages Added](#packages-added)")
	assert.Contains(t, out, "- [Packages Removed](#packages-removed)")
	assert.Contains(t, out, "- [Package Changes](#package-changes)")

	// Check Summary table line for Types Removed
	assert.Contains(t, out, "| Types Removed      |     1 |")

	// Check Breaking Changes line
	assert.Contains(t, out, "- Types Removed: **1**")

	// Check Packages Added section
	assert.Contains(t, out, "### Packages Added")
	assert.Contains(t, out, "- `pkg/foo`")

	// Check Packages Removed section
	assert.Contains(t, out, "### Packages Removed")
	assert.Contains(t, out, "- `pkg/bar`")

	// Check Package Changes header
	assert.Contains(t, out, "### Package Changes")

	// Check that Package `pkg/foo` section is present
	assert.Contains(t, out, "#### Package `pkg/foo`")
	assert.Contains(t, out, "- Added Funcs:")
	assert.Contains(t, out, "    - NewFoo() -> error")

	// Check that Package `pkg/bar` section is present
	assert.Contains(t, out, "#### Package `pkg/bar`")
	assert.Contains(t, out, "- Removed Funcs:")
	assert.Contains(t, out, "    - OldBar() -> error")
	assert.Contains(t, out, "- Removed Types:")
	assert.Contains(t, out, "    - OldType")
}
