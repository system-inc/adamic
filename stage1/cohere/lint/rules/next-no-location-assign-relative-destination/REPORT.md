# Next.js syntax unit: one port, nine measured blockers

This is a partial unit. `@next/next/no-location-assign-relative-destination`
is implemented, registered by directory discovery, and held byte for byte to
unchanged Go cohere on Node and sanitized native. The other nine claimed rules
are unimplemented. [GAPS.md](GAPS.md) records seven JSX parser refusals and two
filename-sensitive verification blockers with executable proving programs.
No blocked listener is registered as a clean stub.

Branch `codex/stage1-nextjs-lint` starts at refreshed main `d090af5`, merged
with registration tip `48ecd9302bf3954a4ddbbd14c28ba09148c1c1a8`.
Claim commit `79dff93ac3df72863bcfb475c11917e82f4b6aa6` was pushed before
implementation. Batch 6 `c5d128a` and batch 8 `4189abd` claims were checked;
their rules are disjoint from the ten names in [the claim](../../claims/nextjs.md).
The helper tip `29990b4` was inspected and not merged: it supplies
options/schema/message helpers, not the missing JSX parser. Its readiness
metadata marks every selected Next.js rule false; the excerpt is retained.

## Implemented behavior

The location listener preserves Go's global receiver names, computed string
properties, optional calls, assignment operator set, scope/shadowing decisions,
import binding shapes, parameters/catch bindings, and navigation-wide finding
spans. Destination resolution uses cooked literal/template prefixes, the left
side of concatenation, nearest declaration/preceding write, and an explicit
cycle guard. The absolute-destination expression uses Adamic's native regular
expressions with the same case-insensitive pattern as Go. Scope links are numeric
parent indexes; there is no owning AST cycle or type checker.

The port follows Go's decisions, including Go's stated divergences from the
original Next.js rule's static evaluator. It offers no fix or suggestion,
matching Go. Its descriptor, rule, exact message, Go adapter, witnesses and
mutant are all in this directory. No shared dispatch, oracle, corpus, compiler
or runtime file was changed; generated registries are ignored.

## Observed parity

The fixture gate captured 316 distinct source/rule/options combinations:
217 inherited cases and 99 location cases. Together with 24 inherited generated
runs, **340 cases** agree on Go, Node and ASan/UBSan native, **159,249 bytes**.
The output includes rendered messages, IDs, UTF-8 ranges, repair metadata and
entire fixed sources. This rule has no repair payload to omit.

There are 44 location witnesses: the minimal upstream navigation plus 43 generated
edge sources. Each has a real Go finding. Edges pair an unshadowed globalThis
positive control with relative/absolute, scope, import, catch, binding cycle,
Unicode/CRLF, escaped strings/keys/templates, optional calls, computed
keys and all tested assignment spellings. The shared owned-witness test runs
all 49 registered witnesses selected and with all rules: **98 runs**, **94,622
identical bytes**. The counts include the inherited five witnesses.

