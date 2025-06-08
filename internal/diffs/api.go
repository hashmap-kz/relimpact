package diffs

import (
	"bytes"
	"fmt"
	"go/types"
	"log"
	"os"
	"os/exec"
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

func (dr *APIDiffRes) String() string {
	return dr.Path + "/" + dr.X
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

	writeSection := func(prefix string, items []APIDiffRes) {
		for _, res := range items {
			sb.WriteString("- ")
			sb.WriteString(prefix)
			sb.WriteString(res.Label)
			sb.WriteString(" in `")
			sb.WriteString(res.Path)
			sb.WriteString("`: ")
			sb.WriteString(res.X)
			sb.WriteString("\n")
		}
	}

	writeSectionSimple := func(prefix string, packages []string) {
		for _, pkg := range packages {
			sb.WriteString("- ")
			sb.WriteString(prefix)
			sb.WriteString(" `")
			sb.WriteString(pkg)
			sb.WriteString("`\n")
		}
	}

	// Packages
	writeSectionSimple("Added Package", d.PackagesAdded)
	writeSectionSimple("Removed Package", d.PackagesRemoved)

	// Funcs
	writeSection("Added ", d.FuncsAdded)
	writeSection("Removed ", d.FuncsRemoved)

	// Vars
	writeSection("Added ", d.VarsAdded)
	writeSection("Removed ", d.VarsRemoved)

	// Consts
	writeSection("Added ", d.ConstsAdded)
	writeSection("Removed ", d.ConstsRemoved)

	// Types
	writeSection("Added ", d.TypesAdded)
	writeSection("Removed ", d.TypesRemoved)

	// Fields
	writeSection("Added ", d.FieldsAdded)
	writeSection("Removed ", d.FieldsRemoved)

	// Methods
	writeSection("Added ", d.MethodsAdded)
	writeSection("Removed ", d.MethodsRemoved)

	return sb.String()
}

func SnapshotAPI(dir string) map[string]APIPackage {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedImports,
		Dir: dir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatal(err)
	}

	api := make(map[string]APIPackage)
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "Errors in package %s:\n", pkg.PkgPath)
			for _, err := range pkg.Errors {
				fmt.Fprintf(os.Stderr, "  %v\n", err)
			}
			continue
		}

		if !strings.HasPrefix(pkg.PkgPath, getModulePath(dir)) {
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
			// TODO: check type is exporter, i.e.: public API

			obj := scope.Lookup(name)
			switch o := obj.(type) {
			case *types.Func:
				if o.Type() != nil {
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
						atype.Methods = append(atype.Methods, m.Name()+signatureString(m.Type().(*types.Signature)))
					}
				default:
					atype.Kind = fmt.Sprintf("%T", ut)
				}

				// Also collect methods of named types
				methodSet := types.NewMethodSet(o.Type())
				for i := 0; i < methodSet.Len(); i++ {
					m := methodSet.At(i)
					if m.Obj().Exported() {
						atype.Methods = append(atype.Methods, m.Obj().Name()+signatureString(m.Obj().Type().(*types.Signature)))
					}
				}

				apkg.Types[name] = atype
			}
		}

		api[pkg.PkgPath] = apkg
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
