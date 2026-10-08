# Wave 23 unified checker landing

This branch merged `origin/area/stage1-lint` at `c4bdc23fa86d55cf7e579989201c11258f4d3a62` into `codex/typeaware-wave-23` with merge commit `9df9ea102598ac939a5a6a9b304f233854661ce0`. Both parents and their tests remain. No rebase, new claim, main push or area push occurred. `codex/lint-port-max-nested-callbacks` at `ca753eaaec95aa923e48e79aa5946fe0102e04ea` is already an ancestor of the fetched area; it needs no new merge.

## Active unified ports

Thirteen descriptor directories now use the area's one checker per run: no-eval, no-extend-native, no-func-assign, no-new-func, no-new-native-nonconstructor, no-new-wrappers, no-throw-literal, no-useless-backreference, prefer-arrow-callback, nexus/correctness-no-collection-misuse, nexus/correctness-no-discarded-pure-result, react/jsx-fragments and react/jsx-no-undef. Each owns its `.a` entry and engine, descriptor, unchanged-production Go adapter, firing witness and compiled verdict-omission mutant. Existing messages are retained verbatim. Option adapters decode upstream options rather than returning nil. Upstream prefixes were checked against the current cohere Test names.

Each owned `facts.a` adapts the existing fact readers and diagnostic payloads to RuleContext. Its virtual `ask` uses only `context.checker.ask`; it does not open another program or provide private recorded answers. Native records the area's answers; Node and emitted JavaScript consume and validate the shared replay transcript. UTF-8 byte ranges are converted to UTF-16 when building findings and edits. Fixes and suggestions preserve order and payloads. Supplied-node visits run through declared kind buckets; file-level listeners retain upstream's SourceFile subscription. No shared registration generator, harness or RuleContext code was edited.

The old standalone suites remain unchanged as checks, with compiler-host and rooted-path API compatibility updates confined to wave-23-owned Go files. They still use production Go rules independently, compare native and sanitizer output, run compiled mutants, assert released-handle refusal and run both frozen corpora. This extra legacy coverage is distinct from the unified replay certification.

## Explicit blockers

| Rule | Exact dependency and reproducer |
| --- | --- |
| @typescript-eslint/only-throw-error | Missing native shared `type_checking.TypeMatchesSomeSpecifier`, `cohere/internal/lint/checking/specifier.go:246`, called from `cohere/internal/lint/rules/typescript/only_throw_error.go:265`. Reproducer: `declare const m: Map<string, string> \| Set<number>; throw m;` with Allow `{From:"lib",Name:["Map"]}` and AllowThrowingAny/Unknown false. All union members must match; the related intersection must allow any matching member. The existing default-only port cannot certify these options. |
| @typescript-eslint/prefer-promise-reject-errors | The same helper, called from `cohere/internal/lint/rules/typescript/prefer_promise_reject_errors.go:198`. Reproducer: `Promise.reject(new Map<string,string>());` with Allow `{From:"lib",Name:["Map"]}`. Missing allow/from file/lib/package and inline-specifier support blocks activation, rather than an adapter silently ignoring options. |
| @typescript-eslint/prefer-reduce-type-parameter | `PreferReduceTypeParameter` fix construction, `cohere/internal/lint/rules/typescript/prefer_reduce_type_parameter.go:207`, includes a zero-width empty removal. Reproducer: `[1].reduce((a,b)=>a+b,0 as number);`. Shared `stage1/cohere/lint/lint.ts:175` rejects it as `nonprogressing fix`. Completed unified code and pending descriptor are in `stage1/cohere/lint/pending-wave23/typescript-prefer-reduce-type-parameter/`. No Go edit was dropped and no guard relaxed. |
| nexus/correctness-no-discarded-outcome | Missing multi-file upstream capture/project helper. `correctnessNoDiscardedOutcomeRun`, `cohere/internal/lint/rules/nexus/correctness_no_discarded_outcome_test.go:83`, calls `RunTypedFiles` at line 89 with four Nexus source files. Shared `stage1/cohere/lint/lint_test.go:378` reopens only the captured subject file, without its virtual dependency files. Reproducer: the upstream prelude import of `parseJson` from `../../libraries/structure/libraries/nexus/source/structured-text/json/Json`, followed by `parseJson(text, "context");`, with its real Json.ts union alias at the upstream Nexus path. Omitting those dependency files can falsely agree on no findings. The original multi-file suite remains active. |
| react-hooks/set-state-in-effect | Missing native HIR lowering and capture analysis: `high_level_intermediate_representation.ForFunctionWithoutManualMemoization`, `cohere/internal/lint/rules/react/set_state_in_effect.go:267`. Existing Go controls, capability inputs and named-kind declarations remain parked; no native findings parity is claimed. |
| react-hooks/set-state-in-render | Missing native HIR lowering and SSA adapter: `high_level_intermediate_representation.ForFunction`, `cohere/internal/lint/rules/react/set_state_in_render.go:183`, and `UnconditionalBlocks` at line 260. Original parked evidence retained. |
| react-hooks/static-components | Missing native HIR lowering and capture analysis: `high_level_intermediate_representation.ForFunction`, `cohere/internal/lint/rules/react/static_components.go:120`. Original parked evidence retained. |
| react/jsx-no-constructed-context-values | Native capture/escape gap: `jsxNoConstructedContextValuesStabilityWalk.functionEvaluation`, `cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:532`; `anyEscapes` at 603 and `jsxNoConstructedContextValuesStaysHome` at 649. Original fresh-versus-cached Go controls and capability inputs remain parked. |

