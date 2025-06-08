# relimpact

**Release Impact Analyzer for Go projects - catch breaking API changes, docs updates & important file diffs - fast.**

[![License](https://img.shields.io/github/license/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/hashmap-kz/relimpact)](https://goreportcard.com/report/github.com/hashmap-kz/relimpact)
[![Go Reference](https://pkg.go.dev/badge/github.com/hashmap-kz/relimpact.svg)](https://pkg.go.dev/github.com/hashmap-kz/relimpact)
[![Workflow Status](https://img.shields.io/github/actions/workflow/status/hashmap-kz/relimpact/ci.yml?branch=master)](https://github.com/hashmap-kz/relimpact/actions/workflows/ci.yml?query=branch:master)
[![GitHub Issues](https://img.shields.io/github/issues/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/issues)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/blob/master/go.mod#L3)
[![Latest Release](https://img.shields.io/github/v/release/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/releases/latest)

---

## Features

- **API Diff** – Track breaking public API changes (structs, interfaces, functions, constants, variables) to prevent
  surprises for your users.
- **Docs Diff** – Section-aware, heading-aware Markdown diff to highlight meaningful content changes, not noisy line
  diffs.
- **Other Files Diff** – Group file changes by extension (.sh, .sql, .json, etc.) to surface important migrations,
  scripts, and auxiliary file updates.
- **Designed for Release PR reviews** – Helps reviewers quickly see the real impact of changes at a glance.
- **Human-friendly Markdown Reports** – Ready to paste into GitHub Releases, Slack, or changelogs.
- **Works in GitHub Actions, GitLab CI, or locally** – Integrates easily into your CI pipelines or local release
  process.
- **No server required** – Pure CLI tool. No services to deploy or manage - works entirely from your Git repo and
  terminal.

---

## Quickstart

### Run on a GitHub PR:

```bash
relimpact --old v1.0.0 --new HEAD --output release-impact.md
```

### Example Output

```markdown
---
## API Diff

- Added Funcs in `pkg/mymodule`: `NewClient(config.Config) -> (*Client, error)`
- Method `DoSomething` in `pkg/mymodule.Client` changed signature: `(ctx context.Context) -> (error)`
- Removed Method `DeprecatedThing` from `pkg/mymodule.Client`: `() -> (string)`

---
## Documentation Changes: `README.md`

### Headings added:
- Advanced Usage

### Section Word Count Changes:
- Section `Quick Start`: 142 -> 155 words
- Section `Deprecated Options`: REMOVED (45 words)
- Section `New Features`: ADDED (67 words)

---
## Other Files Changes

### .sql

- Added:
    - migrations/20240608_add_user_table.sql
- Removed:
    - migrations/20240501_drop_old_index.sql

### .sh

- Modified:
    - scripts/deploy.sh
```

--- 

## GitHub Action

```yaml
name: Release Impact on PR

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  release-impact:
    name: Generate Release Impact Report
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Determine previous tag
        id: prevtag
        run: |
          git fetch --tags
          PREV_TAG=$(git tag --sort=-version:refname | head -n2 | tail -n1)
          echo "prev_tag=$PREV_TAG" >> $GITHUB_OUTPUT

      - uses: hashmap-kz/relimpact-action@v1
        with:
          old_ref: ${{ steps.prevtag.outputs.prev_tag }}
          new_ref: HEAD
          output: release-impact.md

      - name: Upload Release Impact Report
        uses: actions/upload-artifact@v4
        with:
          name: release-impact
          path: release-impact.md
```

---

## Installation

### Docker images are available at [quay.io/hashmap_kz/relimpact](https://quay.io/repository/hashmap_kz/relimpact)

```bash
docker pull quay.io/hashmap_kz/relimpact:latest
```

### Manual Installation

1. Download the latest binary for your platform from
   the [Releases page](https://github.com/hashmap-kz/relimpact/releases).
2. Place the binary in your system's `PATH` (e.g., `/usr/local/bin`).

### Installation script for Unix-Based OS _(requires: tar, curl, jq)_:

```bash
(
set -euo pipefail

OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')"
TAG="$(curl -s https://api.github.com/repos/hashmap-kz/relimpact/releases/latest | jq -r .tag_name)"

curl -L "https://github.com/hashmap-kz/relimpact/releases/download/${TAG}/relimpact_${TAG}_${OS}_${ARCH}.tar.gz" |
tar -xzf - -C /usr/local/bin && \
chmod +x /usr/local/bin/relimpact
)
```

### Homebrew installation

```bash
brew tap hashmap-kz/homebrew-tap
brew install relimpact
```

### Package-Based installation (suitable in CI/CD)

#### Debian

```bash
sudo apt update -y && sudo apt install -y curl
curl -LO https://github.com/hashmap-kz/relimpact/releases/latest/download/relimpact_linux_amd64.deb
sudo dpkg -i relimpact_linux_amd64.deb
```

#### Alpine Linux

```bash
apk update && apk add --no-cache bash curl
curl -LO https://github.com/hashmap-kz/relimpact/releases/latest/download/relimpact_linux_amd64.apk
apk add relimpact_linux_amd64.apk --allow-untrusted
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
