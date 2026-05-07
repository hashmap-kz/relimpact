# relimpact

**Fast API compatibility reports for Go projects.**

[![License](https://img.shields.io/github/license/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/hashmap-kz/relimpact)](https://goreportcard.com/report/github.com/hashmap-kz/relimpact)
[![Go Reference](https://pkg.go.dev/badge/github.com/hashmap-kz/relimpact.svg)](https://pkg.go.dev/github.com/hashmap-kz/relimpact)
[![Workflow Status](https://img.shields.io/github/actions/workflow/status/hashmap-kz/relimpact/ci.yml?branch=master)](https://github.com/hashmap-kz/relimpact/actions/workflows/ci.yml?query=branch:master)
[![GitHub Issues](https://img.shields.io/github/issues/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/issues)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/blob/master/go.mod#L3)
[![Latest Release](https://img.shields.io/github/v/release/hashmap-kz/relimpact)](https://github.com/hashmap-kz/relimpact/releases/latest)

`relimpact` compares two Git refs, snapshots the exported Go API, and reports what changed: 
**breaking changes first**, compatible additions below.

It is not a raw diff tool or a changelog generator. It answers one release question:

> Did this change break public Go API?

---

## Preview

**Breaking Changes**

![Breaking Changes](https://raw.githubusercontent.com/hashmap-kz/assets/main/relimpact/01-relimpact-breaking-changes.png)

**New API**

![New API](https://raw.githubusercontent.com/hashmap-kz/assets/main/relimpact/02-relimpact-new-features.png)

**PR comment**

![Markdown Format](https://raw.githubusercontent.com/hashmap-kz/assets/main/relimpact/04-relimpact-markdown.png)

## Install

Using Go:

```bash
go install github.com/hashmap-kz/relimpact@latest
```

Using Homebrew:

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

## GitHub Actions

Generate a Markdown report and post it as a pull request comment:

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
```

To also keep the HTML report as a CI artifact:

```yaml
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

## License

MIT License. See [LICENSE](./LICENSE) for details.
