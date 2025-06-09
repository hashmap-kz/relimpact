## API Changes

- [Summary](#summary)
- [Breaking Changes](#breaking-changes)
- [Packages Added](#packages-added)
- [Package Changes](#package-changes)

### Summary

|   Kind of Change   | Count |
|--------------------|-------|
| Packages Added     |     2 |
| Packages Removed   |     0 |
| Funcs Added        |     7 |
| Funcs Removed      |     5 |
| Vars Added         |     0 |
| Vars Removed       |     0 |
| Consts Added       |     0 |
| Consts Removed     |     0 |
| Types Added        |     6 |
| Types Removed      |     0 |
| Fields Added       |     0 |
| Fields Removed     |     0 |
| Methods Added      |     0 |
| Methods Removed    |     0 |
| Total Changes      |    20 |

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
  - DiffAPI(map[string]github.com/hashmap-kz/relimpact/internal/diffs.APIPackage, map[string]github.com/hashmap-kz/relimpact/internal/diffs.APIPackage) -> (*github.com/hashmap-kz/relimpact/internal/diffs.APIDiff)
  - DiffDocs(string, string) -> ([]github.com/hashmap-kz/relimpact/internal/diffs.DocDiff)
  - DiffGoMod(string, string) -> (github.com/hashmap-kz/relimpact/internal/diffs.GoModDiff)
  - DiffOther(string, string, string, []string) -> (*github.com/hashmap-kz/relimpact/internal/diffs.OtherFilesDiffSummary)
  - FormatAllDocDiffs([]github.com/hashmap-kz/relimpact/internal/diffs.DocDiff) -> (string)
- Added Type:
  - APIDiff
  - APIDiffRes
  - DocDiff
  - GoModDiff
  - OtherFileDiff
  - OtherFilesDiffSummary
- Removed Funcs:
  - DiffAPI(map[string]github.com/hashmap-kz/relimpact/internal/diffs.APIPackage, map[string]github.com/hashmap-kz/relimpact/internal/diffs.APIPackage)
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

<details>
<summary>Click to expand</summary>

#### Summary:
- Headings added: 10
- Headings removed: 0
- Links added: 4
- Links removed: 0
- Images added: 2
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
- https://codecov.io/gh/hashmap-kz/relimpact
- https://github.com/hashmap-kz/relimpact/releases
- https://quay.io/repository/hashmap_kz/relimpact

#### Images added:
- https://codecov.io/gh/hashmap-kz/relimpact/branch/master/graph/badge.svg
- https://github.com/hashmap-kz/assets/blob/main/relimpact/examples/basic-changelog.png

<details>
<summary>Section Word Count Changes (12 changes)</summary>

- Section `Alpine Linux`: ADDED (2 words)
- Section `Debian`: ADDED (1 words)
- Section `Docker images are available at quay.io/hashmap_kz/relimpact`: ADDED (7 words)
- Section `Example Output`: 2 -> 10 words
- Section `Features`: ADDED (130 words)
- Section `GitHub Action`: ADDED (2 words)
- Section `Homebrew installation`: ADDED (2 words)
- Section `Installation script for Unix-Based OS (requires: tar, curl, jq):`: ADDED (10 words)
- Section `Installation`: ADDED (1 words)
- Section `Manual Installation`: ADDED (24 words)
- Section `Package-Based installation (suitable in CI/CD)`: ADDED (5 words)
- Section `relimpact`: 167 -> 35 words

</details>

</details>


### Doc File: **`docs\basic\README.md`**

<details>
<summary>Click to expand</summary>

#### Summary:
- Headings added: 19
- Headings removed: 0
- Links added: 4
- Links removed: 0
- Images added: 0
- Images removed: 0
- Sections changed: 19

#### Headings added:
- .txt
- .yml
- API Changes
- Breaking Changes
- Dependencies added
- Doc File: README.md
- Doc File: docs\basic\README.md
- Doc File: docs\projects\loki\README.md
- Documentation Changes
- Headings added:
- Links added:
- Other Files Changes
- Package Changes
- Package github.com/hashmap-kz/relimpact/internal/diffs
- Package github.com/hashmap-kz/relimpact/internal/gitutils
- Packages Added
- Summary
- Summary:
- go.mod Changes

#### Links added:
- #breaking-changes
- #package-changes
- #packages-added
- #summary

<details>
<summary>Section Word Count Changes (19 changes)</summary>

- Section `.txt`: ADDED (3 words)
- Section `.yml`: ADDED (6 words)
- Section `API Changes`: ADDED (9 words)
- Section `Breaking Changes`: ADDED (5 words)
- Section `Dependencies added`: ADDED (10 words)
- Section `Doc File: README.md`: ADDED (3 words)
- Section `Doc File: docs\basic\README.md`: ADDED (3 words)
- Section `Doc File: docs\projects\loki\README.md`: ADDED (3 words)
- Section `Documentation Changes`: ADDED (2 words)
- Section `Headings added:`: ADDED (233 words)
- Section `Links added:`: ADDED (245 words)
- Section `Other Files Changes`: ADDED (3 words)
- Section `Package Changes`: ADDED (2 words)
- Section `Package github.com/hashmap-kz/relimpact/internal/diffs`: ADDED (60 words)
- Section `Package github.com/hashmap-kz/relimpact/internal/gitutils`: ADDED (16 words)
- Section `Packages Added`: ADDED (4 words)
- Section `Summary:`: ADDED (66 words)
- Section `Summary`: ADDED (99 words)
- Section `go.mod Changes`: ADDED (2 words)

</details>

</details>


### Doc File: **`docs\projects\loki\README.md`**

<details>
<summary>Click to expand</summary>

#### Summary:
- Headings added: 10
- Headings removed: 0
- Links added: 3
- Links removed: 0
- Images added: 0
- Images removed: 0
- Sections changed: 10

#### Headings added:
- .json
- API Changes
- Breaking Changes
- Other Files Changes
- Package Changes
- Package github.com/grafana/loki/v3/pkg/dataobj/metastore
- Package github.com/grafana/loki/v3/pkg/engine/planner/logical
- Package github.com/grafana/loki/v3/pkg/engine/planner/physical
- Summary
- go.mod Changes

#### Links added:
- #breaking-changes
- #package-changes
- #summary

<details>
<summary>Section Word Count Changes (10 changes)</summary>

- Section `.json`: ADDED (3 words)
- Section `API Changes`: ADDED (7 words)
- Section `Breaking Changes`: ADDED (14 words)
- Section `Other Files Changes`: ADDED (3 words)
- Section `Package Changes`: ADDED (2 words)
- Section `Package github.com/grafana/loki/v3/pkg/dataobj/metastore`: ADDED (31 words)
- Section `Package github.com/grafana/loki/v3/pkg/engine/planner/logical`: ADDED (18 words)
- Section `Package github.com/grafana/loki/v3/pkg/engine/planner/physical`: ADDED (41 words)
- Section `Summary`: ADDED (99 words)
- Section `go.mod Changes`: ADDED (5 words)

</details>

</details>



---
## go.mod Changes

<details>
<summary>Click to expand</summary>

### Dependencies added
- github.com/davecgh/go-spew v1.1.1
- github.com/pmezard/go-difflib v1.0.0
- github.com/stretchr/testify v1.10.0
- gopkg.in/yaml.v3 v3.0.1

</details>



---
## Other Files Changes

### `.txt`

<details>
<summary>Click to expand</summary>

- Added:
  - hack/tostr/api.txt
  - hack/tostr/docs.txt
  - hack/tostr/mods.txt
  - hack/tostr/oth.txt

- Removed:
  - todos.txt

</details>

### `.yml`

<details>
<summary>Click to expand</summary>

- Added:
  - .github/workflows/relimpact.yml

- Modified:
  - .github/workflows/ci.yml
  - .github/workflows/release.yml
  - .goreleaser.yml

</details>



