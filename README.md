# relimpact

**Fast API compatibility reports for Go projects.**

[![License](https://img.shields.io/github/license/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/hashmap-kz/relimpact)](https://goreportcard.com/report/github.com/hashmap-kz/relimpact)
[![Go Reference](https://pkg.go.dev/badge/github.com/hashmap-kz/relimpact.svg)](https://pkg.go.dev/github.com/hashmap-kz/relimpact)
[![Workflow Status](https://img.shields.io/github/actions/workflow/status/hashmap-kz/relimpact/ci.yml?branch=master)](https://github.com/hashmap-kz/relimpact/actions/workflows/ci.yml?query=branch:master)
[![GitHub Issues](https://img.shields.io/github/issues/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/issues)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/blob/master/go.mod#L3)
[![Latest Release](https://img.shields.io/github/v/release/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/releases/latest)

`relimpact` compares two Git refs, snapshots the exported Go API, and shows what changed: breaking changes first,
compatible additions below.

It answers one release question:

> Did this change break public Go API?

---

## Report

**Breaking Changes**

![Breaking Changes](https://raw.githubusercontent.com/hashmap-kz/assets/main/relimpact/01-relimpact-breaking-changes.png)

**New Features**

![New Features](https://raw.githubusercontent.com/hashmap-kz/assets/main/relimpact/02-relimpact-new-features.png)

**PR report**

![Markdown Format](https://raw.githubusercontent.com/hashmap-kz/assets/main/relimpact/04-relimpact-markdown.png)

## Why

Go API changes are easy to miss in a normal diff.

A renamed field, removed method, or changed return type can silently break users. 
`relimpact` turns those changes into a release-friendly report.

It is intentionally focused on **exported Go API only**

## What it detects

`relimpact` reports changes to exported:

- packages
- types
- functions
- methods
- struct fields
- interface methods
- constants
- variables

The report separates:

- **Breaking changes** - changed or removed API.
- **New API** - compatible additions.

## Installation

Using Go:

```bash
go install github.com/hashmap-kz/relimpact@latest
```

Brew:

```bash
brew tap hashmap-kz/homebrew-tap
brew install relimpact
```

Or download a binary from the [Releases page](https://github.com/hashmap-kz/relimpact/releases).

## Usage

Markdown output is the default:

```bash
relimpact --old=v1.0.0 --new=HEAD > api-report.md
```

HTML output:

```bash
relimpact --old=v1.0.0 --new=HEAD --format=html > api-report.html
```

Run against another repository:

```bash
relimpact --dir=/path/to/repo --old=v1.0.0 --new=HEAD
```

Use maximum concurrency:

```bash
relimpact --old=v1.0.0 --new=HEAD --greedy
```

## Output

Markdown is optimized for PR comments and release notes.

HTML is optimized for CI artifacts and release review. It includes:

- compatibility verdict
- summary cards
- breaking changes first
- new API below
- package navigation
- clean signature diffs
- compact groups for added/removed symbols
- readable struct and interface diffs

## Example

```diff
## Breaking changes

### github.com/acme/project/config

#### Changed signatures

- func Load(path string) *Config
+ func Load(path string) (*Config, error)

#### Removed API

Functions

- MustLoad(string) *Config

Struct fields

type Config struct {
-   LegacyMode bool
}

## New API

### github.com/acme/project/config

#### Added API

Functions

+ FromEnv(string) (*Config, error)

Struct fields

type Config struct {
+   Retention RetentionConfig
}
```

## Cache

`relimpact` can cache API snapshots between CI runs.

```bash
RELIMPACT_API_CACHE_DIR=.cache/relimpact-api-cache \
relimpact --old=v1.0.0 --new=HEAD --format=html > api-report.html
```

## GitHub Actions

```yaml
name: API compatibility

on:
  pull_request:
    branches: [ master ]

permissions:
  contents: read
  pull-requests: write

jobs:
  relimpact:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"

      - name: Install relimpact
        run: |
          go install github.com/hashmap-kz/relimpact@latest

      - name: Generate Markdown API report
        run: |
          relimpact \
            --old="${{ github.event.pull_request.base.sha }}" \
            --new="${{ github.event.pull_request.head.sha }}" \
            --format=markdown \
            > api-report.md

      - name: Comment API report on PR
        uses: marocchino/sticky-pull-request-comment@v2
        with:
          header: relimpact-api-report
          recreate: true
          path: api-report.md

      - name: Upload HTML API report
        run: |
          relimpact \
            --old="${{ github.event.pull_request.base.sha }}" \
            --new="${{ github.event.pull_request.head.sha }}" \
            --format=html \
            > api-report.html

      - name: Upload HTML artifact
        uses: actions/upload-artifact@v4
        with:
          name: api-report
          path: api-report.html
```

## Philosophy

`relimpact` is not a changelog generator.

It is a compatibility report for exported Go API.

If it changed what your users can import, call, implement, or compile against, it belongs in the report. 
If it did not, it stays out.

## License

MIT License. See [LICENSE](./LICENSE) for details.
