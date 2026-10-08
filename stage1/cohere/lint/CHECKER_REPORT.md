# Checker on RuleContext

Branch: `lint-checker/harness`, based on `origin/area/stage1-lint`
65017b318da1995237ff3ea2c80f59b055b39ac3, which includes main's pin bump.
Library commit 6131c4c is deliberately absent from this branch. Native driver
validation below used an isolated scratch merge with that commit. Native
integration depends on its returned-result builtins reaching main.

## Implementation and bridge surface

The manifest accepts a run-level `program <tsconfig>` line. The driver opens
one program before rows, uses the tsconfig's own roots, and releases it after
rows. Without the line the run remains syntax-only. Each parsed row gets one
Checker holding the run's handle, path, parser and UTF-16-to-UTF-8 offset table.
`ask(index, question)` uses the bridge's exact byte range and kind selector.
`askFile(FileQuestion)` is the cross-file channel and refuses a ProgramReads
kind absent from the active rule's descriptor.

Descriptors support `typed`, the five upstream ProgramReads names, and named
SyntaxKind listeners. Typed rules are selected only with a checker. Coverage
otherwise prints `skipped <rule> no program`. Fixes use the initial typed
findings, but fix re-lint passes have no checker. The Go oracle builds a real
cohere program from the same tsconfig, checks row source against it, and passes
its checker and ProgramView to the upstream rule.

The harness calls tsgoProgram, tsgoInspect and tsgoRelease. Its wrappers use
library's actual envelopes: `{kind: 'Ok', value: T} | TSGoError`, and
`{kind: 'Ok'} | TSGoError` for release. TSGoError has kind Error and the C
error-buffer message. A bridge inspection Error becomes an Answer refusal and
`refused <rule> <path> <start> <end> <reason>` on the wire. Open and release
Errors also reach the wire. Query and typeParts are not needed by this pilot.
No unlanded question implementation or library commit was added to this branch.

Native recording writes a run header naming the tsconfig and its SHA-256,
plus one transcript per row bound to its path and source SHA-256. Entries frame
the exact selector/question, Value or Error, and payload. Node and emitted
JavaScript use the same Checker to replay. Missing, mismatched and unasked
entries refuse; different program/source headers refuse. The Node fallback
bridge exports return Errors instead of panicking; replay never calls them.

## Pilot and controls

The P1 pilot is `@typescript-eslint/no-unnecessary-boolean-literal-compare`.
It exercises node `type-shape` asks, file `ReadsCompilerOptions/options` asks,
options, nullable repairs, and refusals with a single BinaryExpression listener.
Its module is `.a`. Upstream capture now includes the typed test family and
records typed assertions, so 102 pilot cases compare live Go, native, Node and
emitted JavaScript byte for byte on findings, fixes and suggestions.

The Go comparison found a real nullable fix mismatch in an empty-initializer
`for (; x === true;)`: the port supplied `?? false` where Go uses `x`. The
condition-position check was corrected, then TestRulesAgree passed. A generic
`T extends boolean` control also agrees with live Go facts.

1. No-program coverage: the planted omission prints ignored and fails the
   coverage check; restored selection prints no program and agrees with Go.
2. Missing transcript: removing an entry fails Node's ordinary comparison;
   restoring it passes. An extra entry also fails the reverse guard. Changing
   the program hash fails replay. These facts are obtained from the live
   bridge, and the library scratch native recording verifies the same facts.
3. Bridge refusal: in the isolated library merge, an unsupported checker
   question returns the C message and produces the same refused line on
   native, Node replay and emitted JavaScript replay. Removing refusal
   handling still builds and runs all three, but wire comparison catches each
   mutant. Restoring handling passes all three again. On this branch the test
   remains the named skip `awaits codex/tsgo-errors-as-values` until the prelude
   exposes TSGoError. That skip is pending, never a pass. The bridge call that
   needs the result is tsgoInspect / adamic_tsgo_inspect / C tsgo_inspect.

