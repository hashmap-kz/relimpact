package diffs

import (
	"encoding/json"
	"fmt"
	"go/token"
	"go/types"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashmap-kz/relimpact/internal/loggr"

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

// APITypeBody holds the full definition of a struct or interface type,
// carried through the diff pipeline so the HTML report can render it.
// Nil for all symbol kinds other than whole-type additions and removals.
type APITypeBody struct {
	Kind    string   // "struct", "interface", or the underlying kind string
	Fields  []string // exported fields (structs only)
	Methods []string // exported methods
}

// DiffItem represents a single added or removed public symbol in a package.
type DiffItem struct {
	Label     string
	Path      string
	Signature string       // raw type-checker signature string for this symbol
	TypeBody  *APITypeBody // non-nil only for whole-type adds/removes
}

type APIDiff struct {
	PackagesAdded   []string   `json:"packages_added,omitempty"`
	PackagesRemoved []string   `json:"packages_removed,omitempty"`
	FuncsAdded      []DiffItem `json:"funcs_added,omitempty"`
	FuncsRemoved    []DiffItem `json:"funcs_removed,omitempty"`
	VarsAdded       []DiffItem `json:"vars_added,omitempty"`
	VarsRemoved     []DiffItem `json:"vars_removed,omitempty"`
	ConstsAdded     []DiffItem `json:"consts_added,omitempty"`
	ConstsRemoved   []DiffItem `json:"consts_removed,omitempty"`
	TypesAdded      []DiffItem `json:"types_added,omitempty"`
	TypesRemoved    []DiffItem `json:"types_removed,omitempty"`
	FieldsAdded     []DiffItem `json:"fields_added,omitempty"`
	FieldsRemoved   []DiffItem `json:"fields_removed,omitempty"`
	MethodsAdded    []DiffItem `json:"methods_added,omitempty"`
	MethodsRemoved  []DiffItem `json:"methods_removed,omitempty"`
}

func getCacheDir() string {
	if dir := os.Getenv("RELIMPACT_API_CACHE_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), "relimpact-api-cache") // fallback for local runs
}

func SnapshotAPI(dir string) map[string]APIPackage {
	// TODO: debuglog

	sha := getGitCommitSHA(dir)
	cachePath := filepath.Join(getCacheDir(), sha+".json")
	loggr.Debugf("cache path: %s", cachePath)

	// Try to load from cache
	if data, err := os.ReadFile(cachePath); err == nil {
		var cached map[string]APIPackage
		if json.Unmarshal(data, &cached) == nil {
			loggr.Debugf("cache hit. sha=%s", sha)
			return cached
		}
	}

	loggr.Debugf("cache miss. sha=%s", sha)

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

	// NOTE: this is the most expensive routine in the whole app.

	loadStart := time.Now()

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatal(err)
	}

	loggr.Debugf("packages load. time=%s, sha=%s", time.Since(loadStart).String(), sha)

	modulePath := getModulePath(dir)
	api := make(map[string]APIPackage)

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			loggr.Errorf("error in package: %s", pkg.PkgPath)
			for _, err := range pkg.Errors {
				loggr.Errorf("error details: %v", err)
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

	// TODO: checksum

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
	var b strings.Builder
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
				// whole type added — carry the full definition so the report can render it
				apiDiffResult.TypesAdded = append(apiDiffResult.TypesAdded, DiffItem{
					Label:     "Type",
					Path:      path,
					Signature: tname,
					TypeBody: &APITypeBody{
						Kind:    newType.Kind,
						Fields:  newType.Fields,
						Methods: newType.Methods,
					},
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
		for tname, oldType := range oldPkg.Types {
			if _, ok := newPkg.Types[tname]; !ok {
				// whole type removed — carry the full definition for the report
				apiDiffResult.TypesRemoved = append(apiDiffResult.TypesRemoved, DiffItem{
					Label:     "Type",
					Path:      path,
					Signature: tname,
					TypeBody: &APITypeBody{
						Kind:    oldType.Kind,
						Fields:  oldType.Fields,
						Methods: oldType.Methods,
					},
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

func diffList(label, path string, oldList, newList []string) (added, removed []DiffItem) {
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
			added = append(added, DiffItem{
				Label:     label,
				Path:      path,
				Signature: x,
			})
		}
	}
	for x := range oldSet {
		if !newSet[x] {
			removed = append(removed, DiffItem{
				Label:     label,
				Path:      path,
				Signature: x,
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
