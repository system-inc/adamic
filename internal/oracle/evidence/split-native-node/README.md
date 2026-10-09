# Native/Node oracle split measurements

Base: `7a10c877667582a15326acb10fa22fa7a0c45fb8` (`origin/main`).
Reference: `stage1-split/markdownblocks`, `ad97e448`.

The oracle now exposes **34 fixed parallel gate children**. The slowest isolated
cold-runtime measurement is `shard-019-1`, **43.454 seconds**, including 15.812 seconds
of shared setup. All 34 passed below 45 seconds; none hit the hard 90-second deadline.

**Measurement scope:** these per-shard figures start a prepared Go test binary with
fresh native-runtime and oracle-result cache directories. They include Node identity,
all native runtime builds, fixture lowering, sanitized/release builds, execution and
leak checks. They exclude provisioning the Go binary and pinned dependencies.
Fully cold Go compilation remains over the 60-second budget: the historical cold
invocation was killed at 75.006 seconds before entering a test. No fully cold Go-command
compliance is claimed. A separate actual `go test` invocation of the slowest shard,
using the prepared Go build cache and fresh runtime cache, passed in **41.254 seconds**.

No fixture, compiler, or result-cache implementation changed. `ADAMIC_GATE_UNCACHED=1`
still bypasses oracle results; every shard recorded native/node/probe cache hits of zero.
The native runtime library cache retains its existing semantics. Selecting the old
parent still runs every registered fixture through all shards.

## Deterministic partition and selection

Sort the complete registered fixture list by path, distribute by sorted ordinal modulo
32, then subdivide only measured heavy buckets 003 and 019 into two alternating parts.
The count and subdivision choices are checked-in constants. Names and fixture ownership
of other buckets stay unchanged when a heavy bucket is subdivided.

```sh
GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/^shard-019-1$' -count=1 -parallel=4 -timeout=90s
```

`cmd/adamic-gate` discovers the explicit literal child-name list. Its new timing weights
come from these isolated wall measurements, since Go reports zero elapsed for the shard
callbacks while their parallel fixture descendants perform the work. Gate merge still
requires a terminal event for every planned child. A parent PASS cannot satisfy a missing child.

## Before and adjustments

The supplied original 191.5-second time was not reproduced to completion. Under the new
rule, the original oracle was killed at **90.002 seconds**, exit 137: a deadline failure.
It used a prepared binary and a fresh runtime cache.

| Provisional unit | Wall seconds | Outcome | Replacement |
| --- | ---: | --- | --- |
| 16-shard `000` | 63.639 | over the 60-second budget | 32-bucket `000` and `016` |
| 16-shard `001` | 62.934 | over the 60-second budget | 32-bucket `001` and `017` |
| 16-shard `002` | 79.994 | over the 60-second budget | 32-bucket `002` and `018` |
| 32-shard `003` | 53.204 | under 60, above the 45-second target | `003-0`, `003-1` |
| 33-shard `019` | 49.048 | under 60, above the 45-second target | `019-0`, `019-1` |

Only changed memberships were rerun after individual subdivisions. Passing measurements
were retained. The aborted 16-shard `003` run was stopped when replacing that grid;
its incomplete log is retained but supplies no PASS evidence or final timing.

## Final cold-runtime measurements

Linux amd64, cgroup `cpu.max = 400000 100000` (four CPU equivalents), `GOMAXPROCS=4`,
`-test.parallel=4`, Go 1.27.1, clang 20.1.8, Node v24.19.0.
Each selected unit had a fresh `XDG_CACHE_HOME` outside the repository and
`ADAMIC_GATE_UNCACHED=1`; scratch was `/workspace/scratch`. Runs were sequential.
Every test had a 90-second Go timeout plus an external process-group kill at 90 seconds.
A killed unit is a failure; a completed unit at or above 60 seconds fails the gate budget.
Shared setup ranged from **13.179 to 18.605 seconds**, below the target by itself.
A short gate discovery proof ran during the `001` measurement; its possible contention
is included in that observed time. No test was allowed to finish after its deadline.