Additional controls reject unknown/duplicate ProgramReads and reads on an
untyped descriptor, refuse undeclared ReadsOtherFiles, catch a changed pilot
type verdict by findings comparison, and compare SHA-256 vectors against Go.
Native controls and comparisons use address/undefined sanitizers and the
repository allocator leak checker. No sanitizer or leak finding occurred.

## Reproducible validation

Setup used GOPROXY=https://proxy.golang.org|direct. Timing lines: node 0.043s,
Go 0.043s, clang 0.327s, markdown dependencies 1.403s, submodules 5.307s,
Go build 358.883s, total 359.392s. nproc 5, CPU quota 4.

The required TypeScript source input is v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8. Its extracted archive initially lacked
Git metadata; the profile pin check failed. Fetching that commit into the
external corpus directory established the pin with a clean source diff.
No check was skipped or weakened. Each profile run uses one fresh directory
for both ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS.

Commands from the repository root, after sourcing the setup environment:

```sh
GOMAXPROCS=4 go test -count=1 -v -timeout 10m ./stage1/cohere/lint/registry
GOMAXPROCS=4 go test -count=1 -v -timeout 10m ./stage1/cohere/lint \
  -run '^TestChecker(NoProgramCoverage|ReplayEntryControl|Hashes|BridgeRefusalPending)$'
```

The registry passed (1.641s). Own-branch controls passed (16.746s), with only
the explicitly pending bridge control skipped. Scratch commands add
GOFLAGS=-buildvcs=false because its nested cohere checkout is a shared symlink;
this disables executable VCS stamping, not correctness checks.

```sh
GOFLAGS=-buildvcs=false GOMAXPROCS=4 go test -count=1 -v -parallel 1 \
  -timeout 30m ./stage1/cohere/lint -run '^TestRulesAgree$'
GOFLAGS=-buildvcs=false GOMAXPROCS=4 go test -count=1 -v -parallel 1 \
  -timeout 40m ./stage1/cohere/lint \
  -run '^(TestChecker(NoProgramCoverage|ReplayEntryControl|Hashes|BridgeRefusalPending)|TestOwnedWitnesses|TestProfile(Artifacts|Compilation|SnapshotsAgree))$'
GOFLAGS=-buildvcs=false GOMAXPROCS=4 go test -count=1 -v -parallel 1 \
  -timeout 30m ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$'
```

TestRulesAgree passed (244.103s): 102 typed pilot comparisons and 13,746,146
bytes of syntax-rule comparisons across 3,879 captured cases. Existing explicit
parser-recovery refusals remain visible. The compiler/repository corpus passed (774.385s): 498 files and 28,440,570
bytes identical across Go, sanitized native, Node and emitted JavaScript.

The focused lint package run completed with FAIL (812.892s), not green.
All four checker tests passed in the library scratch merge, including the
required refusal control (176.47s). Profile artifacts passed (98.69s), shared
profile compilation passed (105.39s) with 336 allocations and 336 frees,
and profile snapshots passed (297.18s), comparing 41,805,211 bytes across
Go, release, profiled, Node and emitted JavaScript. The typed pilot's owned
witness also passed all four runtimes (1,996 bytes).

TestOwnedWitnesses then failed on the existing, unchanged
`nexus/consistency-no-single-line-jsdoc` Unicode witness. Native and Node
produce `// Unicode prose.`; Go at the current pin produces
`// \u0085Unicode prose.\u0085`. The minimal source is
`/**\u0085Unicode prose.\u0085*/` followed by `const value = 1;`.
The rule directory has no diff against the area base. It was left unchanged:
this is a real finding in another rule, not permission to weaken the test.
The branch is therefore blocked from a green lint package until that rule's
owner resolves it, in addition to its pending native library dependency.

Reproduce using the saved fixture:

