package diffs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/token"
	"go/types"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type APIPackage struct {
	Funcs  []string           `json:"funcs"`
	Vars   []string           `json:"vars"`
	Consts []string           `json:"consts"`
	Types  map[string]APIType `json:"types"`
}

type APIType struct {
	Kind    string   `json:"kind"`   // struct, interface, etc.
	Fields  []string `json:"fields"` // for structs
	Methods []string `json:"methods"`
}

type APIDiffRes struct {
	Label string
	Path  string
	X     string
}

type APIDiff struct {
	PackagesAdded   []string     `json:"packages_added,omitempty"`
	PackagesRemoved []string     `json:"packages_removed,omitempty"`
	FuncsAdded      []APIDiffRes `json:"funcs_added,omitempty"`
	FuncsRemoved    []APIDiffRes `json:"funcs_removed,omitempty"`
	VarsAdded       []APIDiffRes `json:"vars_added,omitempty"`
	VarsRemoved     []APIDiffRes `json:"vars_removed,omitempty"`
	ConstsAdded     []APIDiffRes `json:"consts_added,omitempty"`
	ConstsRemoved   []APIDiffRes `json:"consts_removed,omitempty"`
	TypesAdded      []APIDiffRes `json:"types_added,omitempty"`
	TypesRemoved    []APIDiffRes `json:"types_removed,omitempty"`
	FieldsAdded     []APIDiffRes `json:"fields_added,omitempty"`
	FieldsRemoved   []APIDiffRes `json:"fields_removed,omitempty"`
	MethodsAdded    []APIDiffRes `json:"methods_added,omitempty"`
	MethodsRemoved  []APIDiffRes `json:"methods_removed,omitempty"`
}

func (d *APIDiff) String() string {
	var sb strings.Builder
	sb.WriteString("## API Changes\n")

	// Build Summary table data
	summary := []struct {
		Name  string
		Count int
	}{
		{"Packages Added", len(d.PackagesAdded)},
		{"Packages Removed", len(d.PackagesRemoved)},
		{"Funcs Added", len(d.FuncsAdded)},
		{"Funcs Removed", len(d.FuncsRemoved)},
		{"Vars Added", len(d.VarsAdded)},
		{"Vars Removed", len(d.VarsRemoved)},
		{"Consts Added", len(d.ConstsAdded)},
		{"Consts Removed", len(d.ConstsRemoved)},
		{"Types Added", len(d.TypesAdded)},
		{"Types Removed", len(d.TypesRemoved)},
		{"Fields Added", len(d.FieldsAdded)},
		{"Fields Removed", len(d.FieldsRemoved)},
		{"Methods Added", len(d.MethodsAdded)},
		{"Methods Removed", len(d.MethodsRemoved)},
	}

	// Compute total changes and breaking changes
	var totalChanges int
	breakingChanges := map[string]int{}
	for _, s := range summary {
		totalChanges += s.Count
		switch s.Name {
		case "Packages Removed", "Funcs Removed", "Vars Removed", "Consts Removed", "Types Removed", "Fields Removed", "Methods Removed":
			breakingChanges[s.Name] = s.Count
		}
	}

	// Write TOC
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
	sb.WriteString("|   Kind of Change   | Count |\n")
	sb.WriteString("|--------------------|-------|\n")
	// width:20,7; minus two spaces:18,5;
	for _, s := range summary {
		sb.WriteString(fmt.Sprintf("| %-18s | %5d |\n", s.Name, s.Count))
	}
	sb.WriteString(fmt.Sprintf("| %-18s | %5d |\n", "Total Changes", totalChanges))

	// Breaking Changes section
	sb.WriteString("\n### Breaking Changes\n\n")
	if len(breakingChanges) == 0 {
		sb.WriteString("_No breaking changes detected._\n")
	} else {
		for k, v := range breakingChanges {
			if v > 0 {
				sb.WriteString(fmt.Sprintf("- %s: **%d**\n", k, v))
			}
		}
	}

	// Packages added/removed
	writeSectionSimple := func(prefix string, packages []string) {
		if len(packages) == 0 {
			return
		}
		sb.WriteString(fmt.Sprintf("\n### %s\n\n", prefix))
		sorted := append([]string{}, packages...)
		sort.Strings(sorted)
		for _, pkg := range sorted {
			sb.WriteString(fmt.Sprintf("- `%s`\n", pkg))
		}
	}

	writeSectionSimple("Packages Added", d.PackagesAdded)
	writeSectionSimple("Packages Removed", d.PackagesRemoved)

	// Group APIDiffRes items by package, by label, by added/removed
	type changeKind string
	const (
		added   changeKind = "Added"
		removed changeKind = "Removed"
	)

	groupByPkgLabel := func(items []APIDiffRes, kind changeKind) map[string]map[string][]string {
		group := make(map[string]map[string][]string)
		for _, res := range items {
			if _, ok := group[res.Path]; !ok {
				group[res.Path] = make(map[string][]string)
			}
			group[res.Path][fmt.Sprintf("%s %s", kind, res.Label)] = append(group[res.Path][fmt.Sprintf("%s %s", kind, res.Label)], res.X)
		}
		return group
	}

	// Build groupings
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

	// Print per package with collapsible <details>
	if len(grouped) > 0 {
		sb.WriteString("\n### Package Changes\n")
		pkgs := make([]string, 0, len(grouped))
		for pkg := range grouped {
			pkgs = append(pkgs, pkg)
		}
		sort.Strings(pkgs)
		for _, pkg := range pkgs {
			sb.WriteString(fmt.Sprintf("\n#### Package `%s`\n\n", pkg))
			sb.WriteString("<details>\n<summary>Click to expand</summary>\n\n")
			labels := make([]string, 0, len(grouped[pkg]))
			for label := range grouped[pkg] {
				labels = append(labels, label)
			}
			sort.Strings(labels)
			for _, label := range labels {
				sb.WriteString(fmt.Sprintf("- %s:\n", label))
				xs := grouped[pkg][label]
				sort.Strings(xs)
				for _, x := range xs {
					sb.WriteString(fmt.Sprintf("    - %s\n", x))
				}
			}
			sb.WriteString("\n</details>\n")
		}
	}

	return sb.String()
}

