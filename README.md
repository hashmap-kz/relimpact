# relimpact

**Release Impact Analyzer for Go projects - API Diff, Docs Diff & more**

[![License](https://img.shields.io/github/license/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/hashmap-kz/relimpact)](https://goreportcard.com/report/github.com/hashmap-kz/relimpact)
[![Go Reference](https://pkg.go.dev/badge/github.com/hashmap-kz/relimpact.svg)](https://pkg.go.dev/github.com/hashmap-kz/relimpact)
[![Workflow Status](https://img.shields.io/github/actions/workflow/status/hashmap-kz/relimpact/ci.yml?branch=master)](https://github.com/hashmap-kz/relimpact/actions/workflows/ci.yml?query=branch:master)
[![GitHub Issues](https://img.shields.io/github/issues/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/issues)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/blob/master/go.mod#L3)
[![Latest Release](https://img.shields.io/github/v/release/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/releases/latest)

---

- **API Diff** - track public API changes (structs, interfaces, functions, constants, variables) to catch breaking
  changes before they reach your users
- **Docs Diff** - section-aware, heading-aware, word count diff for Markdown docs to highlight meaningful content
  changes, not noisy line diffs
- **Other Files Diff** - group changes by extension (.sh, .sql, .json, etc.) to catch important migrations, scripts, and
  auxiliary file updates
- **Perfect for Release PR reviews** - helps reviewers see the real impact of each change at a glance
- **Human-friendly Release Reports in Markdown** - designed to be copy-pasted into GitHub Releases, Slack, or changelogs
- **Works great in GitHub Actions, GitLab CI, and locally** - easily integrates into your CI pipelines or local release
  process
- **Zero server required** - pure CLI tool - no services to deploy or manage, works entirely from your Git repo and your
  terminal

---

## Quickstart

### Run on a GitHub PR:

```bash
relimpact --old v1.0.0 --new HEAD --output release-impact.md
```

### Example Output

```markdown
## API Diff

- Added Funcs in `pkg/mymodule`: `NewClient(config.Config) -> (*Client, error)`
- Method `DoSomething` in `pkg/mymodule.Client` changed signature: `(ctx context.Context) -> (error)`
- Removed Method `DeprecatedThing` from `pkg/mymodule.Client`: `() -> (string)`

## Documentation Changes: `README.md`

### Headings added:
- Advanced Usage

### Section Word Count Changes:
- Section `Quick Start`: 142 -> 155 words
- Section `Deprecated Options`: REMOVED (45 words)
- Section `New Features`: ADDED (67 words)

## Other Files Changes

### .sql

- Added:
    - migrations/20240608_add_user_table.sql

### .sh

- Modified:
    - scripts/deploy.sh
```

---

## Design Notes

`relimpact` helps you understand **what really changed** between Git refs, in a way that is:

> **Human-friendly** / **Structured** / **Noise-free** / **Release-ready**

### 1. Go Source API Changes

- Tracks changes to your **public exported API**:
    - `struct` fields
    - `interfaces` and their methods
    - `functions`
    - `methods`
    - `constants`
    - `variables`

- Built on top of **Go type system and AST parsing**:
    - Uses `golang.org/x/tools/go/packages` to understand the real API, not just text diffs.
    - **Ignores formatting changes**, reordering, comments -> only tracks semantic API impact.
    - Detects **breaking changes**, such as:
        - method signature changes
        - removed fields
        - removed types
        - changed constants
        - new API elements.

### 2. Markdown Docs Changes

- Tracks changes in **Markdown files**:
    - any `.md` in your repo.

- Uses **Markdown AST parsing**:
    - Based on `goldmark` parser.
    - Understands:
        - Headings (added / removed)
        - Links (added / removed)
        - Images (added / removed)
        - **Section-level word count diffs** -> detect real content changes -> not noisy line diffs.

- Provides a **highly readable report**:
    - No messy raw `git diff` output.
    - Clear "Section X: 142 -> 155 words" style diffs.
    - Great for docs-heavy projects and libraries.

### 3. Other Files Changes

- Tracks other file changes, grouped by extension:
    - `.sh`, `.sql`, `.json`, `.yaml`, `.conf`, `.ini`, `.txt`, etc.

- Built on top of **Git diff**:
    - Uses `git diff --name-status` under the hood.
    - Groups files per extension -> clean, easy to review.


---

## License

MIT License. See [LICENSE](./LICENSE) for details.
