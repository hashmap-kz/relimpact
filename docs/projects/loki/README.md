```
relimpact --old=376124ed6 --new=0db57b035 --greedy > changelog.md

real	1m8.444s
user	18m3.306s
sys	1m29.071s
```

---

## API Changes

### Summary

| Kind of Change | Count |
|----------------|-------|
| Packages Added   |     5 |
| Packages Removed |     3 |
| Funcs Added      |    12 |
| Funcs Removed    |    20 |
| Vars Added       |     0 |
| Vars Removed     |     2 |
| Consts Added     |     3 |
| Consts Removed   |     3 |
| Types Added      |     9 |
| Types Removed    |    42 |
| Fields Added     |     4 |
| Fields Removed   |     1 |
| Methods Added    |     0 |
| Methods Removed  |     0 |

### Packages Added

- `github.com/grafana/loki/v3/pkg/dataobj/consumer/logsobj`
- `github.com/grafana/loki/v3/pkg/dataobj/internal/util/protocodec`
- `github.com/grafana/loki/v3/pkg/dataobj/internal/util/windowing`
- `github.com/grafana/loki/v3/pkg/dataobj/sections/logs`
- `github.com/grafana/loki/v3/pkg/dataobj/sections/streams`

### Packages Removed

- `github.com/grafana/loki/v3/pkg/dataobj/internal/encoding`
- `github.com/grafana/loki/v3/pkg/dataobj/internal/sections/logs`
- `github.com/grafana/loki/v3/pkg/dataobj/internal/sections/streams`

### Package Changes

#### Package `github.com/grafana/loki/v3/pkg/canary/writer`

- Added Consts:
    - DefaultLogBatchSize untyped int
    - DefaultLogBatchSizeMax untyped int
    - DefaultLogBatchTimeout untyped int
- Added Funcs:
    - NewPush(string, string, time.Duration, github.com/prometheus/common/config.HTTPClientConfig, string, string,
      string, string, bool, *crypto/tls.Config, string, string, string, string, string, *
      github.com/grafana/dskit/backoff.Config, int, github.com/go-kit/log.Logger) -> (
      github.com/grafana/loki/v3/pkg/canary/writer.EntryWriter, error)
- Added Type:
    - BatchedPush
- Removed Funcs:
    - NewPush(string, string, time.Duration, github.com/prometheus/common/config.HTTPClientConfig, string, string,
      string, string, bool, *crypto/tls.Config, string, string, string, string, string, *
      github.com/grafana/dskit/backoff.Config, github.com/go-kit/log.Logger) -> (*
      github.com/grafana/loki/v3/pkg/canary/writer.Push, error)

#### Package `github.com/grafana/loki/v3/pkg/dataobj`

- Added Funcs:
    - FromBucket(context.Context, github.com/thanos-io/objstore.BucketReader, string) -> (*
      github.com/grafana/loki/v3/pkg/dataobj.Object, error)
    - FromReaderAt(io.ReaderAt, int64) -> (*github.com/grafana/loki/v3/pkg/dataobj.Object, error)
    - NewBuilder() -> (*github.com/grafana/loki/v3/pkg/dataobj.Builder)
    - NewMetrics() -> (*github.com/grafana/loki/v3/pkg/dataobj.Metrics)
- Added Type:
    - Metrics
    - Section
    - SectionBuilder
    - SectionReader
    - SectionType
    - SectionWriter
    - Sections
- Removed Funcs:
    - FromBucket(github.com/thanos-io/objstore.Bucket, string) -> (*github.com/grafana/loki/v3/pkg/dataobj.Object)
    - FromReaderAt(io.ReaderAt, int64) -> (*github.com/grafana/loki/v3/pkg/dataobj.Object)
    - NewBuilder(github.com/grafana/loki/v3/pkg/dataobj.BuilderConfig) -> (*
      github.com/grafana/loki/v3/pkg/dataobj.Builder, error)
    - NewLogsReader(*github.com/grafana/loki/v3/pkg/dataobj.Object, int) -> (*
      github.com/grafana/loki/v3/pkg/dataobj.LogsReader)
    - NewStreamsReader(*github.com/grafana/loki/v3/pkg/dataobj.Object, int) -> (*
      github.com/grafana/loki/v3/pkg/dataobj.StreamsReader)
    - OrderPredicates([]github.com/grafana/loki/v3/pkg/dataobj/internal/dataset.Predicate) -> ([]
      github.com/grafana/loki/v3/pkg/dataobj/internal/dataset.Predicate)
