Built: landing refresh of the existing wave-30 work after main's inherited-static-field compiler fix; no new rules claimed.
Commits: clean 29-commit rebase produces ab243cd553c71336bb65eb0d9725667b2acb029e on current origin/main b8fb957aa839a9e8cb0b54279dd9864fa317bd30; evidence follows on the own branch.
Commands and outputs: all twelve wave-30 gates PASS 417.687s; uncached inherited-static-field Node oracle PASS 0.483s; vet PASS; setup 71s, nproc 5.
Mutants: all 25 successful-exit output mutants and three changed-kind-name metadata mutants caught again; individual evidence follows below.
Not covered: complete JSX source rules, HIR/SSA/capture analyses, new-rule compiler/repository parity or timings, shared visitor integration and full repository gate.

The unit's only pushed branch is codex/typeaware-wave-30. Main advanced from
c01907a70 to b8fb957aa with inherited static field reads. The 29-commit rebase
was clean, retaining the landed change to emit_objects.go. No compiler, shared
harness, generator or integration-owned file was edited by this unit. No push
to main or any area branch is performed. The own history update is required by
Ahra's explicit rebase-and-push instruction and uses an exact remote-tip lease.
The announced ab70f38d4 harness remains outside main; JSX and handed-node shared
integration are still prerequisites. Named rule.json kinds remain the required
interface, not numeric ordinals. Legacy numeric declaration checks describe
older retained evidence and are not claimed as registry integration.

All twelve TestWave30 gates execute on the new compiler base. The eight complete
behavior ports again agree byte for byte with production Go on findings, fixes
and suggestions over positive controls and the frozen 77 compiler/287 repository
root manifests. These corpora produce zero findings for all eight; the controls
exercise nonzero reports. Native sanitizer variants agree and released bridge
handles refuse explicitly. Corpus scope was not expanded for stage3. The three
partial JSX component suites again agree with Go, native sanitizers and emitted
JavaScript; complete JSX source-rule parity is not claimed. The prior three
React HIR/SSA/capture analyses remain PARKED with their blockers named.

The newly landed compiler fixture is checked uncached against Node and emitted
JavaScript/native sanitizer behavior. It passes, with native cache misses 3 and
Node cache misses 2. This filtered compiler oracle supplements the own gates;
it is not the full oracle matrix.

Commands, after source /workspace/adamic-tools/env.sh, each write to retained logs:

- git rebase origin/main; bash cloud/setup.sh. Setup Go/clang/Node/submodules 0s,
  build cache warm 71s, total 71s; nproc 5 and CPU quota 4 cores.
- ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript, with repository and
  compiler manifests set for ADAMIC_WAVE_30, NEXT, THIRD and PROCESS, then
  go test ./stage1/cohere/typeaware -run '^TestWave30' -count=1 -timeout 30m -v.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/inherited_static_field_read.a$'
  -count=1 -timeout 10m -v.
- go vet ./stage1/cohere/typeaware ./bridge/tsgo/...; git diff --check.

Individual mutant observations from the oracle log:

```
wave_30_jsx_components_test.go:102: jsx-fragments: wrong-name descriptor mutant caught by pinned Go kind-name comparison
wave_30_jsx_components_test.go:159: fragments: 1215 Go bytes match native, sanitizer and emitted JavaScript; successful-exit mutant caught only by Go at byte 98
wave_30_jsx_components_test.go:102: jsx-no-undef: wrong-name descriptor mutant caught by pinned Go kind-name comparison
wave_30_jsx_components_test.go:159: undef: 12037 Go bytes match native, sanitizer and emitted JavaScript; successful-exit mutant caught only by Go at byte 22
wave_30_jsx_components_test.go:102: jsx-no-constructed-context-values: wrong-name descriptor mutant caught by pinned Go kind-name comparison
wave_30_jsx_components_test.go:159: context: 148002 Go bytes match native, sanitizer and emitted JavaScript; successful-exit mutant caught only by Go at byte 44006
wave_30_jsx_components_test.go:186: Unicode astral escape mutant exits successfully and is caught only by Go at byte 13368
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
wave_30_process_test.go:223: blocking mutant exits 0; Go catches byte 106
wave_30_react_components_test.go:111: refs: 279296 exact bytes agree, sanitizer and emitted JavaScript agree; mutant exits 0 with empty stderr, Go catches byte 86865
wave_30_react_components_test.go:111: purity: 6523 exact bytes agree, sanitizer and emitted JavaScript agree; mutant exits 0 with empty stderr, Go catches byte 1368
wave_30_react_components_test.go:111: memo: 900 exact bytes agree, sanitizer and emitted JavaScript agree; mutant exits 0 with empty stderr, Go catches byte 435
wave_30_test.go:112: iso-range exits 0; independent Go bytes catch byte 80
wave_30_test.go:112: callback-range exits 0; independent Go bytes catch byte 4166
wave_30_third_test.go:120: timer-range exits 0; independent Go bytes catch byte 89
wave_30_third_test.go:203: 216 state transitions agree; catch-state mutant exits 0 and Go catches byte 5
```

Whole-process native versus Go timings, single observations including checker
loading, not benchmark medians or speedup claims:

```
wave_30_next_test.go:175: repository whole process native=347.772804ms Go=152.648813ms; tsgo: load_ns=71415946 query_ns=67441389 queries=4242 first_query_ns=183736 run_ns=270130489
wave_30_next_test.go:175: compiler whole process native=2.695043923s Go=771.45692ms; tsgo: load_ns=236214240 query_ns=802723289 queries=15564 first_query_ns=5452503 run_ns=2435397212
wave_30_test.go:130: repository whole process native=257.753592ms Go=117.441048ms; tsgo: load_ns=65728269 query_ns=7228414 queries=117 first_query_ns=420641 run_ns=186871197
wave_30_test.go:130: compiler whole process native=1.747891249s Go=289.160515ms; tsgo: load_ns=241186810 query_ns=98798099 queries=71 first_query_ns=73496695 run_ns=1491360708
wave_30_third_test.go:138: repository whole process native=247.547415ms Go=126.638579ms; tsgo: load_ns=65458105 query_ns=0 queries=0 first_query_ns=0 run_ns=177581478
wave_30_third_test.go:138: compiler whole process native=1.642441695s Go=294.086709ms; tsgo: load_ns=236125051 query_ns=0 queries=0 first_query_ns=0 run_ns=1393346486
```

Complete JSX native/Go timings remain unavailable because the positive Go TSX
controls still cannot be extracted by the native parser. Native bindings,
context stability and full visitor registration also remain unfinished.
No full repository gate, expanded compiler corpus or full oracle fixture matrix
is represented. Logs, all gate names, mutant/timing extracts and the exact
29-commit old/new mapping are retained in validation-wave-30-static-landing.
Only verified ELF/archive artifacts in six named completed own scratch roots
were removed to free 2133396864 bytes; source and exact oracle streams were
retained, and the removal manifest is archived. New Adamic source remains .a.