## Validation

Final machine, commands, counts, mutants and wall times are recorded alongside the event logs in `validation-wave23-unified/`. Required setup: Go/clang/Node/submodules 0 seconds, cache warm 36 seconds, total 36 seconds; nproc 5, cgroup quota four cores. The TypeScript corpus is clean v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Legacy manifests preserve the frozen 77 compiler and 287 repository files, using absolute paths so the compiler root is not accidentally prefixed twice.

A detached runner executes retained suites first and the full lint package second, with `GOMAXPROCS=4`, `ADAMIC_TYPESCRIPT_SOURCE`, `ADAMIC_LINT_BENCH=1`, and both profile inputs set to a fresh scratch directory. No test or input is disabled. The area's explicit `TestCheckerBridgeRefusalPending` dependency on TSGoError is reported separately if it skips. Initial failures and interrupted runs are retained as evidence, not counted as completed gates.

The area contains standalone `stage1/cohere/static_single_assignment/` algorithms. The parked React dependencies are the missing high-level-IR lowering/cache and rule analysis integration, not absence of those standalone graph algorithms. The retained listener-declaration check caught the changed no-func-assign Go subscription; its owned named and numeric declarations now match FunctionExpression (220) and FunctionDeclaration (264), without relaxing the comparison or removing any metadata mutants.

The same retained probe caught numeric ast.Kind drift after the checker pin change (for example ThrowStatement 258 became 259). All eighteen historical numeric declarations were refreshed from the current generated ast.Kind definitions, and remain independently compared with live production Go registrations and their compiled mutants. New unified descriptors continue to use validated names.

The first full lint attempt was stopped after the scratch-copy controls found that an external adapter imported the original RuleContext, which has a different nominal identity from the copied one. Adapters now live in each owned rule directory and import the copied context relatively. The runtime checker remains the area's checker; this fixes module identity without editing the harness.

The shared `jsxSources` capture-count guard, `stage1/cohere/lint/jsx_integration_test.go:63`, has a fixed five-rule map. The two newly registered React ports add 46 jsx-fragments and 45 jsx-no-undef cases; the guard reports the exact seven-rule map and refuses it. Both TestJsxLintTrees and enabled TestJsxLintReleaseAndThroughput are blocked on adding those two inventory rows while retaining all existing counts. Production symbols are `JsxFragments`, `cohere/internal/lint/rules/react/jsx_fragments.go:154`, and `JsxNoUndef`, `cohere/internal/lint/rules/react/jsx_no_undef.go:83`. Shared harness edits are outside this unit's authorization, so the guard is unchanged. Per-rule parity and mutants continue; no full-package green is claimed while these guards fail.

