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