```sh
mkdir -p /tmp/checker-jsdoc-repro
cp stage1/cohere/lint/checker-proof/jsdoc-nel.repro.txt /tmp/checker-jsdoc-repro/source.ts
printf '%s\tnexus/consistency-no-single-line-jsdoc\n' /tmp/checker-jsdoc-repro/source.ts > /tmp/checker-jsdoc-repro/manifest.txt
/workspace/checker-library-unit-final/oracle --manifest /tmp/checker-jsdoc-repro/manifest.txt
/workspace/checker-library-unit-final/scanner --manifest /tmp/checker-jsdoc-repro/manifest.txt
node --disable-warning=ExperimentalWarning oracle/node.mjs /workspace/checker-library-unit-final/main.ts --manifest /tmp/checker-jsdoc-repro/manifest.txt
```

Evidence is saved in `checker-proof/`, including the failing gate, the
minimal Go/native/Node reproducer outputs, all controls, successful
TestRulesAgree and corpus logs, and instruction measurements.

A broader lint run was interrupted because repeated legacy mutant builds were
too slow. Before interruption it passed fixes/suggestions, .a rename, node
layout over 4,016 rows (13,792,057 bytes), and shard parity over 4,054 rows
(33,976,741 bytes) at 1, 2 and 5 shards. Its two findings, the nullable for fix
and missing corpus Git metadata, were fixed and their tests rerun. It is not
reported as a completed full package gate.

## Single-threaded measurement

GOMAXPROCS=1, GODEBUG=asyncpreemptoff=1, Valgrind 3.24 Callgrind, one pilot
witness with strict tsconfig. Native release uses clang -std=c11 -O2
-ffp-contract=off -fno-optimize-sibling-calls -DADAMIC_TSGO. The checker archive
uses go build -buildmode=c-archive with default Go optimization. The oracle
uses go build with default optimization. Profile builds add -g. Sanitizer
builds use -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all.

Native: 392,339,258 instructions; Go: 442,628,194 instructions. Outputs match,
stderr is empty. Counts include program startup, checking, rule execution,
fixing and release; they are not isolated rule-body counts. Warm timing uses
one discarded warm-up and five measured runs per runtime. The first medians
(97.28ms native, 92.53ms Go) overlapped gate work. A second, interleaved run
after the gates ended measured 92.92ms native and 83.11ms Go. These are whole
process observations on one witness, not a broad speed claim. Artifacts are
/workspace/checker-library-measured, including callgrind files, build sources,
manifest, pilot-timing.json and pilot-timing-quiet.json.

## Scope left out

Question registration by directory and fleet migration are expressly outside
this unit. A parallel checker, broad typed-project throughput, other five-call
error controls, and the complete repository gate were not added or claimed.
Syntax compiler/repository corpus comparison does not assert typed pilot
coverage on every compiler file. The JSDoc witness failure and pending native dependency remain visible
on the harness branch until library's area reaches main.


## Area merge and whole lint package rerun, October 8

Merged origin/area/stage1-lint at 79f3e2868ffd2b91f8fac187a6cee3466222952b
without rebasing. Merge commit: 62ab7b586618b78f32951aa343634d2d469cdbdf.
The merge contains d45a323be's ECMAScript whitespace correction. The saved
U+0085 JSDoc reproducer now gives identical Node and Go findings and fixes.

