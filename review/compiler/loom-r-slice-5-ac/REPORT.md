Added 13 missing first-line a-check refusal headers; every program body is byte-identical to 505835b7.
Delivery is one commit on dbdf977f5b8b0979d414ebe6f20e7aadd312e309 containing 505835b74876b61595864dcc5ddac6d40ba3767a plus these headers and review evidence; its SHA is in the handoff.
Actual branch compiler and unchanged pinned Gate.aCheck accept 58/58: 42 checked, 0 NotYet, 16 refused, 0 type error; oracle header tests, reader guard and vet pass.
Removed-header and wrong-reason mutants both exit 1 through the actual a-check predicate; the source file is restored and the final scan passes.
Not full-branch green: 13 inherited stale freshness readers and their coverage guard fail. No full gate or prior unrelated consumer/budget rerun is claimed.

The explicit pinned base overrides the current-main boilerplate. origin/main advanced to 99c369e0 during fetch; this branch remains based exactly on dbdf977f5b. The conservative scope is the requested source change plus headers: no compiler/runtime/Go-test edits, no fixture moves, no unrelated test exclusions, and no counts changes relative to 505835b7. No cohere code was copied. No ahra command, PR or Loom submission was made.

Inventory and real predicate

paths.txt is every added, modified, copied or renamed .a in the change, derived from git diff --name-only --diff-filter=ACMR dbdf977f5b 505835b7 -- "*.a". It includes all 13 relocated refusal controls. There are no added or modified review/ .a files. Deleted old paths are not programs in the resulting tree. The final diff adds no other .a files.

check-headers.py invokes unchanged Gate.aCheck from README-pinned commit fbac28c62493f27a788edc02a18bc8edb68de5da. gate-source.py is the exact fetched file, SHA-256 c81d5f29ae6a06c5671fd8caf961828bc040c1f565df0c0f2cb660ce28a29317. Its own build step runs go build -o <tree>-binaries/adamic-acheck ./cmd/adamic on this working tree, then adamic-acheck c <path> for every program. The harness supplies logging and hard time limits without changing the predicate. This measures the checker/lowering/C-emission front end, not clang execution. before.json and final.json retain full real diagnostics, status, exit, duration and emitted-C size/hash.

The pre-header predicate rejected exactly the 13 missing headers. After insertion it exits 0 for all 58. Three existing refusal headers were already accepted and remain unchanged. All 42 checked programs have no expected-error header; there are no NotYet or type-error cases.

