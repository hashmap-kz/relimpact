## API Changes

- [Summary](#summary)
- [Breaking Changes](#breaking-changes)
- [Packages Added](#packages-added)
- [Package Changes](#package-changes)

### Summary

| Kind of Change   | Count |
|------------------|-------|
| Packages Added   | 2     |
| Packages Removed | 0     |
| Funcs Added      | 7     |
| Funcs Removed    | 5     |
| Vars Added       | 0     |
| Vars Removed     | 0     |
| Consts Added     | 0     |
| Consts Removed   | 0     |
| Types Added      | 6     |
| Types Removed    | 0     |
| Fields Added     | 0     |
| Fields Removed   | 0     |
| Methods Added    | 0     |
| Methods Removed  | 0     |
| Total Changes    | 20    |

### Breaking Changes

- Funcs Removed: **5**

### Packages Added

- `github.com/hashmap-kz/relimpact/cmd`
- `github.com/hashmap-kz/relimpact/internal/testutils`

### Package Changes

#### Package `github.com/hashmap-kz/relimpact/internal/diffs`

<details>
<summary>Click to expand</summary>

- Added Funcs:
    - DiffAPI(map[string]github.com/hashmap-kz/relimpact/internal/diffs.APIPackage, map[string]
      github.com/hashmap-kz/relimpact/internal/diffs.APIPackage) -> (*
      github.com/hashmap-kz/relimpact/internal/diffs.APIDiff)
    - DiffDocs(string, string) -> ([]github.com/hashmap-kz/relimpact/internal/diffs.DocDiff)
    - DiffGoMod(string, string) -> (github.com/hashmap-kz/relimpact/internal/diffs.GoModDiff)
    - DiffOther(string, string, string, []string) -> (*
      github.com/hashmap-kz/relimpact/internal/diffs.OtherFilesDiffSummary)
    - FormatAllDocDiffs([]github.com/hashmap-kz/relimpact/internal/diffs.DocDiff) -> (string)
- Added Type:
    - APIDiff
    - APIDiffRes
    - DocDiff
    - GoModDiff
    - OtherFileDiff
    - OtherFilesDiffSummary
- Removed Funcs:
    - DiffAPI(map[string]github.com/hashmap-kz/relimpact/internal/diffs.APIPackage, map[string]
      github.com/hashmap-kz/relimpact/internal/diffs.APIPackage)
    - DiffDocs(string, string) -> ([]string)
    - DiffOtherFiles(string, string, []string) -> (string)

</details>

#### Package `github.com/hashmap-kz/relimpact/internal/gitutils`

<details>
<summary>Click to expand</summary>

- Added Funcs:
    - CheckoutWorktree(string, string) -> (string)
    - CleanupWorktree(string, string)
- Removed Funcs:
    - CheckoutWorktree(string) -> (string)
    - CleanupWorktree(string)

</details>

---

## Documentation Changes

### Doc File: **`README.md`**

#### Summary:

- Headings added: 10
- Headings removed: 0
- Links added: 3
- Links removed: 0
- Images added: 0
- Images removed: 0
- Sections changed: 12

#### Headings added:

- Alpine Linux
- Debian
- Docker images are available at quay.io/hashmap_kz/relimpact
- Features
- GitHub Action
- Homebrew installation
- Installation
- Installation script for Unix-Based OS (requires: tar, curl, jq):
- Manual Installation
- Package-Based installation (suitable in CI/CD)

#### Links added:

- ./docs
- https://github.com/hashmap-kz/relimpact/releases
- https://quay.io/repository/hashmap_kz/relimpact

<details>
<summary>Section Word Count Changes (12 changes)</summary>

- Section `Alpine Linux`: ADDED (2 words)
- Section `Debian`: ADDED (1 words)
- Section `Docker images are available at quay.io/hashmap_kz/relimpact`: ADDED (7 words)
- Section `Example Output`: 2 -> 8 words
- Section `Features`: ADDED (130 words)
- Section `GitHub Action`: ADDED (2 words)
- Section `Homebrew installation`: ADDED (2 words)
- Section `Installation script for Unix-Based OS (requires: tar, curl, jq):`: ADDED (10 words)
- Section `Installation`: ADDED (1 words)
- Section `Manual Installation`: ADDED (24 words)
- Section `Package-Based installation (suitable in CI/CD)`: ADDED (5 words)
- Section `relimpact`: 167 -> 34 words

</details>

### Doc File: **`docs\basic\README.md`**

#### Summary:

- Headings added: 20
- Headings removed: 0
- Links added: 4
- Links removed: 0
- Images added: 0
- Images removed: 0
- Sections changed: 20

#### Headings added:

- .json
- .md
- .txt
- .yaml
- API Changes
- Dependencies updated
- Doc File: README.md
- Documentation Changes
- Example Changelog Output
- Headings added:
- Images added:
- Links added:
- Links removed:
- Other Files Changes
- Package Changes
- Package github.com/example/project/pkg/feature
- Packages Added
- Summary
- Summary:
- go.mod Changes

#### Links added:

- https://example.com/old-docs
- https://github.com/example/project/actions
- https://img.shields.io/github/workflow/status/example/project/CI
- https://pkg.go.dev/github.com/example/project

<details>
<summary>Section Word Count Changes (20 changes)</summary>

- Section `.json`: ADDED (5 words)
- Section `.md`: ADDED (4 words)
- Section `.txt`: ADDED (3 words)
- Section `.yaml`: ADDED (4 words)
- Section `API Changes`: ADDED (2 words)
- Section `Dependencies updated`: ADDED (10 words)
- Section `Doc File: README.md`: ADDED (3 words)
- Section `Documentation Changes`: ADDED (2 words)
- Section `Example Changelog Output`: ADDED (3 words)
- Section `Headings added:`: ADDED (7 words)
- Section `Images added:`: ADDED (16 words)
- Section `Links added:`: ADDED (4 words)
- Section `Links removed:`: ADDED (3 words)
- Section `Other Files Changes`: ADDED (3 words)
- Section `Package Changes`: ADDED (2 words)
- Section `Package github.com/example/project/pkg/feature`: ADDED (58 words)
- Section `Packages Added`: ADDED (3 words)
- Section `Summary:`: ADDED (22 words)
- Section `Summary`: ADDED (93 words)
- Section `go.mod Changes`: ADDED (2 words)

</details>

### Doc File: **`docs\projects\loki\README.md`**

#### Summary:

- Headings added: 18
- Headings removed: 0
- Links added: 0
- Links removed: 0
- Images added: 0
- Images removed: 0
- Sections changed: 18

#### Headings added:

- .yaml
- .yml
- API Changes
- Other Files Changes
- Package Changes
- Package github.com/grafana/loki/v3/pkg/canary/writer
- Package github.com/grafana/loki/v3/pkg/dataobj
- Package github.com/grafana/loki/v3/pkg/dataobj/consumer
- Package github.com/grafana/loki/v3/pkg/dataobj/internal/util/bufpool
- Package github.com/grafana/loki/v3/pkg/ingester
- Package github.com/grafana/loki/v3/pkg/kafka
- Package github.com/grafana/loki/v3/pkg/limits
- Package github.com/grafana/loki/v3/pkg/limits/frontend
- Package github.com/grafana/loki/v3/pkg/logql
- Packages Added
- Packages Removed
- Summary
- go.mod Changes

<details>
<summary>Section Word Count Changes (18 changes)</summary>

- Section `.yaml`: ADDED (9 words)
- Section `.yml`: ADDED (7 words)
- Section `API Changes`: ADDED (2 words)
- Section `Other Files Changes`: ADDED (3 words)
- Section `Package Changes`: ADDED (2 words)
- Section `Package github.com/grafana/loki/v3/pkg/canary/writer`: ADDED (66 words)
- Section `Package github.com/grafana/loki/v3/pkg/dataobj/consumer`: ADDED (14 words)
- Section `Package github.com/grafana/loki/v3/pkg/dataobj/internal/util/bufpool`: ADDED (16 words)
- Section `Package github.com/grafana/loki/v3/pkg/dataobj`: ADDED (96 words)
- Section `Package github.com/grafana/loki/v3/pkg/ingester`: ADDED (19 words)
- Section `Package github.com/grafana/loki/v3/pkg/kafka`: ADDED (8 words)
- Section `Package github.com/grafana/loki/v3/pkg/limits/frontend`: ADDED (38 words)
- Section `Package github.com/grafana/loki/v3/pkg/limits`: ADDED (119 words)
- Section `Package github.com/grafana/loki/v3/pkg/logql`: ADDED (13 words)
- Section `Packages Added`: ADDED (7 words)
- Section `Packages Removed`: ADDED (5 words)
- Section `Summary`: ADDED (93 words)
- Section `go.mod Changes`: ADDED (5 words)

</details>

---

## go.mod Changes

### Dependencies added

- github.com/davecgh/go-spew v1.1.1
- github.com/pmezard/go-difflib v1.0.0
- github.com/stretchr/testify v1.10.0
- gopkg.in/yaml.v3 v3.0.1

---

## Other Files Changes

### `.txt`

- Removed:
    - todos.txt

### `.yml`

- Added:
    - .github/workflows/relimpact.yml

- Modified:
    - .github/workflows/release.yml
    - .goreleaser.yml