The whole lint package was run on this branch, without a rule or test filter:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-wave08-corpus
export ADAMIC_LINT_PROFILE_DIR=/workspace/checker-area-whole.nvxuQI
export ADAMIC_LINT_PROFILE_SNAPSHOTS="$ADAMIC_LINT_PROFILE_DIR"
GOMAXPROCS=4 go test -count=1 -v -timeout 30m ./stage1/cohere/lint
```

The profile directory was fresh at the start. The TypeScript checkout is
clean, with no tracked diff or untracked files, at the required 6.0.3 commit
050880ce59e30b356b686bd3144efe24f875ebc8. Output was written directly to a log.
/usr/bin/time is unavailable, so a Python monotonic clock measured the command.

Result: FAIL, exit 1. Wall time: 319.727 seconds. Go package time: 302.428s.
The native driver cannot type-check against this area's scalar bridge API:

- checker.a:10:88: TS2322, number is not assignable to ProgramResult.
- checker.a:11:66: TS2322, void is not assignable to ReleaseResult.
- checker.a:13:5: TS2322, string is not assignable to InspectResult.

These are the pending library dependency, not a replacement JSDoc finding.
Library 6131c4c remains absent from this branch. No unsafe conversion, library
merge, or check relaxation was used to turn this run green. Profile snapshot
failures cascade from the native profile builds being unavailable. The
no-program coverage, replay-entry and hash controls passed, as did the JSX
whole-tree comparison of 63 captured sources (51,432 bytes).

TestCheckerBridgeRefusalPending still explicitly skips with
`awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer`.
Optional throughput tests also skip because ADAMIC_LINT_BENCH was not requested;
the correctness corpus and profile inputs were supplied.

Pilot askFile confirmation: `ReadsCompilerOptions/options` is issued as
`checker.askFile(new FileQuestion('ReadsCompilerOptions', 'options'))` in
prepare. Expression `type-shape` questions use checker.ask; the forced bridge
refusal control also uses ask. No change to those calls was needed.

Evidence: checker-proof/area-whole-lint.txt, area-whole-time.txt,
area-whole-inputs.txt, area-typescript-clean.json, and area-jsdoc-{go,node}.txt.

## Scalar bridge adapter and renewed package gate

`checker_bridge.a` is the single compatibility boundary for today's raw
bridge returns. It wraps open and inspect values in `{kind: 'Ok', value}`
and release in `{kind: 'Ok'}`. Checker and the driver retain the existing
Answer-or-TSGoError handling above it. Today's bridge still panics on an
error; this adapter does not claim to convert those panics into refusals.
Library 6131c4c remains absent from this branch.

When library's returned-result builtins land, replace this one import in
checker.a:

```ts
import { openBridge as tsgoProgram, inspectBridge as tsgoInspect, releaseBridge as tsgoRelease } from './checker_bridge.a';
```

with:

```ts
import { tsgoProgram, tsgoInspect, tsgoRelease } from 'adamic';
```

Scratch merge b2eed32f631e2e17e4d323270085578b60a219ca includes library
6131c4c, the harness test corrections, and exactly that import change.
`TestCheckerBridgeRefusalPending` passes there in 171.629s. The forced
unsupported checker question reaches the refused wire line on native,
Node and emitted JavaScript. Removing refusal handling builds and runs on
all three but fails their wire comparisons. Restoring it passes all three.
The control on this branch still names its pending dependency:
`awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer`.

The first full adapter run found a test setup defect: the registered typed
mutant was given a syntax-only manifest and silently skipped. A focused
reproducer failed in 84.864s because the boolean verdict mutant survived on
Node. Typed mutant runs now use actual strict projects, the real Go oracle,
and native recordings replayed by Node and emitted JavaScript. The same
mutant is caught only by findings comparison on all three; the focused
corrected run passes in 67.774s. The initial full run was interrupted after
this finding and is explicitly retained as incomplete, not green.

The native binary cache still keys by generated C hash, bridge archive and
instrumentation. Identical builds share a completion signal; different
sources compile independently under the existing bounded mutant workers.
Previously its global lock serialized all compilations. No comparison or
mutant was removed to shorten the gate.

Evidence for these changes is under checker-proof/adapter-*.txt, including
the one-line scratch pass-through patch, the typed mutant before and after,
and the scratch refusal control's removal and restoration.

The renewed full package gate passed, exit 0, on harness code 542ecf7c2.
Go package time was 4662.856s; Python monotonic wall time was 4671.349
seconds (77 minutes 51 seconds). No test filter was used:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-wave08-corpus
export ADAMIC_LINT_PROFILE_DIR=/workspace/checker-adapter-final.iAVfPc
export ADAMIC_LINT_PROFILE_SNAPSHOTS="$ADAMIC_LINT_PROFILE_DIR"
GOMAXPROCS=4 go test -count=1 -v -timeout 90m ./stage1/cohere/lint
```

