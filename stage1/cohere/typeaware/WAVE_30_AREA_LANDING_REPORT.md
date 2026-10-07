Built: rebased wave-30 onto the integrated lint area including its runtime-profile changes; changed owned refusal probes into positive JSX-extraction checks.
Commits: final validated code 3f1c4d1824b6725509ea378e697c13549af8940a on area d65a8f931c98655936ae04c6899f38f14862b73e, containing current main 39638d9e278d38bb5aeae887f46d55a70e47aaad.
Commands and outputs: final twelve own gates PASS 459.883s; focused shared parity PASS 102.543s; runtime tests PASS 8.216s; uncached Node search oracle PASS 0.451s; vet PASS; setup 42s, nproc 5.
Mutants: all 25 successful-exit output mutants, three descriptor-name check mutants and six missing-JSX-output check mutants caught; individual final observations follow below.
Not covered: complete source-rule JSX findings/fixes/suggestions/corpus parity, their timings, native bindings/stability and parked HIR/SSA/capture analyses; full shared package was not repeated on the second runtime snapshot.

The user explicitly instructed rebase onto origin/area/stage1-lint. Initial
area tip 7481e0324 integrates the harness, JSX parser and main 39638d9e. The
30-commit rebase was clean. The initial twelve own gates passed in 467.186s;
the complete shared lint package passed in 1044.879s, including registry-owned
witnesses and mutants, suggestion serialization, .a modules and JSX integration.
The registry package also passed in 0.026s. Optional throughput/profile probes
with unset environment inputs are skipped; no complete performance benchmark
is inferred from these gates.

While those tests ran, area advanced to d65a8f931 with release-path and string
runtime optimizations. The second clean rebase retains all those changes and
31 own commits, including the owned probe update. All twelve own gates were
then repeated on that final runtime snapshot. Its focused shared tests are
TestRulesAgree, TestWitnessScriptKind and TestCompleteSuggestionSerialization.
RuntimeReleasePaths and RuntimeStringEquality pass, and the uncached
TestRuntimeLastIndexOfMatchesNode oracle passes with three native and two Node
cache misses. The entire 17-minute shared package is recorded green on the
first area snapshot; it is not represented as rerun on the final runtime.

Only codex/typeaware-wave-30 is a push target. No main or area branch is pushed.
Current main is an ancestor of the final validated tree. Exact initial and
final own-commit rebase mappings are in validation-wave-30-area. The rebase
accepts integration-owned files unchanged, including runtime and parser work;
no shared generator, harness, parser, runtime or oracle registry was edited by
this unit. History replacement uses a lease against old own tip c28cb343e,
as required by the explicit rebase-and-push instruction.

The old private probe deliberately failed after integration because native
now extracts JsxSelfClosingElement. Its before log is preserved. Both owned
prerequisite probes now require the expected JSX kind and refuse a substituted
output with that node removed. They also reject TypeAssertionExpression on
these six controls. These six output-substitution mutants prove the extraction
checks; they are not native source-rule decision mutants. Production Go still
reports on all six controls. JSX syntax is now available, so prior parser-gap
blocker descriptions are superseded rather than hidden by skips.

The three new JSX claims remain partial components, not certified full rules.
All three production rules ask their checker for symbol/declaration facts.
For jsx-no-undef this includes module/global scope, allowGlobals, .cjs treatment
and file-local declarations. For jsx-fragments it includes imported/bare alias
binding. Constructed context also follows symbols and analyzes provider origin,
construction identity and memo-input stability. Registered RuleContext exposes
source, parser, scanner, parents and finding APIs but no checker program handle
or checker-node mapping. Parser.path exposes the source path: source identity
itself is not missing. Supplying a checker with the oracle's project options,
program lifetime and parsed-node correlation requires type-aware integration.
No shared-context workaround was introduced. Owned rule.json blockers now name
this gap. Their partial metadata is outside the discovered full-rule registry;
no complete factory/visitor/oracle integration or full source findings is claimed.
The earlier React HIR/SSA/capture-dependent claims remain PARKED.

