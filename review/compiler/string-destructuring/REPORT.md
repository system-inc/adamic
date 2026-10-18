Built: plain object bindings from strings lower length as a UTF-16 property read, saving the source once.
Pins: main 87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4; TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8.
Checks: source Node, both backends, ASan/UBSan, release, leaks, counts, vet and IR readers pass; admission is +1/-0.
Mutants: wrong member, skipped source evaluation, missing diagnostic fix and a planted allocation count are caught.
Limits: only length is admitted; full scanner, full packages, full gate and WASI were not run.

## What changed

`internal/lower/collections.go` rejected the source at destructureFrom's heldAs != ir.Object check. A string source now takes a dedicated lowering path. Its holding local has the source's actual representation. Each plain length binding uses ir.StringLength over that saved string, exactly as an explicit string.length read does. Renaming and quoted static length keys work. The source is still evaluated once when length is read repeatedly or the pattern is empty. Destructured string parameters also arrive as strings in `internal/lower/functions.go`; keeping their old object signature would give the backend the wrong representation.

Conservative assumption: admit only the member found in the pinned compiler census and plain static bindings. Other named string members retain NotYet with a location and fix. Computed names, defaults, rest and nested patterns remain NotYet. Object destructuring assignments and for-of string object bindings were not extended. No backend or runtime implementation changed, and no cohere code was copied.

## Pinned TypeScript census

`census.cjs` uses stock TypeScript 6.0.3's parser and checker over the pinned src/compiler tsconfig, including constrained type parameters, unions, variable/parameter/nested/loop bindings and object assignment patterns. Node declarations were supplied from the isolated census tools directory. The complete 324 object binding patterns and 11 object assignment patterns are in `census.json`; none remains any/unknown at a binding. The assignment patterns have no string source.

Exactly one string object binding exists:

| Source | Member | Source type | Text |
|---|---|---|---|
| src/compiler/scanner.ts:1332:15 | length | string | const { length } = scanIdentifierParts() |

This is scanner.ts:778:15 in the adapted scanner named by the brief. No other string member is destructured in the pinned compiler source. The initial perf_hooks binding was unresolved without Node declarations; the final census resolves it as an object and has zero unresolved bindings.

## Fixtures and observations

`internal/oracle/testdata/string_destructuring.a` reduces the scanner call. It adds a call counter, runtime-built strings, renamed and duplicate reads, an empty pattern with effects, empty/ASCII/BMP/supplementary/lone-surrogate strings, and a destructured string parameter. `source.stdout` is the independent Node observation. The ordinary oracle compares this source against JavaScript backend Node, sanitized native and release native and checks leaks.

`unsupported-member.a` is the persistent negative witness, also covered by TestStringDestructuringMemberDiagnostic. The compiler exits 1 at unsupported-member.a:1:9 with:

```text
stage 0 can't lower string member 'charCodeAt' in object destructuring; call methods on the saved string receiver or wrap the call in an arrow yet
```

This is a NotYet refusal, rather than a claim that the language forbids all future support. Its complete path is retained in `refusal.stderr`.

New or newly registered test leaves, on the 4-CPU-quota worker, setup within the test included:

| Leaf | Seconds |
|---|---:|
| TestStringDestructuringMemberDiagnostic | 0.14 |
| TestNativeAgreesWithNode_097/string_destructuring.a, first cold run | 58.55 |
| TestCountsAreRecorded_001/string_destructuring.a | 37.98 |

The oracle cold run was uncached and overlapped other work; it is not a performance claim. No new leaf exceeded 60 seconds. The recorded counts add exactly one row: allocations 34, frees 34, retains 18, releases 46, peak 6, regions 0. Existing rows are unchanged.

## Mutants

The three compiler mutants use Go overlays, independently, without changing the working compiler. Sources have .go.txt extensions and diffs have .patch extensions. `run-mutants.py`, `mutants.json` and individual logs retain their commands and observations.

| Mutant | Catcher | Test-leaf seconds |
|---|---|---:|
| Use charCodeAt(0) instead of length | Source Node comparison fails in native and JS: 97 instead of 3; both mutant programs exit 0 with empty stderr | 15.45 |
| Omit the source declaration for an empty string binding pattern | Source Node comparison fails in both backends: calls 2 instead of 3 | 9.07 |
| Remove the unsupported-member fix | TestStringDestructuringMemberDiagnostic rejects the incomplete diagnostic | 0.24 |
| Planted recorded count is one allocation larger | TestCountsAreRecorded rejects recorded 35 versus measured 34 | 14.84 |

The two semantic mutants emit valid C and run without sanitizer findings; their kills are output disagreements, not compiler warnings or build failures. All mutant tests exit 1. The counts mutant uses the existing ADAMIC_FAMILY_PROBE mechanism and does not edit counts.md.

## Admission against main

The baseline binary is built using a Go overlay containing origin/main's exact copies of the two changed production files. All other production inputs are unchanged. Both compilers ran `c` over the same frozen 2,197 .a/.ts fixture/testdata/gap programs, dedication and examples, excluding declarations, cohere and review. This is lowering/C-emission admission, not a claim about every production module or native execution.

Initial per-program limits were 20 seconds with four probes. Twelve baseline and eight final import-heavy timeouts were resolved with 90-second probes and two workers. Final results:

| | Main | This change |
|---|---:|---:|
| Admitted | 1566 | 1567 |
| Unresolved timeouts | 0 | 0 |

The sole addition is internal/oracle/testdata/string_destructuring.a. No admission regressed. `admission-corpus.json`, before/after JSON, resolved logs and `admission-delta.json` retain the evidence. Earlier provisional timeout regressions in admission-after.log are superseded by the resolved result.

## Commands, setup and limits

