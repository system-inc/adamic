Built: clean landing-first rebase of all 34 own commits onto the updated lint area; no new rule implementation or claims.
Commits: validated code 9ab7f098caaadb73c5441ed84c3dfae66ceec0e7 on area b84a9d9314b65d3d0261ee017e233287b4f071da, containing current main c7991b900362796aefd111474e65eb5398e91953; evidence follows on codex/typeaware-wave-30.
Commands and outputs: twelve own gates PASS 464.787s; mandatory 440-file compiler comparison plus comparator mutant PASS 119.151s; uncached Node PASS 3.639s; lowering PASS 1.169s; records PASS 34.862s; vet PASS; setup 105s, nproc 5.
Mutants: all 25 own output mutants plus three descriptor-name and six JSX-output checks caught; shared emitted-JavaScript extra output and inherited record-runtime memory/read mutants also caught.
Not covered: complete JSX source-rule parity and timings, checker-context integration, parked native HIR/SSA/capture analyses, other packages' required external checks and full repository gate.

Main advanced from 39638d9e to c7991b900. The lint area now includes it at
b84a9d931, retaining its harness, JSX parser and earlier runtime-profile work.
All 34 wave-30 commits rebased cleanly onto that requested area. Integration
introduces proven-predicate/relation lowering and native record runtime. Those
changes are accepted through rebase unchanged. No protected compiler/runtime,
shared harness/generator/oracle list or checker-context file was edited by this
unit. No main or area branch is a push target. History replacement uses an exact
lease against old own branch tip 0acfcadb1, following Ahra's rebase-and-push
instruction. The own branch's prior reports retain their historical SHAs; the
34-commit old/new map is validation-wave-30-predicates-landing/rebase.json.

The eight complete own behavior ports match production Go byte for byte on
findings, fixes and suggestions over controls and frozen 77 compiler/287
repository roots, including sanitizer variants and released-handle refusal
checks. Corpora have zero reports for those eight; positive controls exercise
the findings. No stage3 expansion of those manifests is claimed. Parked React
and partial JSX components again agree with Go, sanitized native and emitted
JavaScript. The six owned JSX probes still require actual JSX extraction.
No newly integrated checker interface is present in registered RuleContext;
its missing checker program/node correlation remains the exact shared gap for
the three partial JSX source rules. Their descriptors still declare named kinds
and partial status, not complete registry integration. The HIR/SSA/capture-based
React claims remain PARKED. No additional rule is claimed.

The required shared TestCompilerAndStage1Agree explicitly receives pinned
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript. All 440 compiler/stage1
files again produce 21021585 identical bytes across Go, Node, emitted JavaScript
and sanitized native; it executes without a skip and passes in 88.39s. The
shared comparator's extra-output JavaScript mutant is rejected in 30.74s.
The combined command passes in 119.151s. Its intentional failing child test is
verified by the passing outer mutant test, not an unresolved gate failure.

All five inherited proven_guards, proven_class_guards, proven_assertions,
proven_satisfies and proven_upcasts fixtures run through the uncached Node oracle.
They pass in 3.639s with 15 native and 10 Node cache misses. New lowering proof
and refusal tests pass in 1.169s. Native record answers match Node and its memory
and read mutants pass in 34.862s. Record mutants include overwrite-key leakage
caught by LeakSanitizer and prematurely freed stored keys caught by
AddressSanitizer. These are inherited runtime proofs, not new lint-rule mutants.

Own mutant observations from the final gate:

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
wave_30_test.go:112: iso-range exits 0; independent Go bytes catch byte 79
wave_30_test.go:112: callback-range exits 0; independent Go bytes catch byte 4157
wave_30_third_test.go:120: timer-range exits 0; independent Go bytes catch byte 89
wave_30_third_test.go:203: 216 state transitions agree; catch-state mutant exits 0 and Go catches byte 5
```

Whole-process native versus Go observations, including checker load and concurrent
validation activity, not benchmark medians or speedup claims:

```
wave_30_next_test.go:175: repository whole process native=358.042031ms Go=168.484072ms; tsgo: load_ns=79955899 query_ns=70962996 queries=4296 first_query_ns=183675 run_ns=271643175
wave_30_next_test.go:175: compiler whole process native=2.877153297s Go=908.23516ms; tsgo: load_ns=252933065 query_ns=862421524 queries=15564 first_query_ns=5975737 run_ns=2597846806
wave_30_test.go:130: repository whole process native=296.183416ms Go=148.842834ms; tsgo: load_ns=87957262 query_ns=8428107 queries=117 first_query_ns=246429 run_ns=201952756
wave_30_test.go:130: compiler whole process native=1.833684878s Go=306.575435ms; tsgo: load_ns=246789909 query_ns=105419181 queries=71 first_query_ns=78971576 run_ns=1567621353
wave_30_third_test.go:138: repository whole process native=260.233724ms Go=143.807458ms; tsgo: load_ns=78153112 query_ns=0 queries=0 first_query_ns=0 run_ns=177038931
wave_30_third_test.go:138: compiler whole process native=1.620734484s Go=302.449243ms; tsgo: load_ns=237946031 query_ns=0 queries=0 first_query_ns=0 run_ns=1367530727
```

Commands after source /workspace/adamic-tools/env.sh, each writes to its log:

- git rebase origin/area/stage1-lint; bash cloud/setup.sh. Go/clang/Node/submodules
  0s, cache warm and total 105s; nproc 5 and CPU quota 4 cores.
- ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript, repository and compiler
  manifests set for ADAMIC_WAVE_30, NEXT, THIRD and PROCESS, then go test
  ./stage1/cohere/typeaware -run '^TestWave30' -count=1 -timeout 30m -v.
- ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript go test
  ./stage1/cohere/lint -run '^(TestCompilerAndStage1Agree|TestEmittedJavaScriptMismatch)$'
  -count=1 -timeout 30m -v.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  '^TestNativeAgreesWithNode$/internal/oracle/testdata/proven_(guards|class_guards|assertions|satisfies|upcasts).a$'
  -count=1 -timeout 10m -v.
- go test ./internal/lower -run
  '^Test(UnprovenPredicateReturnsAreRefused|PredicateBodiesAreProven|ProvenRelationsRefuse|ProvenRelationsErase)$'
  -count=1 -v.
- go test ./internal/native -run
  '^Test(RecordsAgainstNode|RecordMutants|RecordReadMutants)$' -count=1 -v.
- go vet ./stage1/cohere/typeaware ./bridge/tsgo/...; git diff --check.

Exact logs are gzip-preserved with verified decompression and uncompressed
SHA-256 hashes beside the rebase map. Verified binaries/archives in eleven
named completed own scratch directories were removed to free 2192466627 bytes;
sources and observations remain, with both cleanup manifests preserved.
No full repository gate or other packages' required postcss/GraphQL/external
compiler checks are represented as run. No skip was introduced, relaxed or
removed to obtain these results. No new authored Adamic .ts files are added.