The TypeScript checkout was clean before and after, at 6.0.3 commit
050880ce59e30b356b686bd3144efe24f875ebc8. The shared profile directory was
fresh at the start. The longer package timeout accommodates the complete
native mutant builds; it changes no correctness condition.

TestRulesAgree passed in 136.50s with 102 typed cases and 13,743,110 bytes
identical across live Go, native, Node and emitted JavaScript. The compiler
and stage-1 corpus comparison passed on 719 files and 28,685,583 bytes.
All 75 registered mutants passed, including the typed pilot on all three
port runtimes. Shards, owned witnesses, registration, factory hooks, option
decoding and its mutant, and profile comparisons passed. Native sanitized
builds used ASan, UBSan and LeakSanitizer without findings. Counted profile
compilation reported 336 allocations and 336 frees.

The named bridge refusal control remains a pending skip on this branch.
Optional throughput checks were not requested. Existing explicit parser
recovery limits are retained in the complete log; none were relaxed.
This is the whole lint package, not a claim of the repository-wide gate.
Question registration and fleet migration remain outside this unit.

Final evidence: checker-proof/adapter-final-lint.txt,
checker-proof/adapter-final-inputs.txt and checker-proof/adapter-final-time.txt.

## Performance remediation, October 8

Code commit: d85b9e68b4f91ece59194014a9e40c0f755dc6f0. The acceptance
run passed the entire lint package under `-timeout 60m`, without a filter.
Go printed 2979.178s; monotonic command wall time was 2983.204s (49m43s).
The exact merged area tip, 79f3e2868ffd2b91f8fac187a6cee3466222952b,
passed in 3902.748s, wall 3908.597s (65m09s), on the same box. The target
was area time plus at most 150s, while also passing the 60-minute limit.
The candidate is 925.393s faster than area and 1688.145s faster than the
original 4671.349s wall run. Both targets are met.

Both runs used nproc=5, CPU quota 400000/100000 (four cores), GOMAXPROCS=4,
GOFLAGS=-buildvcs=false, the same installed toolchain and compiler pin,
and no competing gate. Median sampled one-minute load: area 1.281,
after 1.312; ranges 0.557–5.129 and 0.938–5.592 respectively. Those are
whole-run observations, not a claim of equal load at every instant. The
original run did not record load, so its load cannot be reconstructed.

An initial untouched-area run with `-timeout 60m` failed at 3600.090s
inside profile snapshots. It is retained as a timeout, not green. Its
3821.625s outer wall also includes cold Go compilation before package timing.
The completed area timing benchmark used `-timeout 90m` solely to obtain
all results. Two intervening attempts were killed by environment/tool
session restarts, after 46 mutants and during the corpus respectively;
neither is counted as completed. The final sequential benchmarks ran in a
detached process with file-based logs, without overlapping their work.

### Where the apparent extra 2100 seconds went

The seat's approximately 2500s is not this box's measured baseline. On this
box, untouched area took 3908.597s: about 1409s above that estimate. Original
harness was 762.752s above this box's area wall time, or 760.108s in Go
package time. Seat conditions were not supplied, so a more specific causal
explanation of the seat-to-box difference would be speculation.

The original mutant children totaled 8556.92 active seconds (75, mean
114.09s); area totaled 5857.65 (74, mean 79.16s). These overlapping child
times cannot be summed into package wall time. Area's measured sweep span
was 1501.581s. The old log has no timestamps; accounting for its sequential
reported durations and final parallel group estimates its sweep at roughly
2180s. That inference attributes about 679s of the 760s package excess to
mutant builds, with about 81s net in other tests and scheduling. It is an
estimate, not an observed old sweep span. The new candidate's measured
75-mutant sweep is 880.419s (mean child 44.56s), saving 621.162s against area.