## Recovery cases that stop shared upstream certification

`TestRulesAgree` stops at the captured `function f() { throw; }` case. The source is `TestNoThrowLiteralHandlesShapesTheCorpusOmits`, `cohere/internal/lint/rules/core/no_throw_literal_test.go:162`; production symbol `NoThrowLiteral` is at `cohere/internal/lint/rules/core/no_throw_literal.go:81`. The unchanged Go corpus guard rejects its parser diagnostics before native, Node or emitted comparison. `stage1/cohere/lint/lint_test.go:382` sends typed rows directly to that guard; capture classification currently handles only the previously enumerated recovery cases. This is a shared typed-recovery classification/refusal gap, not a dropped fixture or successful findings parity.

A separate complete matrix attempts every owned captured case independently with the same gate-built production Go oracle, ASan/UBSan native executable and emitted module, and the source Node loader. It never changes the full gate or marks malformed cases valid. It records each oracle rejection, then continues the remaining cases. In addition to the throw case, Go rejects the legacy-octal string fixtures in `TestNoUselessBackreferenceStaysSilent`, `cohere/internal/lint/rules/core/no_useless_backreference_test.go:150` (`'\1(a)'`) and line 172 (`RegExp('\1(a)')`). Production symbol `NoUselessBackreference` is at `cohere/internal/lint/rules/core/no_useless_backreference.go:77`. These require the same shared recovery handling. Their source, options, command outputs and exact diagnostics are retained; no oracle guard was relaxed.

The isolated matrix uses one strict project per captured source, preserving the source's original file-name suffix and directories and decoding its captured options through the production adapter. Native produces the area's transcript, then Node and emitted JavaScript replay it. Outputs are compared as raw bytes, with nonempty stderr or unsuccessful execution also failing. Per-side elapsed times and SHA-256 hashes are recorded. The matrix runs alongside the full lint gate, so its timings are observations under that concurrent load rather than a controlled speed benchmark.

Two valid prefer-arrow-callback cases produce identical findings and complete edit lists on all four sides, then diverge in the shared fix applicator. Reproducers: `foo(bar || function() { this; }.bind(this));` and `foo(function() {}.bind(this).bind(obj))`. They are at `cohere/internal/lint/rules/core/prefer_arrow_callback_test.go:73` and `:95`; production edit builder `preferArrowCallbackFix` is at `cohere/internal/lint/rules/core/prefer_arrow_callback.go:468`. The shared Go applicator rejects a closing insertion overlapping another edit, then refuses the invalid rewritten candidate and preserves the original source. Shared `stage1/cohere/lint/lint.ts:198` instead constructs Parser directly and panics at `stage1/typescript/parser/parser.ts:48`, before serializing the rejected edits. This is a shared fixer/parser refusal gap. The owned rule does not omit, reorder or reshape Go edits to bypass it. Complete reproducer streams are included in the isolated case evidence.

## Complete owned upstream matrix

936 unique source/rule/options combinations were attempted: 931 four-runtime byte matches, three explicit Go recovery-guard refusals and two shared fixer/parser failures. Thus ten of thirteen unified ports certify every captured upstream case; no-throw-literal certifies 47/48, no-useless-backreference 406/408 and prefer-arrow-callback 81/83, with the exact blockers above. All thirteen firing witnesses and registered mutants are separately checked by the unchanged package. The landed max-nested-callbacks branch also has a fresh 36/36 upstream matrix, with 11,438 identical bytes.

