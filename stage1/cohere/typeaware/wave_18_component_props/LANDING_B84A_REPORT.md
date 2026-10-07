Built: rebased 17 completed native ports and four parked claims onto current main through lint area b84a9d931; no rule implementation changes.
Commits: tested f5324d4118e7fd4b5783239cbceab25b3e82bc90; area b84a9d9314b65d3d0261ee017e233287b4f071da includes current fetched main c7991b900362796aefd111474e65eb5398e91953.
Checks: six owned suites pass together in 864.739s; rebuilt standalone groups, sanitizers, released handles, parked reporters, five shared harness checks, uncached Node proof fixtures and vet pass.
Mutants: 29 Go-byte mutations, six stale-registry checks, three removed-refusal checks, parser-count, Node one-byte and emitted-JavaScript mismatch checks caught their mutations.
Not covered: four HIR/SSA/capture claims remain parked; full gate and its 17 external correctness checks not run; no full shared native JSX rule integration; zero unclaimed ranked rules remain.

Main and area advanced with the stage3 proven relations/predicates, cast/refusal checks and record runtime. Rebased the same owned branch onto origin/area/stage1-lint b84a9d931, which includes origin/main c7991b900. Incoming changes were retained; no protected compiler, shared generator or shared harness file was edited. Only this report, evidence and the owned claim status were added after rebase. No main or area push was attempted.

Setup: bash cloud/setup.sh succeeded. Go, clang, Node and submodules ready in 0s each, cache warm in 156s, total 156s. nproc=5, quota=4 CPUs, memory=17.6 GB; Go 1.27.1, clang 20.1.8, Node 24.19.0. Every test shell sourced /workspace/adamic-tools/env.sh. For space, removed 56 compiled ELF/archive artifacts totaling 1,091,805,745 bytes from the completed constructor group and previous-run artifacts in not-yet-started reactor/base/title groups. Sources and logs were preserved. The earlier attempt to clear all groups detected that the oracle had started and left them untouched.

Commands and results:

