# Library regex and Unicode child guards

Branch `codex/library-childguard` starts at main
`efe9f4042049234e5a52639fe77b47c311fd530c`. Main's shared childguard landed
at 04:35 MDT. No pending-branch history or childguard implementation changes
are imported by this unit.

The converted runners use shared defaults: 30 minutes to first output,
two minutes without output thereafter, and a 60-minute ceiling. Every
execution gets a fresh guard. Native regex's two sequential modes no longer
share a context. Progress comes from completed work in the child, never a
parent heartbeat. Node JSON and Unicode results remain on stdout; their
progress goes to stderr so result parsing and comparisons remain strict.

| Inventory row | Site | Progress and guard |
|---|---|---|
| 13 | native random Node oracle; regexp nodeResults | Synchronous stderr after every 128 completed cases; childguard.Run |
| 18 | native runRegexCases, limited and unlimited | Flushed printf after every 128 completed cases and at the end; a fresh childguard.CombinedOutput for each mode |
| 19 | Unicode runScanner | Existing per-property/range output renews childguard.Run |
| 21 | Unicode runNode, all three scripts | Legacy values every 4,096 units; legacy/Unicode scans every eight completed patterns; childguard.Run |

Safe sites remain unchanged except for one-line comments:

- native regexp_test step-limit watchdog: the child enforces a 1,000-instruction CPU budget.
- regexp oct6NodeLoops: 38 bounded loops, 179 executions, measured 0.092s against 30s (>300x margin).
- Unicode string properties: 7,906 bounded sequence answers, measured 0.343s against 120s (>340x margin).

## Proof

`run-library-childguard-proofs.py` uses Go overlays, leaving repository sources
untouched during mutations. Each of eight child paths is replaced first by a
silent infinite loop and then by one output followed by an infinite loop:
limited native, unlimited native, native's Node oracle, matcher nodeResults,
property scanner, legacy values, legacy scans and Unicode scans. Only the
proof substitutes two-second first-output/stall windows. Successful compilation
and the expected named childguard termination are required; an outer timeout
is a failed proof. All 16 normal-load checks passed. All 16 passed again under ten CPU spinners
before the unmutated comparisons ran: 32 successful kill proofs, with no
compiler-warning or outer-timeout catch counted.

The saturation phase runs ten CPU-bound workers for `nproc=5`; cgroup quota
is four CPUs. Every worker must remain alive through all tests and have consumed
CPU. Start/end load, elapsed time and each worker's CPU time are recorded in
`load.json`. Workers are terminated and joined after the proof. Production
childguard windows remain unchanged. The deliberate saturation run's outer
Go test bound is two hours; this is not a replacement child deadline or a claim
that the complete repository gate was run.

```sh
source /workspace/adamic-tools/env.sh
python3 internal/native/testdata/run-library-childguard-proofs.py --logs /tmp/library-childguard-final
# The runner precompiles the touched packages, runs all overlays normally,
# repeats all overlays with ten spinners, then runs every converted comparison:
go test ./internal/childguard ./internal/native ./internal/regexp ./internal/unicodeproperties -run '^(TestRegExpBytecode.*|TestRegExpNativeStepLimit|TestMatcher.*|TestNodeAgrees|TestNodeStringProperties|TestCanonicalizeLegacyNode|TestCanonicalizeUnicodeNode|TestProgress|TestStalled|TestCeiling|TestExitIsNotGuardError|TestKillsProcessGroup|TestNoFirstOutput)$' -count=1 -v -timeout=2h
go vet ./internal/native ./internal/regexp ./internal/unicodeproperties
gofmt -l internal/native/regexp_test.go internal/regexp/matcher_oracle_test.go internal/regexp/oct6_oracle_test.go internal/unicodeproperties/node_test.go internal/unicodeproperties/canonicalize_test.go
git diff --check
```

All selected comparisons passed under the ten-worker load:

| Package | Seconds | Observation |
|---|---:|---|
| childguard | 9.260 | Progress, silence, stalled output, ceiling, exit identity and process-group tests passed |
| native | 204.951 | 127,369 test262 cases and 10,000 random cases, each in both modes, plus UTF-16 identity and CPU budget |
| regexp | 15.185 | All matcher tests, including every Node caller and existing semantic mutants |
| unicodeproperties | 1,567.411 | Legacy 26.45s, Unicode 1,512.47s, property scanner 27.89s, string properties 0.56s; zero disagreements |

The full loaded phase (including its hang proofs) took 1,773.166s. Every
spinner remained alive throughout, consuming 568.08 to 673.53 CPU seconds.
Load average was 10.49 at completion. All workers were terminated and joined.
Vet, formatting and diff checks passed. `load.json` records the measurements;
`proof-logs.tar.gz` contains complete normal/loaded hang logs, comparisons,
setup, polling and static-check logs. No timing threshold in the test sources
was raised to achieve these results.

## Preparation observations

Initial setup on main `749a69ad` took 83.076s: Node 0.036s, Go 0.081s,
markdown 0.177s, submodules 0.189s, clang 0.422s, Go build 82.781s,
cache warm 83.034s. Node 24.19.0, Go 1.27.1, clang 20.1.8.
Environment `/workspace/adamic-tools/env.sh`, `nproc=5`, quota four CPUs,
memory 17.6 GB. Main was fetched every 45 seconds until the helper landed.

The original baseline passed native regex in 58.727s, matcher Node comparisons
in 1.818s, and the Unicode comparisons in 533.585s. Unicode canonicalization
examined 5,653,004,288 code points; the scanner examined 1,910,702,080, both
with zero disagreements. An isolated scratch preflight against the pending API
also passed all eight silent mutations and all progressing comparisons
(native 51.490s, matcher 1.940s, Unicode 501.010s, shared guard 9.139s).
Its first proof supervisor stopped cold Go compilation before any child ran;
compilation was then separated from kill proof supervision and the retry passed.
That timeout is not counted as a caught hang.

Limits: no whole repository or whole oracle gate is claimed. Runtime regex and
Unicode implementations, fixture payloads, counts and the three safe deadline
policies are not changed. The fast gate owns remaining repository coverage.