func SnapshotAPI(dir string) map[string]APIPackage {
	// TODO: debuglog

	sha := getGitCommitSHA(dir)
	cachePath := filepath.Join(os.TempDir(), "relimpact-api-cache", sha+".json")

	// Try to load from cache
	if data, err := os.ReadFile(cachePath); err == nil {
		var cached map[string]APIPackage
		if json.Unmarshal(data, &cached) == nil {
			return cached
		}
	}

	//nolint:gocritic
	// cfg := &packages.Config{
	// 	Mode: packages.NeedName |
	// 		packages.NeedTypes |
	// 		packages.NeedSyntax |
	// 		packages.NeedTypesInfo |
	// 		packages.NeedImports,
	// 	Dir: dir,
	// }

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports,
		Dir:  dir,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatal(err)
	}

	modulePath := getModulePath(dir)
	api := make(map[string]APIPackage)

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "Errors in package %s:\n", pkg.PkgPath)
			for _, err := range pkg.Errors {
				fmt.Fprintf(os.Stderr, "  %v\n", err)
			}
			continue
		}

		if !strings.HasPrefix(pkg.PkgPath, modulePath) {
			continue
		}

		apkg := APIPackage{
			Funcs:  []string{},
			Vars:   []string{},
			Consts: []string{},
			Types:  make(map[string]APIType),
		}

		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			if !token.IsExported(name) {
				continue
			}

			obj := scope.Lookup(name)
			switch o := obj.(type) {
			case *types.Func:
				if o.Type() != nil {
					//nolint:errcheck
					sig := o.Type().(*types.Signature)
					apkg.Funcs = append(apkg.Funcs, name+signatureString(sig))
				}
			case *types.Var:
				if o.IsField() {
					continue
				}
				apkg.Vars = append(apkg.Vars, name+" "+o.Type().String())
			case *types.Const:
				apkg.Consts = append(apkg.Consts, name+" "+o.Type().String())
			case *types.TypeName:
				t := o.Type().Underlying()
				atype := APIType{}
				switch ut := t.(type) {
				case *types.Struct:
					atype.Kind = "struct"
					for i := 0; i < ut.NumFields(); i++ {
						f := ut.Field(i)
						if f.Exported() {
							atype.Fields = append(atype.Fields, f.Name()+" "+f.Type().String())
						}
					}
				case *types.Interface:
					atype.Kind = "interface"
					for i := 0; i < ut.NumMethods(); i++ {
						m := ut.Method(i)
						//nolint:errcheck
						atype.Methods = append(atype.Methods, m.Name()+signatureString(m.Type().(*types.Signature)))
					}
				default:
					atype.Kind = fmt.Sprintf("%T", ut)
				}

				methodSet := types.NewMethodSet(o.Type())
				for i := 0; i < methodSet.Len(); i++ {
					m := methodSet.At(i)
					if m.Obj().Exported() {
						//nolint:errcheck
						atype.Methods = append(atype.Methods, m.Obj().Name()+signatureString(m.Obj().Type().(*types.Signature)))
					}
				}

				apkg.Types[name] = atype
			}
		}

		api[pkg.PkgPath] = apkg
	}

	// Save to cache
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o750); err == nil {
		if data, err := json.MarshalIndent(api, "", "  "); err == nil {
			//nolint:errcheck
			_ = os.WriteFile(cachePath, data, 0o600)
		}
	}

	return api
}

