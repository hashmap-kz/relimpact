## Example Changelog Output

---

## API Changes

### Summary

| Kind of Change   | Count |
|------------------|-------|
| Packages Added   | 1     |
| Packages Removed | 0     |
| Funcs Added      | 3     |
| Funcs Removed    | 1     |
| Vars Added       | 1     |
| Vars Removed     | 0     |
| Consts Added     | 2     |
| Consts Removed   | 1     |
| Types Added      | 2     |
| Types Removed    | 1     |
| Fields Added     | 2     |
| Fields Removed   | 1     |
| Methods Added    | 1     |
| Methods Removed  | 0     |

### Packages Added

* `github.com/example/project/pkg/newfeature`

### Package Changes

#### Package `github.com/example/project/pkg/feature`

* Added Funcs:

    * NewFeature() -> (\*Feature, error)
    * Feature.DoSomething(int) -> error
* Removed Funcs:

    * Feature.OldMethod(string) -> bool
* Added Vars:

    * DefaultTimeout time.Duration
* Added Consts:

    * MaxRetries int
    * DefaultUserAgent string
* Removed Consts:

    * LegacyModeEnabled bool
* Added Types:

    * FeatureConfig
    * FeatureState
* Removed Types:

    * LegacyFeature
* Added Fields:

    * Type `FeatureConfig` Fields:

        * MaxRetries int
        * EnableCache bool
* Removed Fields:

    * Type `LegacyFeature` Fields:

        * DeprecatedField string
* Added Methods:

    * FeatureState.Reset() -> void

---

## Documentation Changes

### Doc File: `README.md`

#### Summary:

* Headings added: 3
* Headings removed: 0
* Links added: 2
* Links removed: 1
* Images added: 1
* Images removed: 0
* Sections changed: 2

#### Headings added:

* Getting Started
* API Reference
* Contributing

#### Links added:

* [https://pkg.go.dev/github.com/example/project](https://pkg.go.dev/github.com/example/project)
* [https://github.com/example/project/actions](https://github.com/example/project/actions)

#### Links removed:

* [https://example.com/old-docs](https://example.com/old-docs)

#### Images added:

* [https://img.shields.io/github/workflow/status/example/project/CI](https://img.shields.io/github/workflow/status/example/project/CI)

<details>
<summary>Section Word Count Changes (2 changes)</summary>

* Section `Introduction`: 123 -> 150 words
* Section `Installation`: ADDED (75 words)

</details>

---

## go.mod Changes

### Dependencies updated

* github.com/foo/bar v1.2.3 -> v1.3.0
* github.com/example/lib v0.9.5 -> v1.0.0

---

## Other Files Changes

### `.yaml`

* Modified:

    * .github/workflows/ci.yaml
    * config/app-config.yaml

### `.md`

* Modified:

    * CONTRIBUTING.md
    * CHANGELOG.md

### `.json`

* Added:

    * config/default-settings.json

* Modified:

    * package-lock.json

### `.txt`

* Removed:

    * docs/obsolete.txt