Measured additions and overlapping components:

| Component | Observed cost |
|---|---:|
| 102 typed cases, all four runtimes combined | 65.779s |
| Live Go program and lint within those cases | 7.558s |
| Native program, lint and recording within those cases | 10.176s |
| Node transcript replay within those cases | 39.657s |
| Emitted JavaScript replay within those cases | 8.389s |
| Own-branch no-program, replay and hash controls, combined | 13.68s |
| Pending bridge refusal control on own branch | 0s, named pending skip |

The component rows overlap the typed-case total; do not add them twice.
The actual `TestRulesAgree` net increase over area is 13.93s because saved
setup work offsets most of the new typed runtime work. Each captured typed
case retains its own strict project: combining script files would introduce
shared globals and change the upstream tests' meaning. Within each manifest,
the native driver and the Go oracle already built one program before all
rows; neither was opening a program per row. Those lifetimes were preserved.

A separate paired build probe consumed identical saved generated C and
unchanged sanitizer flags. Old monolithic BuildTSGo: 64.483s; new split cold:
74.616s; identical warm split: 4.273s. All outputs matched the live Go oracle.
Cold structural builds can regress: FactoryHooks is 213.73s after versus
117.53s for area. The full tables retain this cost instead of hiding it.

On the probe's one pilot project, native bridge load was 75.043ms, two asks
0.271ms, and cohere's real program.Build was 41.333ms. Whole native process
was 118.574ms, with recording 120.080ms; the 1.506ms difference is one noisy
sample, not a stable recording benchmark. Node replay was 463.033ms and
emitted replay 109.235ms. These micro measurements use the saved original C
for an identical-C build comparison; the 102-case totals above measure the
actual candidate. They show that program loading and recording are small
compared with repeated compiler builds, without inventing an exact causal
allocation of every second.

### Fix and preserved checks

The harness caches immutable generated C and JavaScript by the complete
static import/export dependency contents and absolute paths, including
type-only dependencies. Each distinct source is loaded and lowered once,
and both outputs come from that same lowering. It no longer emits ordinary
C before TSGoC emits it again. Mutants change the content key; rename and
outside-module-copy controls retain path-sensitive invalidation.

Checker builds use the already-landed native.BuildSplitTSGo runtime and
object caches, with Jobs=1 per build under the existing four mutant workers.
No compiler implementation changed. Sanitizer flags remain -O1 -g
-fsanitize=address,undefined -fno-sanitize-recover=all, with the same strict
C flags. Release/counted builds reuse the landed split cache; the profiling
build retains its full -O2 -g compilation. Generated JavaScript is written
to a fresh file per test so emitted-output mutants cannot poison the cache.
Each mutant binary is cleaned up after its subtest rather than retaining
75 checker-linked binaries that cannot be reused.

Syntax-only rows no longer compute a source hash without recording/replay.
Ordinary typed runs no longer construct transcript keys or store facts unless
recording/replay needs them. Recording, replay, missing/unasked-entry guards,
program/source hash binding, refusals, and fix-pass exclusion are unchanged.

The entire package passed, including all 102 typed cases, 75 mutants,
719 compiler/repository files, shards, options, registration, factory hooks,
profiles and paused controls. Sanitized native comparisons had no findings;
counted profile instrumentation reported 259 allocations and 259 frees.
The allocator observation changed from 336/336 alongside removal of unused
hash/transcript work; no release check was relaxed. Pending skip still
names `awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError
from the C error buffer`; optional throughput tests were not requested.
Question registration, fleet migration, and the full repository gate remain
outside this unit. Library 6131c4c is not an ancestor of this branch.

Commands, after sourcing /workspace/adamic-tools/env.sh:

