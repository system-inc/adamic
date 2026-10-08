# Wave 30 port on checker facts

Base: d845dccde413c89643293e808626344d12e3f023, lint-checker/facts. This branch is lint-rules/facts-wave-30. No legacy bridge files, private checker, shared helper, harness or registration lists are changed.

## Completed implementation

nexus/consistency-no-iso-string-date-cut is implemented with supplied-node listeners for CallExpression, ElementAccessExpression and VariableDeclaration. Its single module uses the shared SymbolDetails decoder and the harness checker. Local declaration spans correlate against this file only. Date member provenance is queried through askFile(ReadsDefaultLibrary). The descriptor declares exactly upstream ReadsCompilerOptions and ReadsDefaultLibrary; it does not add ReadsOtherFiles. Findings have the exact upstream message and no fixes or suggestions.

## Claims stopped or already completed

| Rule | Exact upstream call/location | Missing question/helper |
| --- | --- | --- |
| `nexus/correctness-no-callback-in-parse-try` | `checker.SkipAlias(symbol, ctx.TypeChecker)` at `cohere/internal/lint/rules/nexus/correctness_no_callback_in_parse_try.go:247` | alias-declarations / alias-followed-symbol-details |
| `nexus/correctness-no-collection-misuse` | `part.AsLiteralType().Value()` at `cohere/internal/lint/rules/nexus/correctness_no_collection_misuse.go:258` | literal-value for a type identity, preserving string contents |
| `nexus/correctness-no-discarded-outcome` | `checker.Checker_getAwaitedType(ctx.TypeChecker, valueType)` at `cohere/internal/lint/rules/nexus/correctness_no_discarded_outcome.go:135` | awaited-shape and type-declaration ancestry/union arm identity |
| `nexus/correctness-no-process-exit-after-output` | `ctx.TypeChecker.GetAliasedSymbol(symbol)` at `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:413` | alias-declarations; also resolved signature declaration/body and selected return type at :314/:342 |
| `nexus/correctness-require-blocking-standard-streams` | `analysis.ctx.TypeChecker.GetAliasedSymbol(symbol)` at `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:586` | alias-declarations; also program runtime module graph and foreign function bodies |
| `react/jsx-fragments` | `declaration.AsVariableDeclaration().Initializer` at `cohere/internal/lint/rules/react/jsx_fragments.go:337` | foreign declaration initializer syntax; no ReadsOtherFiles is declared by this upstream rule |
| `react/jsx-no-constructed-context-values` | `walk.ctx.TypeChecker.GetResolvedSignature(call)` at `cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:488` | resolved signature declaration/body; native shared construction/stability/escape analysis remains absent |
| `react-hooks/purity` | `high_level_intermediate_representation.ForFunction(ctx, functionNode)` at `cohere/internal/lint/rules/react/purity.go:250` | shared native React HIR, SSA and render/capture analysis |
| `react-hooks/refs` | `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)` at `cohere/internal/lint/rules/react/refs.go:184` | shared native React HIR, SSA and capture analysis |
| `react-hooks/preserve-manual-memoization` | `hir.CloneFunction(hir.ForFunction(ctx, functionNode))` at `cohere/internal/lint/rules/react/preserve_manual_memoization.go:195` | shared native React HIR/SSA/reactive scopes; hir.AnalyzePreservedManualMemoization at :199 |

Already completed by wave 1 according to Ahra, so not ported again: nexus/correctness-no-uncleared-race-timeout, nexus/correctness-no-discarded-pure-result, react/jsx-no-undef. Their upstream cases and mutants were not rerun as owned work in this unit.

The remaining JSX fragment blocker concerns a symbol whose variable declaration is in another file. Complete symbol records expose its span, but not its initializer AST. Same-file initializer lookup does not prove full rule behavior for that case. Opening/parsing the foreign file, adding a private question, or declaring ReadsOtherFiles where upstream declares none would violate the unit requirements. No partial registered rule is installed.

