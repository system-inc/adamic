Built: emission-only CPU profile of the unmodified stage-1 lint port through Go APIs, task #3he8f8g.
Commits: base cd0ecb04d63fe873383154cb6470a3c66861d0cb; profile evidence 60b137d7; final delivery SHA is reported with the pushed branch.
Commands and outputs: GOMAXPROCS=4 /tmp/c-emission-profile; lower 4.124505348s, native.C 72.431460293s, 6,405,618 C bytes.
Mutants: none; this delivery adds evidence only, with no compiler, fixture, or test change.
Not covered: sf-yepesta's mutant snapshot, the other two port snapshots, full oracle C equivalence, an implemented fix, or optimized cold timing.

The implemented fix and full byte-comparison evidence are in [fix/REPORT.md](fix/REPORT.md).

## Observation

Fresh branch compiler/c-emission-profile from origin/main. Recipe is
stage1/cohere/lint/shared_test.go:102-114: registry.Generate, load.Load on
stage1/cohere/lint/main.ts, lower.Lower(context.Background()), native.C.
The probe uses these same APIs on the unmodified in-place port. It does not
run the test harness's cache, JavaScript emission, clang, or the port binary.
CPU profiling starts immediately before native.C and stops immediately after it;
the wall clock excludes profile startup/shutdown and writing C to disk.

The first run overlapped setup: lower 11.282688507s, native.C 82.261666932s.
The second fresh-process run followed setup: lower 4.124505348s,
native.C 72.431460293s. Both emitted 6,405,618 bytes. Four-CPU cgroup quota
(cpu.max 400000 100000), nproc 5, GOMAXPROCS=4. This is a newly provisioned
instance but the second run has warmed Go build dependencies and filesystem
state; do not label it a cold optimized measurement.

## Culprit and scaling

internal/native/emit_objects.go:427 dynamicProperties, whole-program walk at
line 429; fieldTypesNeeded at line 447 invokes it at 448 and performs another
whole-program walk at 452. shapeWith calls dynamicProperties before even
checking its shape cache (line 265). objectLiteral calls fieldTypesNeeded
once per field (line 205). These facts do not vary per call, but each query
walks every expression in every function and main by reflection.

The profile attributes 37.74s cumulative to dynamicProperties and 38.81s to
fieldTypesNeeded (these overlap and must not be added). Secondary instances
of the same pattern: fieldReadinessNeeded, emit_objects.go:461/466, 5.56s;
hasRecordStorage, entries_records.go:32/34, 5.10s. walkExpressions totals
68.27s cumulative (97.67% of 69.90s CPU samples). This identifies repeated
program scans as the dominant cause, rather than C buffer concatenation.
The earlier source-only candidate, borrowChain, does not appear among the
profile's top 50; it is not the measured dominant cause here.

Inference: with N IR expressions and Q shape/field/readiness/record queries,
this work is O(N*Q). When fields, object literals and accesses grow with the
program's statements, Q grows with N and the cost is quadratic. Function count
contributes through their bodies; string-literal bytes and type count alone
are not the relevant input dimensions. No synthetic scaling series was run.

## Proposed fix and expected saving

Compute dynamic-property presence, field-type metadata presence, record-storage
presence and the set of readiness field names in one walk per emitter. Keep the
existing CheckedFields and UninitializedFields semantics, including hand-built
IR fallbacks. Cache both false and true answers. Preserve traversal order and
all declaration/shape creation order. Do not cache globally across programs.

Expected saving, an inference from the profile: remove most of the roughly
68 CPU seconds spent in repeated walks; single-digit-second emission is a
reasonable target, not a measured result. writeFieldSlot also scans layouts
and the program and can be assessed after the invariant queries are cached.
This evidence-only unit does not ship the cache: the required complete oracle
and three-port byte-equivalence validation has not been completed. No optimized
native.C time or mutant result is claimed.

## Reproduction