The eight existing complete behavior ports still match production Go byte for
byte on findings, fixes and suggestions over positive controls and the frozen
77 compiler/287 repository roots. Those corpora contain zero reports for all
eight; nonzero controls exercise each decision. Native sanitizer variants agree
and released checker handles refuse explicitly. Corpus scope is not expanded to
all newly integrated source files. JSX and parked React components match Go,
sanitized native and emitted JavaScript. New bridge questions are not added.
No new authored Adamic .ts files or additional rule claims are introduced.

Final individual mutant observations:

```
wave_30_jsx_components_test.go:102: jsx-fragments: wrong-name descriptor mutant caught by pinned Go kind-name comparison
wave_30_jsx_components_test.go:159: fragments: 1215 Go bytes match native, sanitizer and emitted JavaScript; successful-exit mutant caught only by Go at byte 98
wave_30_jsx_components_test.go:102: jsx-no-undef: wrong-name descriptor mutant caught by pinned Go kind-name comparison
wave_30_jsx_components_test.go:159: undef: 12037 Go bytes match native, sanitizer and emitted JavaScript; successful-exit mutant caught only by Go at byte 22
wave_30_jsx_components_test.go:102: jsx-no-constructed-context-values: wrong-name descriptor mutant caught by pinned Go kind-name comparison
wave_30_jsx_components_test.go:159: context: 148002 Go bytes match native, sanitizer and emitted JavaScript; successful-exit mutant caught only by Go at byte 44006
wave_30_jsx_components_test.go:186: Unicode astral escape mutant exits successfully and is caught only by Go at byte 13368
wave_30_jsx_prerequisites_test.go:61: react/jsx-fragments: Go reports; native parser extracts JsxSelfClosingElement; missing-node mutant caught; full rule/analysis parity is not established
wave_30_jsx_prerequisites_test.go:61: react/jsx-no-constructed-context-values: Go reports; native parser extracts JsxSelfClosingElement; missing-node mutant caught; full rule/analysis parity is not established
wave_30_jsx_prerequisites_test.go:61: react/jsx-no-undef: Go reports; native parser extracts JsxSelfClosingElement; missing-node mutant caught; full rule/analysis parity is not established
wave_30_listeners_test.go:63: listener-mutant-consistency_no_iso_string_date_cut caught only by Go comparison at byte 37
wave_30_listeners_test.go:63: listener-mutant-correctness_no_callback_in_parse_try caught only by Go comparison at byte 86
wave_30_listeners_test.go:63: listener-mutant-correctness_no_collection_misuse caught only by Go comparison at byte 123
wave_30_listeners_test.go:63: listener-mutant-correctness_no_discarded_outcome caught only by Go comparison at byte 168
wave_30_listeners_test.go:63: listener-mutant-correctness_no_discarded_pure_result caught only by Go comparison at byte 209
wave_30_listeners_test.go:63: listener-mutant-correctness_no_uncleared_race_timeout caught only by Go comparison at byte 251
wave_30_listeners_test.go:63: listener-mutant-correctness_no_process_exit_after_output caught only by Go comparison at byte 296
wave_30_listeners_test.go:63: listener-mutant-correctness_require_blocking_standard_streams caught only by Go comparison at byte 346
wave_30_next_test.go:157: collection-boundary exits 0; independent Go bytes catch byte 430
wave_30_next_test.go:157: outcome-range exits 0; independent Go bytes catch byte 27672
wave_30_next_test.go:157: pure-range exits 0; independent Go bytes catch byte 1119
wave_30_process_test.go:76: graph mutant exits 0; Go catches byte 988
wave_30_process_test.go:147: exit mutant exits 0; Go catches byte 78
wave_30_process_test.go:223: blocking mutant exits 0; Go catches byte 107
wave_30_react_components_test.go:111: refs: 279296 exact bytes agree, sanitizer and emitted JavaScript agree; mutant exits 0 with empty stderr, Go catches byte 86865
wave_30_react_components_test.go:111: purity: 6523 exact bytes agree, sanitizer and emitted JavaScript agree; mutant exits 0 with empty stderr, Go catches byte 1368
wave_30_react_components_test.go:111: memo: 900 exact bytes agree, sanitizer and emitted JavaScript agree; mutant exits 0 with empty stderr, Go catches byte 435
wave_30_react_prerequisites_test.go:64: react-hooks/purity: Go reports; native parser extracts JsxElement; missing-node mutant caught; full rule/analysis parity is not established
wave_30_react_prerequisites_test.go:64: react-hooks/refs: Go reports; native parser extracts JsxElement; missing-node mutant caught; full rule/analysis parity is not established
wave_30_react_prerequisites_test.go:64: react-hooks/preserve-manual-memoization: Go reports; native parser extracts JsxSelfClosingElement; missing-node mutant caught; full rule/analysis parity is not established
wave_30_test.go:112: iso-range exits 0; independent Go bytes catch byte 80
wave_30_test.go:112: callback-range exits 0; independent Go bytes catch byte 4166
wave_30_third_test.go:120: timer-range exits 0; independent Go bytes catch byte 89
wave_30_third_test.go:203: 216 state transitions agree; catch-state mutant exits 0 and Go catches byte 5
```