```sh
python3 /tmp/checker-package-timed.py area-detached /workspace/checker-area-before 90m
python3 /tmp/checker-package-timed.py harness-after /workspace/adamic 60m
```

The runner sets GOPROXY=https://proxy.golang.org|direct, GOMAXPROCS=4,
GOFLAGS=-buildvcs=false and ADAMIC_TYPESCRIPT_SOURCE to the clean v6.0.3
checkout at 050880ce59e30b356b686bd3144efe24f875ebc8. Each run gets a fresh
single directory for both required profile inputs. It invokes
`go test -json -count=1 -timeout <limit> ./stage1/cohere/lint` and writes
stdout/stderr directly to a log. Go's final JSON event says 2979.181s;
the 3ms difference from its printed package time is event delivery.

### Original run: every top-level test

The untimestamped original -v log supplies Go's reported own active times,
which exclude t.Parallel pause time. Exact old run/pause/continue timestamps
cannot be recovered from that log. TestMutants' 7.36s is its own body,
excluding parallel child work; it is not the sweep's wall time.

| Top-level test | Active time (s) | Paused | Result |
|---|---:|:---:|---|
| TestCheckerNoProgramCoverage | 4.79 | no | PASS |
| TestCheckerBridgeRefusalPending | 0.00 | no | SKIP |
| TestCheckerReplayEntryControl | 5.61 | no | PASS |
| TestCheckerHashes | 0.52 | no | PASS |
| TestNestedConstructorGap | 0.13 | no | PASS |
| TestEmittedJavaScriptMismatch | 98.83 | no | PASS |
| TestDotARename | 92.32 | no | PASS |
| TestCompleteSuggestionSerialization | 137.72 | no | PASS |
| TestSuggestionAlongsideAutomaticFix | 75.98 | no | PASS |
| TestWitnessScriptKind | 14.52 | no | PASS |
| TestJsxLintReleaseAndThroughput | 0.00 | no | SKIP |
| TestJsxLintTrees | 44.20 | no | PASS |
| TestRulesAgree | 136.50 | no | PASS |
| TestCompilerAndStage1Agree | 546.55 | no | PASS |
| TestLegacyMutants | 80.10 | no | PASS |
| TestThroughput | 0.00 | no | SKIP |
| TestNodeTableIsLinkOnly | 69.49 | no | PASS |
| TestShardsAgree | 291.60 | no | PASS |
| TestMutants | 7.36 | no | PASS |
| TestProfileArtifacts | 81.35 | no | PASS |
| TestProfileCompilation | 79.19 | no | PASS |
| TestProfileSnapshotsAgree | 294.56 | no | PASS |
| TestOwnedWitnesses | 31.83 | no | PASS |
| TestRegistrationMutant | 69.15 | no | PASS |
| TestFactoryHooks | 135.13 | no | PASS |
| TestNestedOutsideModuleCopy | 0.86 | no | PASS |
| TestDecodedOptionsAndMutant | 81.79 | no | PASS |
| TestOptionAndComparatorGaps | 0.78 | yes | PASS |
| TestCommentFoldMutant | 96.05 | yes | PASS |
| TestCountGuardMutant | 103.01 | yes | PASS |
| TestPositionIndexMutant | 108.47 | yes | PASS |
| TestDecorationOptionMutant | 109.74 | yes | PASS |

### Same-box area before and candidate after: every top-level test

Active intervals are calculated from timestamped JSON events: run-to-pause
plus continue-to-pass for paused tests. Go own durations are also shown;
event delivery and cleanup can produce small differences. TestMutants'
interval includes parallel children, so its own body and complete span are
shown separately. Skips are explicitly skips; an absent area test is —.