func signatureString(sig *types.Signature) string {
	var b bytes.Buffer
	b.WriteString("(")
	for i := 0; i < sig.Params().Len(); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(sig.Params().At(i).Type().String())
	}
	b.WriteString(")")
	if sig.Results().Len() > 0 {
		b.WriteString(" -> (")
		for i := 0; i < sig.Results().Len(); i++ {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(sig.Results().At(i).Type().String())
		}
		b.WriteString(")")
	}
	return b.String()
}

func DiffAPI(oldAPI, newAPI map[string]APIPackage) *APIDiff {
	apiDiffResult := &APIDiff{}

	for path, newPkg := range newAPI {
		oldPkg, ok := oldAPI[path]

		// packages +
		if !ok {
			apiDiffResult.PackagesAdded = append(apiDiffResult.PackagesAdded, path)
			continue
		}

		// Funcs
		funcsAdd, funcsRem := diffList("Funcs", path, oldPkg.Funcs, newPkg.Funcs)
		apiDiffResult.FuncsAdded = append(apiDiffResult.FuncsAdded, funcsAdd...)
		apiDiffResult.FuncsRemoved = append(apiDiffResult.FuncsRemoved, funcsRem...)

		// Vars
		varsAdded, varsRemoved := diffList("Vars", path, oldPkg.Vars, newPkg.Vars)
		apiDiffResult.VarsAdded = append(apiDiffResult.VarsAdded, varsAdded...)
		apiDiffResult.VarsRemoved = append(apiDiffResult.VarsRemoved, varsRemoved...)

		// Consts
		constsAdded, constsRemoved := diffList("Consts", path, oldPkg.Consts, newPkg.Consts)
		apiDiffResult.ConstsAdded = append(apiDiffResult.ConstsAdded, constsAdded...)
		apiDiffResult.ConstsRemoved = append(apiDiffResult.ConstsRemoved, constsRemoved...)

		// Types
		for tname, newType := range newPkg.Types {
			oldType, ok := oldPkg.Types[tname]
			if !ok {
				// types +
				apiDiffResult.TypesAdded = append(apiDiffResult.TypesAdded, APIDiffRes{
					Label: "Type",
					Path:  path,
					X:     tname,
				})
				continue
			}

			// fields
			fieldsAdded, fieldsRemoved := diffList(fmt.Sprintf("Type `%s` Fields", tname), path, oldType.Fields, newType.Fields)
			apiDiffResult.FieldsAdded = append(apiDiffResult.FieldsAdded, fieldsAdded...)
			apiDiffResult.FieldsRemoved = append(apiDiffResult.FieldsRemoved, fieldsRemoved...)

			// methods
			methodsAdded, methodsRemoved := diffList(fmt.Sprintf("Type `%s` Methods", tname), path, oldType.Methods, newType.Methods)
			apiDiffResult.MethodsAdded = append(apiDiffResult.MethodsAdded, methodsAdded...)
			apiDiffResult.MethodsRemoved = append(apiDiffResult.MethodsRemoved, methodsRemoved...)
		}
		// types -
		for tname := range oldPkg.Types {
			if _, ok := newPkg.Types[tname]; !ok {
				apiDiffResult.TypesRemoved = append(apiDiffResult.TypesRemoved, APIDiffRes{
					Label: "Type",
					Path:  path,
					X:     tname,
				})
			}
		}
	}

	// packages -
	for path := range oldAPI {
		if _, ok := newAPI[path]; !ok {
			apiDiffResult.PackagesRemoved = append(apiDiffResult.PackagesRemoved, path)
		}
	}

	return apiDiffResult
}

func diffList(label, path string, oldList, newList []string) (added, removed []APIDiffRes) {
	oldSet := make(map[string]bool)
	for _, x := range oldList {
		oldSet[x] = true
	}
	newSet := make(map[string]bool)
	for _, x := range newList {
		newSet[x] = true
	}

	for x := range newSet {
		if !oldSet[x] {
			added = append(added, APIDiffRes{
				Label: label,
				Path:  path,
				X:     x,
			})
		}
	}
	for x := range oldSet {
		if !newSet[x] {
			removed = append(removed, APIDiffRes{
				Label: label,
				Path:  path,
				X:     x,
			})
		}
	}

	return added, removed
}

func getModulePath(dir string) string {
	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getGitCommitSHA(dir string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to get commit SHA in %s: %v", dir, err)
	}
	return strings.TrimSpace(string(out))
}