From the repository root:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/c-emission-setup.log 2>&1
source /workspace/adamic-tools/env.sh
mkdir -p .scratch-c-emission
cp review/compiler/c-emission-profile/probe.go.txt .scratch-c-emission/main.go
go build -o /tmp/c-emission-profile ./.scratch-c-emission > /tmp/c-emission-build.log 2>&1
GOMAXPROCS=4 /tmp/c-emission-profile > review/compiler/c-emission-profile/profile.log 2>&1
go build -o /tmp/adamic-pprof cmd/pprof > /tmp/c-emission-pprof-build.log 2>&1
/tmp/adamic-pprof -top -cum -nodecount=10 /tmp/c-emission-profile review/compiler/c-emission-profile/lint.pb.gz
```

Installed go tool pprof failed with `go: no such tool "pprof"`; building the
bundled cmd/pprof succeeded. Probe source is .go.txt, never compilable Go under
review. The generated C remains scratch data at /tmp/c-emission-lint.c.
setup.log preserves every timing line: Go 0.035s, Node 0.036s, clang 0.246s,
markdown dependencies 1.028s, submodules 16.937s, cache warm 285.155s,
total 285.183s. Setup succeeded, no toolchain workaround beyond bundled pprof.

## Top 10 cumulative functions

Unfiltered pprof output, including emitter entry points and runtime wrappers:

```text
File: c-emission-profile
Build ID: fa145ac134c5a681f9e57d71c36668ff02040cbf
Type: cpu
Duration: 72.43s, Total samples = 69.90s (96.50%)
Showing nodes accounting for 28.94s, 41.40% of 69.90s total
Dropped 220 nodes (cum <= 0.35s)
Showing top 10 nodes out of 64
      flat  flat%   sum%        cum   cum%
     2.12s  3.03%  3.03%     68.74s 98.34%  github.com/system-inc/adamic/internal/native.walkExpressions.func1
     0.10s  0.14%  3.18%     68.27s 97.67%  github.com/system-inc/adamic/internal/native.walkExpressions
     0.95s  1.36%  4.54%     68.12s 97.45%  github.com/system-inc/adamic/internal/native.walkStatement
         0     0%  4.54%     67.71s 96.87%  github.com/system-inc/adamic/internal/native.(*emitter).statement
    25.77s 36.87% 41.40%     67.68s 96.82%  github.com/system-inc/adamic/internal/native.walkStatement.func1
         0     0% 41.40%     67.66s 96.80%  github.com/system-inc/adamic/internal/native.cProgram
         0     0% 41.40%     67.64s 96.77%  github.com/system-inc/adamic/internal/native.(*emitter).statementAt
         0     0% 41.40%     67.59s 96.70%  github.com/system-inc/adamic/internal/native.C (inline)
         0     0% 41.40%     67.59s 96.70%  main.main
         0     0% 41.40%     67.26s 96.22%  runtime.main
```

The wider attribution and source-line listings are in top-50.txt and queries.txt.
Cumulative times include descendants and can overlap; recursive walkers can
have cumulative totals larger than their top-level caller.

## Validation and roadmap contribution

No tests or fixtures were added or touched, so counts.md and test-leaf grain
are unchanged. No correctness check or mutant claim is introduced. No whole
package test or full gate was run. Lane-check output is in lane-checks.log.
This lands a measured cause and a concrete cache proposal toward the C emission
performance unit #3he8f8g. The brief provides no numeric roadmap step; none is
invented. Conservative scope assumption: profiling the unmodified lint recipe
satisfies the requested one-program measurement; it does not imply coverage of
all named snapshots.

Lane checks passed: `lane checks 0.3 s: gofmt and tools on 0 Go files, t.Parallel on 0 test packages`. The prescribed fetch left no origin/cloud/merge-tree ref because this checkout has a narrow fetch refspec. Explicitly fetching main, devtools/fast-gate and cloud/merge-tree into their origin refs fixed it; the same lane-check script then exited 0.