| Top-level test | Area Go own (s) | Area active intervals (s) | After Go own (s) | After active intervals (s) | After result |
|---|---:|---:|---:|---:|---|
| TestCheckerNoProgramCoverage | — | — | 4.79 | 4.79 | pass |
| TestCheckerBridgeRefusalPending | — | — | 0.00 | 0.00 | skip |
| TestCheckerReplayEntryControl | — | — | 8.24 | 8.24 | pass |
| TestCheckerHashes | — | — | 0.65 | 0.65 | pass |
| TestNestedConstructorGap | 0.19 | 0.19 | 0.15 | 0.15 | pass |
| TestEmittedJavaScriptMismatch | 60.49 | 60.50 | 79.60 | 79.60 | pass |
| TestDotARename | 113.61 | 113.61 | 21.90 | 21.89 | pass |
| TestCompleteSuggestionSerialization | 112.54 | 112.54 | 83.73 | 83.73 | pass |
| TestSuggestionAlongsideAutomaticFix | 61.26 | 61.26 | 16.84 | 16.84 | pass |
| TestWitnessScriptKind | 56.62 | 56.62 | 9.55 | 9.56 | pass |
| TestJsxLintReleaseAndThroughput | 0.00 | 0.00 | 0.00 | 0.00 | skip |
| TestJsxLintTrees | 50.54 | 50.54 | 44.67 | 44.67 | pass |
| TestRulesAgree | 102.15 | 102.14 | 116.08 | 116.08 | pass |
| TestCompilerAndStage1Agree | 638.55 | 638.55 | 535.59 | 535.59 | pass |
| TestLegacyMutants | 65.38 | 65.38 | 102.15 | 102.15 | pass |
| TestThroughput | 0.00 | 0.00 | 0.00 | 0.00 | skip |
| TestNodeTableIsLinkOnly | 44.25 | 44.26 | 79.10 | 79.10 | pass |
| TestShardsAgree | 269.11 | 269.11 | 257.50 | 257.50 | pass |
| TestMutants | 6.31 | 1501.58 | 7.41 | 880.42 | pass |
| TestProfileArtifacts | 56.21 | 56.21 | 76.36 | 76.36 | pass |
| TestProfileCompilation | 57.41 | 57.41 | 59.57 | 59.57 | pass |
| TestProfileSnapshotsAgree | 265.60 | 265.60 | 286.88 | 286.88 | pass |
| TestOwnedWitnesses | 80.05 | 80.05 | 27.83 | 27.83 | pass |
| TestRegistrationMutant | 54.93 | 54.93 | 29.27 | 29.27 | pass |
| TestFactoryHooks | 117.53 | 117.53 | 213.73 | 213.73 | pass |
| TestNestedOutsideModuleCopy | 0.72 | 0.72 | 0.73 | 0.73 | pass |
| TestDecodedOptionsAndMutant | 116.69 | 116.69 | 19.18 | 19.18 | pass |
| TestOptionAndComparatorGaps | 0.96 | 0.97 | 0.88 | 0.89 | pass |
| TestDecorationOptionMutant | 75.30 | 76.33 | 21.49 | 22.09 | pass |
| TestCommentFoldMutant | 65.76 | 70.04 | 22.97 | 23.79 | pass |
| TestCountGuardMutant | 77.30 | 77.30 | 23.80 | 24.43 | pass |
| TestPositionIndexMutant | 70.05 | 76.26 | 23.56 | 23.55 | pass |

Full event logs, load samples, failed 60m baseline, extraction scripts,
paired build probe source/output and typed-runtime totals are in
checker-proof/performance-*.txt and checker-proof/timing-*.txt.


Refreshed library proof: scratch merge
08ced317cfefaff83f1854be56fcba2e112e04d0 includes 6131c4c and code d85b9e68b,
with the existing one-line bridge pass-through import. Required
TestCheckerBridgeRefusalPending passed in 177.752s. The forced unsupported
question produces the same refused wire line on native, Node replay and
emitted replay. Removing the refusal handler at the same source path triggers
a fresh compilation, and the wire comparator catches it on all three;
restoring it passes again. This also proves the new content cache does not
hide same-path edits. Only the scratch worktree contains library history.
Evidence: checker-proof/performance-library-refusal.txt.