| File | Status | Refusal reason / checker code | Header added |
|---|---|---|---|
| `internal/lower/testdata/placeholders/nonliteral.a` | refused | the non-null assertion ! | no |
| `internal/oracle/testdata/4ddd17f_weak_single_narrowed.a` | checked |  | no |
| `internal/oracle/testdata/backend_stop_catches.a` | checked |  | no |
| `internal/oracle/testdata/d96d304_try_stack.a` | checked |  | no |
| `internal/oracle/testdata/non_null_refused/non_null_possible_read.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_possible_typeerror.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_argument.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_checked.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_comparison.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_conditional.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_element.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_expression.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_logical.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_null.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_return.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_spread.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/non_null_refused/non_null_refuse_undefined.a` | refused | the non-null assertion ! | yes |
| `internal/oracle/testdata/placeholder_nonnull_alias_reset.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_assignment_result.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_before_use.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_field.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_local.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_before_use.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_coalesce.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_copy.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_equality.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_json.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_loose.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_optional.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_truthiness.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_null_typeof.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_observe.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_return_assignment.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_saved_leak.a` | checked |  | no |
| `internal/oracle/testdata/placeholder_nonnull_scanner.a` | checked |  | no |
| `stage3/parser-ahead/factories/already-complete.a` | checked |  | no |
| `stage3/parser-ahead/factories/complete.a` | checked |  | no |
| `stage3/parser-ahead/factories/escaped.a` | checked |  | no |
| `stage3/parser-ahead/factories/scout.a` | refused | a cast the runtime can't check | no |
| `stage3/parser-ahead/node-array/fields.a` | checked |  | no |
| `stage3/parser-ahead/node-array/presence.a` | refused | an unproven relation from { kind: number | no |
| `stage3/parser-ahead/rulings/factory-unset.a` | checked |  | no |
| `stage3/parser-ahead/rulings/factory-use.a` | checked |  | no |
| `stage3/parser-ahead/rulings/node-array-json.a` | checked |  | no |
| `stage3/parser-ahead/rulings/node-array-keys.a` | checked |  | no |
| `stage3/parser-ahead/rulings/node-array-length.a` | checked |  | no |
| `stage3/parser-ahead/rulings/node-array-unset.a` | checked |  | no |
| `stage3/parser-ahead/rulings/null-placeholder-pending.a` | checked |  | no |
| `stage3/parser-ahead/rulings/own-key-order.a` | checked |  | no |
| `stage3/parser-next/speculation/arrow-tryparse.a` | checked |  | no |
| `stage3/parser-next/speculation/misfit.a` | checked |  | no |
| `stage3/parser-next/speculation/named-lookahead.a` | checked |  | no |
| `stage3/parser-next/speculation/named-tryparse.a` | checked |  | no |
| `stage3/parser-next/speculation/truthiness-rewind.a` | checked |  | no |
| `stage3/parser-next/stack/benchmark.a` | checked |  | no |
| `stage3/parser-next/stack/legal.a` | checked |  | no |
| `stage3/parser-next/stack/pathological.a` | checked |  | no |

Headers added

Each of the following 13 files gains exactly `// a-check: refused the non-null assertion !` as its first line. No type-error header was necessary. headers.json records every exact header and original body hash; headers.patch is the zero-context insertion diff.

- `internal/oracle/testdata/non_null_refused/non_null_possible_read.a`
- `internal/oracle/testdata/non_null_refused/non_null_possible_typeerror.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_argument.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_checked.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_comparison.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_conditional.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_element.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_expression.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_logical.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_null.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_return.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_spread.a`
- `internal/oracle/testdata/non_null_refused/non_null_refuse_undefined.a`

Body bytes, line positions and counts

body-proof.json and final-integrity.json compare the entire post-header body against git show 505835b7:<path>, not normalized text. All 13 comparisons are byte-for-byte equal. Header insertion moves each refusal location by one line, preserving its column and reason; diagnostic-shifts.json records every old/new location.

The required exact-path/basename/family grep over *_test.go and testdata is retained in path-references.txt. Active Oracle readers are TestImpossibleNonNullFixturesAreRefused, TestPossibleNonNullAdamicAssertionsAreRefused and TestCheckedNonNullAdamicRefusal. Their .a cases check refusal kind, reason, fix and suffix without pinning line positions. The :1:13 pins belong to untouched .ts controls; the :47:16 pin belongs to untouched non_null_refused/initialized.a. No active positional expectation needed an edit. Historical source reports and captures remain observations of their original revision; fresh post-header diagnostics are in final.json.

internal/oracle/counts.md is byte-identical to 505835b7, SHA-256 1c750ef4bdddda92d0e3314944234fe2fd7c89db486424720222a34b9b9219d8. No counts regeneration was run: only refused, uncounted controls gained comments. TestCountsAreRecordedTable passes.

Header mutants

prove-headers.py edits non_null_refused/non_null_possible_read.a in place, runs the actual compiler and pinned Gate.aCheck on each mutant, and restores the exact original bytes in finally. No intentionally headerless .a mutant is left in review evidence.

| Mutant | Actual result | Predicate catcher |
|---|---|---|
| Remove first-line header | Compiler still refuses the non-null assertion; a-check exits 1 | refused, expected checked |
| Replace reason with deliberately wrong slice5 header reason | Compiler still refuses the non-null assertion; a-check exits 1 | refused, expected refused deliberately wrong slice5 header reason |

header-mutants.json contains commands and exact failure strings; removed-header.json and wrong-reason.json contain real diagnostics. The final complete scan runs after both restorations and accepts every file.

Tests and commands

All test output goes directly to logs. GOMAXPROCS=4 and GOFLAGS preserve -buildvcs=false -trimpath and append -p=4; GOCACHEPROG is unset for the local Go cache. Every go test has -count=1 -json -timeout 90s and an outer 150s or 180s limit. A-check compiler invocations have 45s limits; build steps have 120s limits, with an outer 240s script limit. No test leaf added or edited in this pass; observed affected-test and guard times are in test-seconds.json.

```sh
python3 review/compiler/loom-r-slice-5-ac/check-headers.py before
python3 review/compiler/loom-r-slice-5-ac/prove-headers.py
python3 review/compiler/loom-r-slice-5-ac/check-headers.py final
go test ./internal/oracle ./internal/ir -run '^(TestImpossibleNonNullFixturesAreRefused|TestPossibleNonNullAdamicAssertionsAreRefused|TestCheckedNonNullAdamicRefusal|TestCallTargetReaders)$' -count=1 -json -timeout 90s
go test ./internal/oracle -run '^(TestImpossibleNonNullFixturesAreRefused|TestPossibleNonNullAdamicAssertionsAreRefused|TestCheckedNonNullAdamicRefusal|TestCountsAreRecordedTable)$' -count=1 -json -timeout 90s
go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle
```

Both Oracle/reader commands and vet exit 0. The final Oracle command, after mutant restoration, completes in 1.596s. TestCallTargetReaders passes in 2.74s. The initial Oracle .ts controls also run source Node and ASan/UBSan native observations; their largest observed leaf is 8.26s. No output comparison or diagnostic expectation was weakened.

Inherited freshness failure found by the path audit

The prior relocation left 13 internal/fresh generated readers naming ../oracle/testdata/non_null_*.a instead of the files now under non_null_refused/. All 13 selected readers fail loading absent files; TestFreshCorpusUnitsCoverEveryProgram fails at generated index 550 for ../oracle/testdata/non_null_possible_read.a. The focused command exits 1 in 0.007s. These are not header-mutant kills and are not claimed green:

```sh
go test ./internal/fresh -run '^(TestFreshCorpusUnitsCoverEveryProgram|TestFreshWrites____oracle_testdata_non_null_(possible_(read|typeerror)|refuse_(argument|checked|comparison|conditional|element|expression|logical|null|return|spread|undefined))_a_[0-9a-f]+)$' -count=1 -json -timeout 90s
```

inherited-freshness.json proves every old path was already absent in 505835b7 and that fresh_test.go, corpus_units_test.go and corpus_coverage_test.go are unchanged from that revision. reader-failures.json and logs/fresh-path-readers.jsonl retain the exact failures. This pass does not modify another package's generated tests or suppress their coverage to force green. A freshness reader relocation update remains necessary before claiming the full branch is green.

Setup and limits

The initial source application failed with No space left on device while creating internal/lower/testdata/placeholders; review output could not be created, so setup did not start. Removed only the disposable local product cache /home/agent/.cache/adamic-build/local (about 11 GB), restored the exact requested 505835b7 tree, and verified an empty restoration.txt diff before any header edit. Setup then succeeded. Source env: /workspace/adamic-tools/env.sh; GOPROXY=https://proxy.golang.org|direct; nproc=5; cpu.max=400000 100000, a four-CPU quota. Go 1.27.1, Node 24.19.0 and house clang 20.1.8 on Linux. Timing lines:

- setup: go ready (0.068s)
- setup: submodules ready (0.071s)
- setup: node ready (0.078s)
- setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.012s
- setup: markdown dependencies ready (0.207s)
- setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.457s)
- setup: shared cache off (ADAMIC_GOCACHE_OFF=1) (0.459s)
- setup: go build ready (17.055s)
- setup: test binaries deferred (use --warm-tests) (17.238s)
- setup: build cache warm (17.239s)
- setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=off load-before=0.00 0.00 0.00 1/136 922 load-after=0.80 0.18 0.06 1/141 1396
- setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (17.263s)
- setup: source /workspace/adamic-tools/env.sh
- setup: logs /tmp/adamic-gate/setup.KhxjxK

Not covered: whole packages, the full fast gate, ahra/Loom submission, a whole-tree scan beyond the requested changed .a inventory, or a repeat of the original compiler change's unrelated consumer suites and recorded semantic mutants. This pass adds no semantic compiler/runtime change; 505835b7's existing validation limitations and runtime-clearance report still apply. Its over-budget consumer tests were not rerun. The header phase is green; the inherited freshness failures above prevent full-branch clearance.