- `go test ./stage1/cohere/typeaware -run '^TestWave18' -count=1 -timeout=30m -v`: PASS 864.739s, all six together. Constructor 83.30s, core 85.43s, preference 101.10s, React prepared cores 455.09s, original two-rule suite 89.89s, title 49.92s. Exact per-test times and full output are archived.
- `python3 stage1/cohere/typeaware/wave_18_jsx/validate.py --scratch /workspace/wave18-jsx-reproduce` and `validate_bridge.py`: rebuilt stage0/Go oracle/bridge/native after removing native-asan bootstrap sentinel; PASS complete findings, automatic edits and suggestions over 130 valid controls, four option profiles, both frozen corpora, normal/sanitized runs, three per-rule mutants, numeric-question mutation and released handles.
- Component-property `validate.py` and `validate_bridge.py --scratch /workspace/wave18-component`: freshly rebuilt in the same way; PASS 73 controls, six option profiles, both frozen corpora, normal/sanitized complete records, two per-rule mutants, numeric-question mutation and released handles.
- `validate_partial.py --scratch /workspace/wave18-area/partial`: PASS reporting, sanitizer, three Go-byte mutations, three removed-refusal mutations and compiling parser-count mutation. Four production source-analysis claims remain parked: set-state-in-effect, set-state-in-render, static-components and jsx-no-constructed-context-values. Named blockers remain native HIR/SSA/capture and callback/return escape analysis, rather than JSX syntax support.
- `go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1 -timeout=10m`: PASS 0.505s/0.158s.
- Shared harness filter `TestEmittedJavaScriptMismatch|TestDotARename|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix|TestWitnessScriptKind`: PASS 212.365s, holding Go/Node/emitted-JavaScript/native complete records. Expected inner failure from the emitted-JavaScript mutant is caught by the passing outer test.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle` filtered to one-byte mutation, native-agrees-with-Node closures/method_closures/generic_functions/regions/regions_throw/inherited_static_field_read/runtime_last_index_of/proven_assertions/proven_class_guards/proven_guards/proven_satisfies/proven_upcasts, plus TestRuntimeLast and TestProven: PASS 17.876s, native 43 misses and Node 29 misses, no hits. Prefix selection also covers regexp_cycle_closures. Newly landed proof fixtures match Node and native/emitted JavaScript under sanitizers.
- `go vet ./...`: PASS, empty output.

Pinned inputs were supplied explicitly: ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript; dedicated title slice /workspace/wave-18-jsx-slice; compiler config /workspace/wave-18-typescript/src/compiler/tsconfig.json with /tmp/wave-18-compiler.manifest (77 roots); repository config /workspace/adamic/tsconfig.json with /tmp/wave-18-repository.manifest (287 roots). Corpus config/manifest environment variables were also set for both standalone validators. Independent Go cohere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db. Corpora have zero findings for these groups; positive controls and compiling mutants hold rule behavior. Complete record bytes, not counts alone, were compared.

No selected test skipped in the final oracle, harness or Node logs. This was a filtered worker validation, not the full repository gate. The 17 newly required compiler/postcss/graphql/parser correctness checks outside these selected packages were not run and are not claimed green. No skip, input requirement or check was weakened or deleted. If integration finds a failure there, this report is not evidence that those checks passed.

Three alternating process medians in seconds, under concurrent validation load: JSX compiler native 1.879767 / Go 0.410173, repository 0.353360 / 0.191587. Property pair compiler native 4.634198 / Go 0.353257, repository 0.624999 / 0.170339. Full per-suite timing rows are archived. React prepared input aggregate native 0.018238s / full Go 0.290416s / Go preparation 0.271942s uses different pipelines and is not an end-to-end speed comparison. Timings are observations, not an isolated runtime improvement claim.

The title retains its isolated native JSX parser slice, and prepared React cores retain Go-prepared syntax/analysis rows. No claim of complete shared native driver integration or production native HIR analysis is made. Older rule APIs remain unchanged. No new rule or regex matcher was introduced.

Selection audit: 455 actual origin heads, 33 distinct claim blobs, all 197 ranked rules, zero eligible unclaimed entries. No new claim was made. Origin refs were fetched explicitly using +refs/heads/*:refs/remotes/origin/*; the default fetch config is narrower. Actual remote head list and selection JSON are archived. Only codex/typeaware-wave-18 is pushed.

Every mutation/refusal observation follows; complete subprocess stdout/stderr and journals are compressed in validation-b84a. Byte-only mutants compiled, exited 0 and emitted empty stderr; compile errors were not counted as mutation evidence.

```text
wave18-now-oracle.log:     wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1100
wave18-now-oracle.log:     wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11255
wave18-now-oracle.log:     wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13042
wave18-now-oracle.log:     wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
wave18-now-oracle.log:     wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-now-oracle.log:     wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 560
wave18-now-oracle.log:     wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 41361
wave18-now-oracle.log:     wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-now-oracle.log:     wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1593
wave18-now-oracle.log:     wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 32342
wave18-now-oracle.log:     wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 21969
wave18-now-oracle.log:     wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 32254
wave18-now-oracle.log:     wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
wave18-now-oracle.log:     wave_18_react_cores_test.go:176: render-unconditional: exit 0, empty stderr; independent complete-byte comparison catches byte 400
wave18-now-oracle.log:     wave_18_react_cores_test.go:176: effect-setter: exit 0, empty stderr; independent complete-byte comparison catches byte 1507
wave18-now-oracle.log:     wave_18_react_cores_test.go:176: static-creator: exit 0, empty stderr; independent complete-byte comparison catches byte 50
wave18-now-oracle.log:     wave_18_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3067
wave18-now-oracle.log:     wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-now-oracle.log:     wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4344
wave18-now-oracle.log:     wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1662
wave18-now-oracle.log:     wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7707
wave18-now-oracle.log:     wave_18_test.go:196: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-now-oracle.log:     wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 50
wave18-now-jsx.log: fragments: compiled, exit 0, empty stderr; Go byte comparison catches byte 446
wave18-now-jsx.log: undef: compiled, exit 0, empty stderr; Go byte comparison catches byte 8288
wave18-now-jsx.log: adjacent: compiled, exit 0, empty stderr; Go byte comparison catches byte 27811
wave18-now-jsx.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-now-jsx-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 53
wave18-now-jsx-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-now-component.log: placement: compiled, exit 0, empty stderr; Go byte comparison catches byte 5981
wave18-now-component.log: style: compiled, exit 0, empty stderr; Go byte comparison catches byte 321
wave18-now-component.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-now-component-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 49
wave18-now-component-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-now-partial.log: set_state_in_effect: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-now-partial.log: set_state_in_render: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-now-partial.log: static_components: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-now-partial.log: base JSX parser: two nodes; compiling count mutant caught by expected output
```
