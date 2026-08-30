# API compatibility report

`269945d` -> `5a0c267`

> **Breaking API changes detected.** Review before release.

| Breaking | Changed | Removed | Added | Packages |
|---:|---:|---:|---:|---:|
| 1 | 0 | 1 | 45 | 7 |

## Contents

### Breaking changes

- `github.com/hashmap-kz/relimpact` - 3

### New API

- `github.com/hashmap-kz/relimpact/internal/diffs` - 51
- `github.com/hashmap-kz/relimpact/internal/loggr` - 20
- `github.com/hashmap-kz/relimpact/internal/testutils` - 4
- `github.com/hashmap-kz/relimpact/cmd` - 2
- `github.com/hashmap-kz/relimpact/internal/gitutils` - 2
- `github.com/hashmap-kz/relimpact/internal/version` - 1

## Breaking changes

### `github.com/hashmap-kz/relimpact`

#### Removed API

**type GoPackage struct**

```diff
type GoPackage struct {
- Export     []string
- ImportPath string
- Name       string
}
```

## New API

### `github.com/hashmap-kz/relimpact/internal/diffs`

#### Added API

**Functions**

```diff
+ DiffAPI(map[string]diffs.APIPackage, map[string]diffs.APIPackage) *diffs.APIDiff
+ DiffDocs(string, string) []diffs.DocDiff
+ DiffGoMod(string, string) diffs.GoModDiff
+ DiffOther(string, string, string, []string) *diffs.OtherFilesDiffSummary
+ FormatAllDocDiffs([]diffs.DocDiff) string
+ SnapshotAPI(string) map[string]diffs.APIPackage
```

**type APIDiff struct**

```diff
type APIDiff struct {
+ ConstsAdded     []diffs.APIDiffRes
+ ConstsRemoved   []diffs.APIDiffRes
+ FieldsAdded     []diffs.APIDiffRes
+ FieldsRemoved   []diffs.APIDiffRes
+ FuncsAdded      []diffs.APIDiffRes
+ FuncsRemoved    []diffs.APIDiffRes
+ MethodsAdded    []diffs.APIDiffRes
+ MethodsRemoved  []diffs.APIDiffRes
+ PackagesAdded   []string
+ PackagesRemoved []string
+ TypesAdded      []diffs.APIDiffRes
+ TypesRemoved    []diffs.APIDiffRes
+ VarsAdded       []diffs.APIDiffRes
+ VarsRemoved     []diffs.APIDiffRes
}
```

**type APIDiffRes struct**

```diff
type APIDiffRes struct {
+ Label string
+ Path  string
+ X     string
}
```

**type APIPackage struct**

```diff
type APIPackage struct {
+ Consts []string
+ Funcs  []string
+ Types  map[string]diffs.APIType
+ Vars   []string
}
```

**type APIType struct**

```diff
type APIType struct {
+ Fields  []string
+ Kind    string
+ Methods []string
}
```

**type DocDiff struct**

```diff
type DocDiff struct {
+ File              string
+ HeadingsAdded     []string
+ HeadingsRemoved   []string
+ ImagesAdded       []string
+ ImagesRemoved     []string
+ LinksAdded        []string
+ LinksRemoved      []string
+ SectionWordChange []string
}
```

**type DocInfo struct**

```diff
type DocInfo struct {
+ Headings    []string
+ Images      []string
+ Links       []string
+ SectionWord map[string]int
}
```

**type GoModDiff struct**

```diff
type GoModDiff struct {
+ DependenciesAdded   []string
+ DependenciesRemoved []string
+ DependenciesUpdated []string
}
```

**type OtherFileDiff struct**

```diff
type OtherFileDiff struct {
+ Added    []string
+ Ext      string
+ Modified []string
+ Other    []string
+ Removed  []string
}
```

**type OtherFilesDiffSummary struct**

```diff
type OtherFilesDiffSummary struct {
+ Diffs []diffs.OtherFileDiff
}
```

### `github.com/hashmap-kz/relimpact/internal/loggr`

#### Added API

**Types**

```diff
+ LogLevel
```

**Functions**

```diff
+ Debug(string)
+ Debugf(string, []any)
+ Error(string)
+ Errorf(string, []any)
+ Fatal(string)
+ Fatalf(string, []any)
+ Info(string)
+ Infof(string, []any)
+ Init(loggr.LogLevel, string)
+ Trace(string)
+ Tracef(string, []any)
+ Warn(string)
+ Warnf(string, []any)
```

**Variables**

```diff
+ Logger *loggr.LevelLogger
```

**Constants**

```diff
+ LevelDebug = 1
+ LevelError = 4
+ LevelInfo = 2
+ LevelTrace = 0
+ LevelWarn = 3
```

**type LevelLogger struct**

```diff
type LevelLogger struct {
}
```

### `github.com/hashmap-kz/relimpact/internal/testutils`

#### Added API

**Functions**

```diff
+ ProjectRoot(*testing.T) string
+ ReadTestData(*testing.T, string) []byte
+ RunGit(*testing.T, string, []string)
+ RunGo(*testing.T, string, []string)
```

### `github.com/hashmap-kz/relimpact/cmd`

#### Added API

**Functions**

```diff
+ CreateChangelog(string, string, string) string
+ CreateChangelogSequential(string, string, string) string
```

### `github.com/hashmap-kz/relimpact/internal/gitutils`

#### Added API

**Functions**

```diff
+ CheckoutWorktree(string, string) string
+ CleanupWorktree(string, string)
```

### `github.com/hashmap-kz/relimpact/internal/version`

#### Added API

**Variables**

```diff
+ Version string
```