| Rule | Cases matching / attempted | Go seconds | Sanitized native seconds | Node seconds | Emitted JavaScript seconds |
| --- | ---: | ---: | ---: | ---: | ---: |
| nexus/correctness-no-collection-misuse | 6/6 | 0.438 | 0.686 | 1.900 | 0.687 |
| nexus/correctness-no-discarded-pure-result | 22/22 | 1.540 | 2.528 | 6.966 | 2.369 |
| no-eval | 91/91 | 6.754 | 10.460 | 29.518 | 9.015 |
| no-extend-native | 60/60 | 4.257 | 6.896 | 19.015 | 5.363 |
| no-func-assign | 48/48 | 3.768 | 5.617 | 16.172 | 5.056 |
| no-new-func | 41/41 | 2.819 | 4.699 | 13.264 | 3.846 |
| no-new-native-nonconstructor | 15/15 | 1.015 | 1.670 | 4.611 | 1.163 |
| no-new-wrappers | 23/23 | 1.765 | 2.653 | 7.361 | 2.336 |
| no-throw-literal | 47/48 | 3.295 | 5.261 | 15.065 | 4.482 |
| no-useless-backreference | 406/408 | 29.048 | 45.135 | 129.332 | 35.585 |
| prefer-arrow-callback | 81/83 | 6.165 | 9.200 | 26.525 | 8.320 |
| react/jsx-fragments | 46/46 | 3.676 | 5.372 | 16.386 | 4.610 |
| react/jsx-no-undef | 45/45 | 5.257 | 6.625 | 21.748 | 5.448 |

Times sum successful isolated process executions, including program load, lint, native recording or Node replay as appropriate. They exclude the blocked/failing cases, and were measured concurrently with the full gate. Matrix outer wall: 574.624 seconds. Max-nested-callbacks: Go 4.861 seconds, sanitized native 5.472 seconds, Node 19.635 seconds, emitted JavaScript 4.874 seconds, outer wall 34.860 seconds. No speedup claim follows from these measurements.

## Final all-input gate result

Command: `go test -json ./stage1/cohere/lint -count=1 -timeout=90m`, with the setup environment sourced and every benchmark, TypeScript and profile input set by `validation-wave23-unified/landing-final-runner.py`.

Package FAIL: 31 pass, 3 fail, 1 skip at top level; including subtests, 123 pass, 3 fail, 1 skip. Failed tests are TestRulesAgree (typed malformed throw fixture), TestJsxLintTrees and TestJsxLintReleaseAndThroughput (the two new JSX inventory rows). Every failure is retained, with its exact reproducer above. The full gate was not relaxed or filtered. Only skip: TestCheckerBridgeRefusalPending, `awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer`. No input-related skip occurred.

All 88 registered mutants caught, including all thirteen new verdict omissions and the landed max-nested-callbacks mutant. All firing witnesses passed the fresh unchanged TestOwnedWitnesses run (11.89 seconds), with source Node, emitted JavaScript and sanitized native matching the independent Go oracle. Compiler/stage1 corpus comparison passed on 969 files (477.68 seconds). Throughput, shards, profiles, cache-byte, replay integrity, suggestions, factory hooks and other controls passed. The full syntax corpus runs retain the area's explicit unavailable-checker coverage model; the seventeen old typed suites separately retain live native/Go checks on their two frozen corpora. No emitted-JavaScript certification is claimed for the four unactivated original type-aware migrations or the four parked analysis rules.

Package elapsed 2600.703 seconds; outer wall 2610.145 seconds (43m30s). nproc 5; quota four cores. One-minute load at start 0.689, end 1.710, sampled peak 5.302; final 1/5/15-minute averages 1.710/1.829/2.431. The complete event log and exact timing/count JSON are in validation-wave23-unified.

The retained seventeen rule ports passed all seven rule suites, on both corpora and under sanitizers, with 24 compiled rule/fact/option mutants caught and released-handle checks retained. Its historical listener probe initially failed stale named/numeric metadata after the checker pin change; the corrected focused re-run passed all eighteen compiled metadata mutants (12.52 seconds). The initial package-level failure is kept as evidence, not reported as a clean package pass. Bridge checker tests pass (0.204 seconds); registry, owned gofmt, lint/typeaware vet and whitespace checks pass. No shared test, generator or protected compiler file was edited.

The branch remains blocked on the named shared checker/helper, recovery, fixer, JSX inventory and HIR/capture dependencies. Integration can review the completed ports and the unapplied two-row JSX proposal. No new rules were claimed.