| Shard | Wall seconds | Shared setup seconds | Fixtures | Killed at 90s | Result |
| --- | ---: | ---: | ---: | --- | --- |
| `shard-000` | 40.116 | 16.167 | 30 | no | PASS |
| `shard-001` | 41.013 | 15.316 | 30 | no | PASS |
| `shard-002` | 40.209 | 13.413 | 30 | no | PASS |
| `shard-003-0` | 30.582 | 15.770 | 15 | no | PASS |
| `shard-003-1` | 35.233 | 16.178 | 15 | no | PASS |
| `shard-004` | 41.056 | 16.120 | 30 | no | PASS |
| `shard-005` | 36.934 | 15.928 | 29 | no | PASS |
| `shard-006` | 37.757 | 16.117 | 29 | no | PASS |
| `shard-007` | 29.376 | 13.303 | 29 | no | PASS |
| `shard-008` | 33.086 | 13.179 | 29 | no | PASS |
| `shard-009` | 35.680 | 13.281 | 29 | no | PASS |
| `shard-010` | 39.262 | 18.088 | 29 | no | PASS |
| `shard-011` | 40.014 | 15.868 | 29 | no | PASS |
| `shard-012` | 38.409 | 15.140 | 29 | no | PASS |
| `shard-013` | 35.483 | 16.863 | 29 | no | PASS |
| `shard-014` | 34.834 | 15.162 | 29 | no | PASS |
| `shard-015` | 28.731 | 18.605 | 29 | no | PASS |
| `shard-016` | 35.574 | 17.715 | 29 | no | PASS |
| `shard-017` | 39.562 | 15.529 | 29 | no | PASS |
| `shard-018` | 34.608 | 16.081 | 29 | no | PASS |
| `shard-019-0` | 22.032 | 16.686 | 15 | no | PASS |
| `shard-019-1` | 43.454 | 15.812 | 14 | no | PASS |
| `shard-020` | 31.481 | 13.570 | 29 | no | PASS |
| `shard-021` | 31.239 | 13.660 | 29 | no | PASS |
| `shard-022` | 28.860 | 13.702 | 29 | no | PASS |
| `shard-023` | 27.728 | 13.512 | 29 | no | PASS |
| `shard-024` | 26.378 | 13.985 | 29 | no | PASS |
| `shard-025` | 32.448 | 13.709 | 29 | no | PASS |
| `shard-026` | 32.181 | 13.428 | 29 | no | PASS |
| `shard-027` | 26.309 | 13.438 | 29 | no | PASS |
| `shard-028` | 27.796 | 13.380 | 29 | no | PASS |
| `shard-029` | 26.063 | 13.789 | 29 | no | PASS |
| `shard-030` | 27.705 | 13.307 | 29 | no | PASS |
| `shard-031` | 29.185 | 13.407 | 29 | no | PASS |

## Union and failure proof

- `TestNativeOracleShardUnion` passed (0.065 seconds in the final measurement round):
  all **933 registered fixtures**, exactly once across 34 shards. It rejects omitted
  and repeated shards and checks name-based ownership independently of registration order.
- Combining the final raw logs produced **933 distinct fixture terminal PASS records**,
  exactly 933 total, with no duplicates. `fixture-terminal-union.json` records that set;
  `final-times.json` records each shard's fixture count, setup time and zero cache hits.
- `TestNativeOracleShardPlantedFailure` passed in **31.475 seconds** with fresh runtime
  caches. Its subprocess mutates the lowered dedication, runs its owning `shard-001`
  through the live fixture comparison, and must exit 1 with the dedication's terminal
  FAIL and `stdout differs`. A timeout or other exit cannot satisfy the proof.
- `TestNativeOracleGateShardCoverage` passed in **0.664 seconds** through `go test`.
  It enumerates the actual 34 names, supplies parent PASS plus the other 33 child PASS
  records, and requires `shard-000` to remain missing in the gate's coverage map.
- `git diff --check` passed. The whole compatibility parent was not rerun: it deliberately
  retains the complete workload and is selected by children in the gate.

## Reproduction and retained evidence

`measure.py` runs selected names with a prepared binary, requires unused cache directories,
and enforces the process-group deadline. Use repeated `--run` options for selected children;
it stops at the first failure or over-target unit so only replacements need rerunning.
Run with the tool environment sourced and disk-backed `TMPDIR` set:

```sh
python3 internal/oracle/evidence/split-native-node/measure.py \
  --binary /workspace/oracle-shards.test --out /workspace/new-cold-evidence \
  --run TestNativeAgreesWithNode/shard-019-1
```

`logs.tar.gz` contains raw logs for all measurement rounds, the killed original baseline,
the final coverage/planted proofs, and the actual `go test` command of the slowest shard.
`measurement-round-*.json` preserves provisional measurements without turning over-budget
or unrun units into green gate results. The historical 75-second shell timing files are
retained from the previous rule; subsequent tests use the new 90-second deadline.