- Removed Type:
    - AndPredicate
    - BuilderConfig
    - FlushStats
    - LabelFilterPredicate
    - LabelMatcherPredicate
    - LogMessageFilterPredicate
    - LogsPredicate
    - LogsReader
    - Metadata
    - MetadataFilterPredicate
    - MetadataMatcherPredicate
    - NotPredicate
    - OrPredicate
    - Predicate
    - Record
    - SelectivityScore
    - Stream
    - StreamsPredicate
    - StreamsReader
    - TimeRangePredicate
- Removed Vars:
    - ErrBuilderEmpty error
    - ErrBuilderFull error

#### Package `github.com/grafana/loki/v3/pkg/dataobj/consumer`

- Added Type `Config` Fields:
    - BuilderConfig github.com/grafana/loki/v3/pkg/dataobj/consumer/logsobj.BuilderConfig
- Removed Type `Config` Fields:
    - BuilderConfig github.com/grafana/loki/v3/pkg/dataobj.BuilderConfig

#### Package `github.com/grafana/loki/v3/pkg/dataobj/internal/util/bufpool`

- Added Funcs:
    - GetReader(io.Reader) -> (*bufio.Reader)
    - GetUnsized() -> (*bytes.Buffer)
    - PutReader(*bufio.Reader)
    - PutUnsized(*bytes.Buffer)

#### Package `github.com/grafana/loki/v3/pkg/ingester`

- Added Funcs:
    - NewKafkaConsumerFactory(github.com/grafana/loki/v3/pkg/logproto.PusherServer,
      github.com/prometheus/client_golang/prometheus.Registerer, int) -> (
      github.com/grafana/loki/v3/pkg/kafka/partition.ConsumerFactory)
- Removed Funcs:
    - NewKafkaConsumerFactory(github.com/grafana/loki/v3/pkg/logproto.PusherServer,
      github.com/prometheus/client_golang/prometheus.Registerer) -> (
      github.com/grafana/loki/v3/pkg/kafka/partition.ConsumerFactory)

#### Package `github.com/grafana/loki/v3/pkg/kafka`

- Added Type `Config` Fields:
    - MaxConsumerWorkers int

#### Package `github.com/grafana/loki/v3/pkg/limits`

- Added Funcs:
    - New(github.com/grafana/loki/v3/pkg/limits.Config, github.com/grafana/loki/v3/pkg/limits.Limits,
      github.com/go-kit/log.Logger, github.com/prometheus/client_golang/prometheus.Registerer) -> (*
      github.com/grafana/loki/v3/pkg/limits.Service, error)
- Added Type:
    - Service
- Added Type `Config` Fields:
    - BucketDuration time.Duration
    - WindowSize time.Duration
- Removed Consts:
    - PartitionPending github.com/grafana/loki/v3/pkg/limits.PartitionState
    - PartitionReady github.com/grafana/loki/v3/pkg/limits.PartitionState
    - PartitionReplaying github.com/grafana/loki/v3/pkg/limits.PartitionState
