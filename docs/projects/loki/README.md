## API Changes

- [Summary](#summary)
- [Breaking Changes](#breaking-changes)
- [Package Changes](#package-changes)

### Summary

|   Kind of Change   | Count |
|--------------------|-------|
| Packages Added     |     0 |
| Packages Removed   |     0 |
| Funcs Added        |     0 |
| Funcs Removed      |     1 |
| Vars Added         |     0 |
| Vars Removed       |     0 |
| Consts Added       |     0 |
| Consts Removed     |     0 |
| Types Added        |     0 |
| Types Removed      |     2 |
| Fields Added       |     0 |
| Fields Removed     |     2 |
| Methods Added      |     2 |
| Methods Removed    |     3 |
| Total Changes      |    10 |

### Breaking Changes

- Types Removed: **2**
- Fields Removed: **2**
- Methods Removed: **3**
- Funcs Removed: **1**

### Package Changes

#### Package `github.com/grafana/loki/v3/pkg/dataobj/metastore`

<details>
<summary>Click to expand</summary>

- Added Type `Metastore` Methods:
    - StreamIDs(context.Context, time.Time, time.Time, []*github.com/prometheus/prometheus/model/labels.Matcher) -> ([]string, [][]int64, error)
- Removed Type `Metastore` Methods:
    - StreamIDs(context.Context, time.Time, time.Time, []*github.com/prometheus/prometheus/model/labels.Matcher) -> ([]string, [][]int64, []int, error)

</details>

#### Package `github.com/grafana/loki/v3/pkg/engine/planner/logical`

<details>
<summary>Click to expand</summary>

- Removed Funcs:
    - NewShard(uint32, uint32) -> (*github.com/grafana/loki/v3/pkg/engine/planner/logical.ShardInfo)
- Removed Type:
    - ShardInfo
- Removed Type `MakeTable` Fields:
    - Shard github.com/grafana/loki/v3/pkg/engine/planner/logical.Value

</details>

#### Package `github.com/grafana/loki/v3/pkg/engine/planner/physical`

<details>
<summary>Click to expand</summary>

- Added Type `Catalog` Methods:
    - ResolveDataObj(github.com/grafana/loki/v3/pkg/engine/planner/physical.Expression) -> ([]github.com/grafana/loki/v3/pkg/engine/planner/physical.DataObjLocation, [][]int64, error)
- Removed Type:
    - ShardInfo
- Removed Type `Catalog` Methods:
    - ResolveDataObj(github.com/grafana/loki/v3/pkg/engine/planner/physical.Expression) -> ([]github.com/grafana/loki/v3/pkg/engine/planner/physical.DataObjLocation, [][]int64, [][]int, error)
    - ResolveDataObjWithShard(github.com/grafana/loki/v3/pkg/engine/planner/physical.Expression, github.com/grafana/loki/v3/pkg/engine/planner/physical.ShardInfo) -> ([]github.com/grafana/loki/v3/pkg/engine/planner/physical.DataObjLocation, [][]int64, [][]int, error)
- Removed Type `DataObjScan` Fields:
    - Sections []int

</details>


---
## go.mod Changes

_No changes detected._


---
## Other Files Changes

### `.json`

- Modified:
    - pkg/ui/frontend/package-lock.json
