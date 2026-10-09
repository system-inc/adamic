# Native/Node oracle split: incomplete measurement

Base: `7a10c877667582a15326acb10fa22fa7a0c45fb8` (`origin/main`).
Reference inspected: `stage1-split/markdownblocks`, `ad97e448`.
No split is implemented. No fixture or compiler changes are made.
Shard count: not chosen. Slowest shard: not measured. Union proof: not run.

## Machine and deadlines

Linux amd64, cgroup `cpu.max = 400000 100000` (four CPU equivalents),
`GOMAXPROCS=4`, Go 1.27.1, clang 20.1.8, Node v24.19.0.
Tools loaded from `/workspace/adamic-tools/env.sh`.
`ADAMIC_GATE_UNCACHED=1`, `TMPDIR=/workspace/scratch`.
A fresh `XDG_CACHE_HOME=/workspace/oracle-cold-setup-cache` was used;
this also makes Go's default build cache cold. Provisioning the pinned submodules
completed before timing. Neither invocation ran concurrently with provisioning.

Every invocation was bounded by `timeout --signal=KILL 75s`.
The test invocation also supplied `-timeout=75s` and `-parallel=4`.
The external deadline includes compilation, which Go's test timeout excludes.

| Invocation | Wall seconds | Cooked | Observation |
| --- | ---: | --- | --- |
| Cold shared-setup test, including Go compilation | 75.006 | yes | Killed before entering the test; exit 137 |
| Prepare oracle test binary, reusing completed Go cache entries | 75.006 | yes | Killed without producing the binary; exit 137 |

These are over-budget measurements, not oracle failures. No shard was run.
The supplied 191.5-second original-test time was not independently reproduced.
There is no after-split measurement and no claimed speedup.

## Measurement probe and commands

A temporary test was added and then removed before this commit:

```go
func TestNativeShardColdSetupMeasurement(t *testing.T) {
    started := time.Now()
    identity(t)
    t.Logf("cold shared identity setup: %.3fs", time.Since(started).Seconds())
}
```

```sh
source /workspace/adamic-tools/env.sh
export GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 TMPDIR=/workspace/scratch
export XDG_CACHE_HOME=/workspace/oracle-cold-setup-cache
{ time timeout --signal=KILL 75s go test ./internal/oracle   -run '^TestNativeShardColdSetupMeasurement$' -count=1 -parallel=4   -timeout=75s -v > /workspace/oracle-setup.log; } 2> /workspace/oracle-setup.time
{ time timeout --signal=KILL 75s go test -c   -o /workspace/oracle-measure.test ./internal/oracle   > /workspace/oracle-compile.log; } 2> /workspace/oracle-compile.time
```

Both stdout files were empty. Shell timing evidence is included alongside this file.
No test assertions executed and no PASS evidence was obtained.

## Findings and remaining work

`TestNativeAgreesWithNode` has no explicit parent build setup. However,
`identity` in `cache_test.go` uses `sync.Once` to prepare Node identity and
three native runtime variants (sanitized, release, counted). The uncached
paths still call it to record misses. Each isolated shard must pay this setup;
its runtime duration remains unmeasured. Native runtime libraries use their
existing disk cache; `ADAMIC_GATE_UNCACHED` currently bypasses oracle result
caching, not that runtime cache. Preserve that distinction.

Cold Go compilation alone exceeded the 75-second deadline on this attempt.
Splitting fixture membership cannot reduce this shared compilation work.
This report stops without selecting an unmeasured shard count or changing the
compiler. A continuation needs to account explicitly for how the gate provisions
Go binaries/build caches, then measure runtime-cold setup and fixture shards
under the same hard deadline. Do not treat prepared-binary timings as fully
cold command timings.

The stage 1 reference has fixed `shard-NNN` children, exact union validation,
and planted disagreement checks. Adamic's `cmd/adamic-gate` currently enumerates
this oracle's children directly from the registered fixture AST; a split needs
to update that enumeration too. Existing coverage validation rejects parent-only
evidence for a missing child, but new oracle shard coverage and planted-failure
proofs have not been implemented or verified.