All test output went to log files. Every build, test, census and probe had a hard limit. Shells sourced `/workspace/adamic-tools/env.sh`; setup was started with GOPROXY='https://proxy.golang.org|direct'.

`timeout 600 bash cloud/setup.sh` exited 124 during the repository-wide build-cache warming step. Printed timing lines: Go ready 0.193s; Node ready 0.258s; clang ready 0.985s; markdown install step-duration 1.094s; markdown ready 1.619s; submodules ready 201.171s; shared cache ready 216.529s. No done/build-cache-warm timing was printed. nproc=5, cpu.max=400000 100000 (4 CPU quota). Go 1.27.1, Node 24.19.0, clang 20.1.8. Tool readiness and the sanitizer probe had completed; the generated env and scoped cached builds were the workaround.

An initial exact-main build hit its 180-second cold compilation limit. Duplicate cold test builds were stopped before running tests. A first fixture selection used the wrong 32-bit hash, ran no oracle fixture and was excluded; the correct FNV-64 shard 097 is the successful observation. The final baseline and tip builds passed. No overlong test leaf was rerun.

Final focused commands (each redirected to its named review log):

```sh
timeout 180 go build -overlay review/compiler/string-destructuring/main-overlay.json -o /workspace/scratch/string-destructuring-main ./cmd/adamic
timeout 180 go build -o /workspace/scratch/string-destructuring-tip ./cmd/adamic
ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./internal/oracle -run '^TestNativeAgreesWithNode_097$/internal/oracle/testdata/string_destructuring.a$' -count=1 -timeout 90s -v
timeout 120 go test ./internal/oracle -run '^TestCountsAreRecorded_001$/internal/oracle/testdata/string_destructuring.a$' -count=1 -timeout 90s -v -args -update-counts
timeout 120 go vet ./internal/lower ./internal/oracle
timeout 420 python3 review/compiler/string-destructuring/run-mutants.py
ADAMIC_FAMILY_PROBE='TestCountsAreRecorded=internal/oracle/testdata/string_destructuring.a' timeout 120 go test ./internal/oracle -run '^TestCountsAreRecorded_001$/internal/oracle/testdata/string_destructuring.a$' -count=1 -timeout 90s -v
timeout 480 python3 review/compiler/string-destructuring/admission.py before /workspace/scratch/string-destructuring-main
timeout 480 python3 review/compiler/string-destructuring/admission.py after /workspace/scratch/string-destructuring-tip
timeout 480 python3 review/compiler/string-destructuring/resolve-admission.py before /workspace/scratch/string-destructuring-main
timeout 420 python3 review/compiler/string-destructuring/resolve-admission.py after /workspace/scratch/string-destructuring-tip
```

The lower diagnostic also passed in the focused lower run and the reader-package run. The reader plan is the recursive reverse-import closure over go list's production/test/external-test imports, plus both changed packages: 49 packages, 46 selected test leaves, six compile-only packages with no selected tests. Full per-package commands, exit codes and wall times are recorded in the reader JSON and logs. No whole package or full gate was run. IR TestCallTargetReaders ran with -timeout 90s and passed in 54.60s, including its exports product setup. go vet on both changed packages passed. gofmt -l and git diff --check produced no output.

The broad CSS TestCompositionMatchesGo_Setup exceeded the 120-second command limit while making cold composition products and was not rerun. TestClosedEmptyArrayUnionGap is the bounded replacement reader witness and passes in 2.97s. The failed setup remains in its original log. This is a scoped reader check, not a green claim for that broad CSS setup.

The first reader batch reached its 600-second command limit after 31 package results. Only unfinished packages were run in the follow-up batch; successful leaves and the overlong CSS setup were not rerun. Reader completion is recorded in final-reader-results.json. Remaining behavior outside this unit: string methods as detached bindings, numeric/computed members, defaults/rest/nested string patterns, string object assignment and for-of patterns, namespaces and the complete native scanner. No full scanner output or scanner speed claim is made.

Census reproduction (source/setup logs and dependency versions are retained):

```sh
timeout 120 git clone --depth 1 --branch v6.0.3 --filter=blob:none --sparse https://github.com/microsoft/TypeScript.git /workspace/scratch/string-destructuring-typescript
timeout 90 git -C /workspace/scratch/string-destructuring-typescript sparse-checkout set src/compiler src/_namespaces src/compat
timeout 120 npm install --prefix /workspace/scratch/string-destructuring-tools --no-audit --no-fund typescript@6.0.3
timeout 90 npm install --prefix /workspace/scratch/string-destructuring-tools --no-audit --no-fund @types/node
CENSUS_TYPESCRIPT=/workspace/scratch/string-destructuring-tools/node_modules/typescript timeout 90 node review/compiler/string-destructuring/census.cjs /workspace/scratch/string-destructuring-typescript
timeout 600 python3 review/compiler/string-destructuring/run-readers.py
timeout 120 go test ./stage1/cohere/css -run '^TestClosedEmptyArrayUnionGap$' -count=1 -timeout 90s -v
timeout 240 python3 review/compiler/string-destructuring/run-remaining-readers.py
```

The scanner TestScannerShardCoverage failed at corpus setup: its child git clone could not connect to github.com:443. Its failure log is preserved. The local TestGapStandsWhereGapsMdSays lowering witness is the bounded replacement, with its own log. This does not claim corpus coverage passed. The first green push exceeded the requested 30-minute deadline: checkout began at 11:13:31 UTC; cold setup/cache warming and the bounded reader batches delayed completion beyond 11:43:31 UTC. No partial state was pushed as green.

```sh
timeout 120 go test ./stage1/typescript/scanner -run '^TestGapStandsWhereGapsMdSays$' -count=1 -timeout 90s -v
python3 review/compiler/string-destructuring/finalize-readers.py
```