## Validation

Implementation commit: 8d025b1849dc0afabad998b21854b43a414fbf4c. Validation changes only add this report and evidence; the tested rule code is unchanged.

- `go run ./cmd/lint-registry`, `go run ./cmd/adamic types stage1/cohere/lint/main.ts`, `go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry`, gofmt and git diff whitespace checks pass. Logs are in `evidence/`.
- The lint package ran once: `go test -json -count=1 -timeout=3h ./stage1/cohere/lint`. All compiler, benchmark, profile compilation and profile snapshot inputs were supplied; exact environment is in `evidence/gate-inputs.json`.
- PASS: 34 top-level tests; FAIL: 0; SKIP: 1. Including subtests: 119 pass, 0 fail, 1 skip. The sole inherited skip is `TestCheckerBridgeRefusalPending`, awaiting codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer. No input-dependent check skipped, and no shared check was edited or bypassed.
- Wall time: 2210.150 seconds (36m50s); Go package time: 2204.356 seconds. nproc: 5; CPU quota and GOMAXPROCS: 4. Load average at start: 3.140/3.755/1.730; at end: 1.689/2.012/2.301. Sampled one-minute load range: 1.000 to 6.982. Full output and sampled load are in `evidence/lint.jsonl.gz` and `evidence/load.jsonl.gz`; counts are in `evidence/gate-summary.json`.
- ISO upstream prefix `TestConsistencyNoIsoStringDateCut` captures all six test families. All 44 unique rule/file/options/source cases matched Go across Node, emitted JavaScript and ASan/UBSan native, including exact findings, fixed source and suggestions. The harness captured 88 records; raw records and counts are in `evidence/iso-capture.jsonl.gz` and `evidence/upstream-counts.json`.
- The owned witness fires in Go and matches all three ports. `TestOwnedWitnesses` passes in 9.18 seconds. The mutant `Date declaration identity inverted` compiles and runs and is caught only by comparison with Go on native, Node and emitted JavaScript; its subtest passes in 10.81 seconds. All three catch lines are in `evidence/mutant-lines.txt`. All 81 registered mutants pass.
- `TestCompilerAndStage1Agree` passes over 883 compiler/stage1 files in 406.83 seconds. Shard comparison and release/profile snapshot comparison pass. Sanitized native is exercised by upstream, witness and mutation comparisons; the full log preserves the allocator/release checks and explicit recovery boundaries.
- Full-inventory throughput (77 files, 28,312 findings, best of five) is Go 1.552850 seconds versus native 12.336956 seconds, 7.94x slower. This is the harness inventory benchmark, not an isolated ISO-rule measurement.
- `bash cloud/setup.sh` succeeds. Timing lines: Node 0.024s; Go 0.024s; submodules 0.060s; verified markdown tool step 0.008s / ready 0.075s; clang 0.221s; Go build 217.475s; cache warm 217.716s; total 217.749s. Full setup output is in `evidence/setup.log.gz`.

## Released handles and bridge canaries

`ADAMIC_TSGO_CORPUS=/workspace/wave-30-typescript go test -v -count=1 -timeout=3h ./bridge/tsgo -run '^TestBridge$'` passes. The C ABI rejects stale and zero handles, preserves returned outputs across release, and distinguishes subsequent program handles. The released-handle mutant is caught by the stale-handle assertion; input/output length mutants are caught by ASan. Native and Go produce 54,982 identical bytes for 1,600 positions across four compiler files under ASan/UBSan/LSan. Complete output is in `evidence/bridge.log.gz`.

## Work stopped

The ten rules in `blocked-rules.json` were stopped before implementation: each has zero matched upstream cases and no mutant run in this unit. Their exact upstream calls and absent questions/helpers are listed above. No shared checker or analysis implementation was copied or modified. The three rules already completed by wave 1 were not duplicated.