The final source-corpus gate compares **197 files**, **12,484,673 identical
bytes**, across Go, source on Node and sanitized native. The external TypeScript
corpus is v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`, 77 compiler
files; the remaining files are stage1 sources. The cohere submodule remains
`715ba94f3608a6500086b1076ce5cb7e51b836db`, unchanged.

Separately, original Go assertions ran for all ten selected rules. The capture
retained **315 filename/source/rule/options cases**, with positive findings in
every family. These are oracle observations, **not 315 native parity cases**.
The multi-file route tests also ran, but their custom assertions are not all
captured. Seven shortest captured Go-positive JSX sources are retained with
exact matching Node/native parser refusals and a clean non-JSX control. The
filename probe records one finding at each original filename and zero after
renaming, for both document import and Pages Router typo listeners.

## Setup and commands

`bash cloud/setup.sh > /tmp/nextjs-setup.log 2>&1` succeeded. `nproc` is 5,
with four CPUs of cgroup quota and 17.6 GB memory. Its timing lines were:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (100s)
setup: done in 100s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Toolchain commands source `/workspace/adamic-tools/env.sh`. Go 1.27.1,
clang 20.1.8, Node 24.19.0, Linux x86_64. Test outputs went to regular files.

```sh
ASAN_OPTIONS=detect_leaks=1 UBSAN_OPTIONS=halt_on_error=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_LINT_BENCH=1 go test ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses|TestCompilerAndStage1Agree|TestThroughput|TestMutants/next-location-relative-guard' -count=1 -v -timeout 20m > /tmp/nextjs-verification.log 2>&1
ASAN_OPTIONS=detect_leaks=1 UBSAN_OPTIONS=halt_on_error=1 go test ./stage1/cohere/lint -run 'TestOwnedWitnesses|TestMutants/next-location-relative-guard' -count=1 -v -timeout 20m > /tmp/nextjs-edge-verification.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestTheOracleCatchesOneByte -count=1 -v -timeout 10m > /tmp/nextjs-external-oracle.log 2>&1
go vet ./... > /tmp/nextjs-vet-final.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/rules/next-no-location-assign-relative-destination > /tmp/nextjs-gofmt.log
/workspace/scratch/cohere --no-fix --no-cache stage1/cohere/lint/rules/next-no-location-assign-relative-destination > /tmp/nextjs-source-gate-final.log 2>&1
```

These passed. Fixture/corpus/mutant/throughput gate: 101.203s; corrected edge
and mutant gate: 31.629s. The external one-byte mutant passed in 6.579s, with
native and Node cache misses. Vet and gofmt printed nothing. The source gate
reported three checked modules, all Adamic-ready, with no findings.
The complete touched-package gate also passed: lint 225.219s; registry 0.034s.
Its command was:

```sh
ASAN_OPTIONS=detect_leaks=1 UBSAN_OPTIONS=halt_on_error=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint/... -count=1 -v -timeout 20m > /tmp/nextjs-lint-gate.log 2>&1
```

[Complete gate](evidence/lint-gate.log) retains every test and mutant result.
The count benchmark is intentionally skipped in this full run; its completed
five-round observations are in [fixture-parity.log](evidence/fixture-parity.log).
No complete repository-wide `go test ./...` is claimed.

## Mutants and throughput

The location family mutant reverses `relative(...)`'s reporting guard. It compiles
and executes successfully on both Node and sanitized native, then fails Go byte
parity by suppressing the rooted navigation finding. The corrected edge gate
also executes its successfully compiled mutant against all owned witnesses.
The final complete gate caught all six registered-rule behavior mutants on
both source Node and sanitized native, after successful compilation/execution:

| Rule/check          | Mutation                          | Observation that caught it          |
| ------------------- | --------------------------------- | ----------------------------------- |
| Location navigation | Invert relative destination guard | Missing rooted navigation finding   |
| no-debugger         | Suppress fix                      | Repair kind and fixed source differ |
| no-empty            | Report empty function body        | Extra finding                       |
| eqeqeq              | Apply suggestion as automatic fix | Repair kind and fixed source differ |
| no-var              | Suppress declaration              | Missing finding                     |
| no-duplicate-case   | Suppress repeated case            | Missing finding                     |

The infrastructure controls also passed on both backends: a valid wrong
subscription removes the debugger listener; omitting finish changes the hook
sequence; ignoring decoded AllowEmptyCatch adds a finding. A separate nested
module-copy mutant is caught by Node resolution and native module loading;
that is an instrument check, not a successfully executing behavior mutant.
Ten invalid descriptor mutations are rejected by discovery: unknown node kind,
duplicate name, missing visit export, no interested kinds, unknown field,
unsafe name, missing finish export, missing class, missing oracle export and
missing factory. These are expected validation failures, not rule mutant kills. No refusal or compile failure counts as killing
the location mutant. The parser's clean control validates the refusal instrument;
the external uncached oracle's one-byte mutant validates outside-compiler parity.

Five interleaved fresh-process count rounds over the 77 compiler files, all six
registered rules, gave the following best-of-five measurements:

| Backend        |  Seconds | Findings/second |
| -------------- | -------: | --------------: |
| Go cohere      | 0.266465 |        1,692.53 |
| Native release | 1.370375 |          329.11 |
| Source on Node | 0.920522 |          489.94 |

Each counts 451 findings. They are the inherited rules' findings; this compiler
corpus contains no location-navigation finding. These timings measure the
six-rule runner, not a Next.js-heavy workload or a speedup. Startup, reads,
parsing and traversal are included; compilation, rendering and fixing are
excluded. Correctness gates use sanitizers; this timing binary uses release
optimization.

## Limits and development corrections

Nine selected rules are not implemented or held to parity. No Next.js application
checkout, binder, suppression/configuration layer, general JSX AST, filesystem
route model or parser recovery is claimed. The raw Go corpus and blocker ledger
are evidence for the next unit, not listener parity through a Go-supplied AST.
The shared registration harness's TS-only oracle and lost filenames must be fixed
by its owner before registering path-dependent rules faithfully.

A source-gate run identified a constant RegExp constructor, mutation of a visited
array argument and an unused initialization. The final implementation uses a
literal, an immutable visited path and definite assignment. An early generated
witness put its positive control before a local declaration of `location`; that
declaration shadows the whole file, so Go correctly reported no finding. The
control now uses unshadowed `globalThis.location`, and the corrected complete
witness comparison passes. Earlier runs are not reported as final evidence.