Final whole-process native versus Go measurements, including checker load and
observed during validation rather than benchmark medians:

```
wave_30_next_test.go:175: repository whole process native=389.959687ms Go=175.287025ms; tsgo: load_ns=84830906 query_ns=85264292 queries=4296 first_query_ns=213850 run_ns=298797599
wave_30_next_test.go:175: compiler whole process native=2.684874526s Go=794.283573ms; tsgo: load_ns=235972968 query_ns=803556090 queries=15564 first_query_ns=7200944 run_ns=2427729481
wave_30_test.go:130: repository whole process native=285.505907ms Go=138.811177ms; tsgo: load_ns=79796096 query_ns=8627851 queries=117 first_query_ns=284226 run_ns=200443103
wave_30_test.go:130: compiler whole process native=1.76263881s Go=306.728712ms; tsgo: load_ns=250662235 query_ns=117342681 queries=71 first_query_ns=94728732 run_ns=1498712014
wave_30_third_test.go:138: repository whole process native=259.265752ms Go=126.652253ms; tsgo: load_ns=77706746 query_ns=0 queries=0 first_query_ns=0 run_ns=174326547
wave_30_third_test.go:138: compiler whole process native=1.640825465s Go=358.017191ms; tsgo: load_ns=248680308 query_ns=0 queries=0 first_query_ns=0 run_ns=1377583361
```

Whole-rule timings for the partial JSX ports remain unavailable. Validation
costs and earlier syntax-lint runtime profiles do not establish their speed.

Commands after source /workspace/adamic-tools/env.sh, each writes to its log:

- git rebase origin/area/stage1-lint; bash cloud/setup.sh. Go, clang, Node and
  submodules 0s, warm 42s, total 42s; nproc 5, CPU quota 4 cores.
- ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript with repository/compiler
  manifests for ADAMIC_WAVE_30, NEXT, THIRD and PROCESS, then go test
  ./stage1/cohere/typeaware -run '^TestWave30' -count=1 -timeout 30m -v, on both
  area snapshots as described above.
- go test ./stage1/cohere/lint/registry -count=1 -v.
- go test ./stage1/cohere/lint -count=1 -timeout 30m -v, on the first snapshot.
- go test ./internal/native -run '^TestRuntime(ReleasePaths|StringEquality)$'
  -count=1 -v; ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle
  -run '^TestRuntimeLastIndexOfMatchesNode$' -count=1 -v, on the final snapshot.
- go test ./stage1/cohere/lint -run
  '^(TestRulesAgree|TestWitnessScriptKind|TestCompleteSuggestionSerialization)$'
  -count=1 -timeout 15m -v, on the final snapshot.
- go vet ./stage1/cohere/typeaware ./bridge/tsgo/...; git diff --check.

All logs, before-failure evidence and provenance mappings are retained beside
this report in validation-wave-30-area. Only verified binaries/archives from
four named completed own scratch roots were removed to free 1483875632 bytes;
source and exact oracle streams remain, with a removal manifest preserved.
No full repository gate or full compiler oracle matrix is claimed.