- Removed Funcs:
    - NewEvictor(context.Context, time.Duration, github.com/grafana/loki/v3/pkg/limits.Evictable,
      github.com/go-kit/log.Logger) -> (*github.com/grafana/loki/v3/pkg/limits.Evictor, error)
    - NewIngestLimits(github.com/grafana/loki/v3/pkg/limits.Config, github.com/grafana/loki/v3/pkg/limits.Limits,
      github.com/go-kit/log.Logger, github.com/prometheus/client_golang/prometheus.Registerer) -> (*
      github.com/grafana/loki/v3/pkg/limits.IngestLimits, error)
    - NewOffsetReadinessCheck(*github.com/grafana/loki/v3/pkg/limits.PartitionManager) -> (
      github.com/grafana/loki/v3/pkg/limits.PartitionReadinessCheck)
    - NewPartitionLifecycler(github.com/grafana/loki/v3/pkg/limits.Config, *
      github.com/grafana/loki/v3/pkg/limits.PartitionManager,
      github.com/grafana/loki/v3/pkg/kafka/partition.OffsetManager, *github.com/grafana/loki/v3/pkg/limits.UsageStore,
      github.com/go-kit/log.Logger) -> (*github.com/grafana/loki/v3/pkg/limits.PartitionLifecycler)
    - NewPartitionManager() -> (*github.com/grafana/loki/v3/pkg/limits.PartitionManager)
    - NewPlaybackManager(github.com/grafana/loki/v3/pkg/limits.Consumer, *
      github.com/grafana/loki/v3/pkg/limits.PartitionManager, *github.com/grafana/loki/v3/pkg/limits.UsageStore,
      github.com/grafana/loki/v3/pkg/limits.PartitionReadinessCheck, string, github.com/go-kit/log.Logger,
      github.com/prometheus/client_golang/prometheus.Registerer) -> (*
      github.com/grafana/loki/v3/pkg/limits.PlaybackManager)
    - NewRateLimitsAdapter(github.com/grafana/loki/v3/pkg/limits.Limits) -> (*
      github.com/grafana/loki/v3/pkg/limits.RateLimitsAdapter)
    - NewSender(github.com/grafana/loki/v3/pkg/limits.Producer, string, int, string, github.com/go-kit/log.Logger,
      github.com/prometheus/client_golang/prometheus.Registerer) -> (*github.com/grafana/loki/v3/pkg/limits.Sender)
    - NewUsageStore(github.com/grafana/loki/v3/pkg/limits.Config) -> (*github.com/grafana/loki/v3/pkg/limits.UsageStore)
- Removed Type:
    - CondFunc
    - Consumer
    - Evictable
    - Evictor
    - IngestLimits
    - IterateFunc
    - PartitionLifecycler
    - PartitionManager
    - PartitionReadinessCheck
    - PartitionState
    - PlaybackManager
    - Producer
    - RateBucket
    - RateLimitsAdapter
    - Sender
    - Stream
    - UsageStore

#### Package `github.com/grafana/loki/v3/pkg/limits/frontend`

- Removed Funcs:
    - NewNopCache() -> (*github.com/grafana/loki/v3/pkg/limits/frontend.NopCache[K, V])
    - NewRingGatherer(github.com/grafana/dskit/ring.ReadRing, *github.com/grafana/dskit/ring/client.Pool, int,
      github.com/grafana/loki/v3/pkg/limits/frontend.Cache[string, *github.com/grafana/loki/v3/pkg/limits/proto.GetAssignedPartitionsResponse],
      github.com/go-kit/log.Logger) -> (*github.com/grafana/loki/v3/pkg/limits/frontend.RingGatherer)
    - NewTTLCache(time.Duration) -> (*github.com/grafana/loki/v3/pkg/limits/frontend.TTLCache[K, V])
- Removed Type:
    - Cache
    - ExceedsLimitsGatherer
    - NopCache
    - RingGatherer
    - TTLCache

#### Package `github.com/grafana/loki/v3/pkg/logql`

- Added Funcs:
    - NewCountMinSketchEvalStepEvaluator(context.Context, github.com/grafana/loki/v3/pkg/logql.SampleEvaluatorFactory, *
      github.com/grafana/loki/v3/pkg/logql.CountMinSketchEvalExpr, github.com/grafana/loki/v3/pkg/logql.Params) -> (*
      github.com/grafana/loki/v3/pkg/logql.countMinSketchEvalStepEvaluator, error)

---

## go.mod Changes

_No changes detected._



---

## Other Files Changes

### `.yaml`

- Added:
    - relyance.yaml

- Modified:
    - .github/workflows/helm-release.yaml
    - .github/workflows/helm-tagged-release-pr.yaml
    - .github/workflows/helm-weekly-release-pr.yaml
    - tools/dev/kafka/loki-local-config.debug.yaml
    - tools/stream-generator/loki-local-config.debug.yaml

### `.yml`

- Added:
    - .github/workflows/relyance.yml

- Modified:
    - .github/workflows/dependabot_reviewer.yml
    - .github/workflows/logql-analyzer.yml



