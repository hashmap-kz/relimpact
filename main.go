package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/types"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/hashmap-kz/relimpact/internal/diffs"

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

func main() {
	oldRef := flag.String("old", "", "Old git ref")
	newRef := flag.String("new", "", "New git ref")
	flag.Parse()

	if *oldRef == "" || *newRef == "" {
		log.Fatal("Usage: apidiff --old <ref> --new <ref>")
	}

	tmpOld := checkoutWorktree(*oldRef)
	defer cleanupWorktree(tmpOld)

	tmpNew := checkoutWorktree(*newRef)
	defer cleanupWorktree(tmpNew)

	oldAPI := snapshotAPI(tmpOld)
	newAPI := snapshotAPI(tmpNew)

	diffAPI(oldAPI, newAPI)

	// docs diffs

	docsDiffs := diffs.DiffDocs(tmpOld, tmpNew)
	if len(docsDiffs) > 0 {
		for _, section := range docsDiffs {
			fmt.Println(section)
		}
	}

	// others

	// Define which file types you care about
	// TODO: configurable
	includeExts := []string{".sh", ".sql", ".json", ".yaml", ".yml", ".conf", ".ini", ".txt", ".csv"}

	otherSection := diffs.DiffOtherFiles(*oldRef, *newRef, includeExts)
	if otherSection != "" {
		fmt.Println(otherSection)
	}
}

func checkoutWorktree(ref string) string {
	tmpDir, err := os.MkdirTemp("", "apidiff-"+ref)
	if err != nil {
		log.Fatal(err)
	}
	run("git", "worktree", "add", "--detach", tmpDir, ref)
	return tmpDir
}

func cleanupWorktree(path string) {
	run("git", "worktree", "remove", "--force", path)
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func snapshotAPI(dir string) map[string]APIPackage {
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

func diffAPI(oldAPI, newAPI map[string]APIPackage) {
	fmt.Println("# API Diff\n")

	for path, newPkg := range newAPI {
		oldPkg, ok := oldAPI[path]
		if !ok {
			fmt.Printf("## Package added: `%s`\n\n", path)
			continue
		}

		// Funcs
		diffList("Funcs", path, oldPkg.Funcs, newPkg.Funcs)
		// Vars
		diffList("Vars", path, oldPkg.Vars, newPkg.Vars)
		// Consts
		diffList("Consts", path, oldPkg.Consts, newPkg.Consts)
		// Types
		for tname, newType := range newPkg.Types {
			oldType, ok := oldPkg.Types[tname]
			if !ok {
				fmt.Printf("- Type added in `%s`: `%s`\n", path, tname)
				continue
			}

			// diffList(fmt.Sprintf("Type `%s` Fields", tname), path, oldType.Fields, newType.Fields)
			diffs.DiffStructFields(path, tname, oldType.Fields, newType.Fields)

			diffList(fmt.Sprintf("Type `%s` Methods", tname), path, oldType.Methods, newType.Methods)
		}
		for tname := range oldPkg.Types {
			if _, ok := newPkg.Types[tname]; !ok {
				fmt.Printf("- Type removed from `%s`: `%s`\n", path, tname)
			}
		}
	}

	for path := range oldAPI {
		if _, ok := newAPI[path]; !ok {
			fmt.Printf("## Package removed: `%s`\n\n", path)
		}
	}
}

func diffList(label, path string, oldList, newList []string) {
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
			fmt.Printf("- Added %s in `%s`: `%s`\n", label, path, x)
		}
	}
	for x := range oldSet {
		if !newSet[x] {
			fmt.Printf("- Removed %s from `%s`: `%s`\n", label, path, x)
		}
	}
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
