package diffs

import (
	"testing"

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

func TestAPIDiffString(t *testing.T) {
	diff := &APIDiff{
		PackagesAdded: []string{"pkg/newpkg"},
		FuncsAdded: []APIDiffRes{
			{Label: "Func", Path: "pkg/mymodule", X: "`NewClient(config.Config) -> (*Client, error)`"},
		},
		MethodsRemoved: []APIDiffRes{
			{Label: "Method", Path: "pkg/mymodule.Client", X: "`DeprecatedThing() -> (string)`"},
		},
	}

	md := diff.String()

	// Check that key strings are present
	assert.Contains(t, md, "Added Package `pkg/newpkg`")
	assert.Contains(t, md, "Added Func in `pkg/mymodule`: `NewClient(config.Config) -> (*Client, error)`")
	assert.Contains(t, md, "Removed Method in `pkg/mymodule.Client`: `DeprecatedThing() -> (string)`")
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
