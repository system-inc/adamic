Built: rebased 17 completed ports and four parked claims onto current main through the latest lint area; no rule implementation changes.
Commits: tested d88d5dd9eb7c2eca17f29471794864198e226277; area bb2ece564842c4b2f909b9f75c27e74c2efa4f29 includes fetched main 4e0bfda50a19c705a1aac0d9932e08483806d61c.
Checks: five owned suites pass in 441.360s, React in 540.639s; standalone groups, sanitizers, released handles, parked reporters, bridge, registry, seven shared harness checks, uncached Node and vet pass.
Mutants: 29 Go-byte mutations, six stale-registry checks, three removed-refusal checks, parser-count, Node one-byte, emitted-JavaScript mismatch and ignored decoded-option mutation were caught.
Not covered: four HIR/SSA/capture claims parked; full gate and its 17 external correctness checks unrun; full shared native JSX rule integration not claimed; zero unclaimed ranked rules.

Main landed the previous lint/runtime area. The latest area adds configured witness-sidecar handling and its regression data/tests. Rebased cleanly onto that area. No protected compiler, shared generator or shared harness file was edited; incoming shared changes were retained. Only this owned report, evidence and claim status were added. No push to main or any area branch was attempted.

Setup: bash cloud/setup.sh succeeded. Go/clang/Node/submodules ready in 0s each, cache warm in 141s, total 141s. nproc=5, CPU quota=4, memory=17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. All test shells sourced /workspace/adamic-tools/env.sh. Disk remained ample, about 18 GB free; no cache clearing or build recovery was needed this turn.

The five non-React TestWave18 suites were run in one Go process with an exact name filter; TestWave18ReactCores ran in a second process with disjoint scratch directories. Both used -count=1 -timeout=30m -v and returned zero. Five-suite result 441.360s: constructor 99.13s, core 89.79s, preferences 103.08s, original suite 96.09s, title 53.26s. React result 540.639s. Complete findings, automatic edits and suggestions match independent pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, normally and under sanitizers. Positive controls hold behavior where the corpora produce zero findings. Every planned byte-only mutation compiled, exited zero and emitted empty stderr before being caught by Go bytes.

Inputs were supplied explicitly: ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript; title slice /workspace/wave-18-jsx-slice; compiler config /workspace/wave-18-typescript/src/compiler/tsconfig.json and frozen /tmp/wave-18-compiler.manifest (77 roots); repository config /workspace/adamic/tsconfig.json and frozen /tmp/wave-18-repository.manifest (287 roots). Corpus environment variables were supplied to both standalone validators too. Frozen manifests and full subprocess output are archived.

Standalone JSX validate.py/validate_bridge.py passed 130 valid controls, four option profiles, three rule-byte mutations, numeric-question mutation, released handles and both normal/sanitized corpora. Component-property validators passed 73 valid controls, six option profiles, two rule-byte mutations, numeric-question mutation, released handles and both corpora. Those validators reused previously built artifacts because their source, compiler and bridge are unchanged by this test/documentation-only rebase; owned Go suites rebuilt their own binaries. Numeric question mutants change Go records at bytes 53 and 49. Stale registry mutants violate the required released-handle exit-70 contract.

Parked reporter validation passed four findings (2355 complete bytes), sanitizers, three byte mutations, three removed-refusal mutations and the compiling parser-count mutation. set-state-in-effect, set-state-in-render, static-components and jsx-no-constructed-context-values remain parked on native HIR/SSA/capture and callback/return escape analysis. JSX syntax is supported. React core tests use Go-prepared rows; title retains its private native JSX parser slice. Full production analysis and complete shared native driver integration are not claimed.

`go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1 -timeout=10m` passed in 0.753s/0.461s. Shared lint tests TestEmittedJavaScriptMismatch, TestDotARename, TestCompleteSuggestionSerialization, TestSuggestionAlongsideAutomaticFix, TestWitnessScriptKind, TestOwnedWitnesses and TestDecodedOptionsAndMutant passed together in 247.568s. This includes the newly changed configured owned-witness handling (22.34s). Ignoring decoded options is caught on emitted JavaScript and native (38.82s); the adapter guard was not relaxed. Shared complete-record Go/Node/emitted-JavaScript/native comparisons all pass. The expected inner failure from the emitted-JavaScript mutation is caught by its passing outer test.

Filtered ADAMIC_GATE_UNCACHED=1 internal/oracle checks passed in 2.574s: TestTheOracleCatchesOneByte and native-agrees-with-Node closures/method_closures/generic_functions/regions/regions_throw plus typeof_dispatch, typeof_null, typeof_null_compare, typeof_null_roll, typeof_null_slots, typeof_null_switch and typeof_string_literal. Prefix matching also selected regexp_cycle_closures. Native 40 misses, Node 27 misses, zero hits. `go vet ./...` passed with empty output. No selected test skipped. The full repository gate and its 17 external compiler/postcss/graphql/parser correctness checks were not selected and are not claimed green. No skip or input requirement was weakened or removed.

Three alternating process medians in seconds under concurrent compilation/validation load: JSX compiler native 5.899771 / Go 1.231942, repository 0.759734 / 0.567226; property pair compiler native 4.630398 / Go 0.416696, repository 0.717987 / 0.202077. All individual suite timing rows are archived. React prepared-input aggregate native 0.019604s versus full Go 0.326079s and preparation 0.330307s uses different pipelines, so cannot establish end-to-end speed. These are loaded-run observations, not isolated speed measurements or an improvement claim.

Audit: 542 actual origin heads, 42 distinct claim blobs, all 197 ranked rules, zero eligible unclaimed entries. No new claim or rule was added. Main and area are included by ancestry, and only codex/typeaware-wave-18 is pushed. Earlier reports retain the 17 ports' source and integration limits.

All mutation/refusal observations follow. Complete logs, canonical findings and subprocess journals are compressed under validation-ready. Expected inner test failures are preserved, not counted as unexpected suite failures.

```text
wave18-ready-oracle.log:     wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1100
wave18-ready-oracle.log:     wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11255
wave18-ready-oracle.log:     wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13042
wave18-ready-oracle.log:     wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
wave18-ready-oracle.log:     wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-ready-oracle.log:     wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 560
wave18-ready-oracle.log:     wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 41361
wave18-ready-oracle.log:     wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-ready-oracle.log:     wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1593
wave18-ready-oracle.log:     wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 32342
wave18-ready-oracle.log:     wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 21969
wave18-ready-oracle.log:     wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 32254
wave18-ready-oracle.log:     wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
wave18-ready-oracle.log:     wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-ready-oracle.log:     wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4344
wave18-ready-oracle.log:     wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1662
wave18-ready-oracle.log:     wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7707
wave18-ready-oracle.log:     wave_18_test.go:196: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-ready-oracle.log:     wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 50
wave18-ready-react.log:     wave_18_react_cores_test.go:176: render-unconditional: exit 0, empty stderr; independent complete-byte comparison catches byte 400
wave18-ready-react.log:     wave_18_react_cores_test.go:176: effect-setter: exit 0, empty stderr; independent complete-byte comparison catches byte 1507
wave18-ready-react.log:     wave_18_react_cores_test.go:176: static-creator: exit 0, empty stderr; independent complete-byte comparison catches byte 50
wave18-ready-react.log:     wave_18_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3067
wave18-ready-jsx.log: fragments: compiled, exit 0, empty stderr; Go byte comparison catches byte 446
wave18-ready-jsx.log: undef: compiled, exit 0, empty stderr; Go byte comparison catches byte 8288
wave18-ready-jsx.log: adjacent: compiled, exit 0, empty stderr; Go byte comparison catches byte 27811
wave18-ready-jsx.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-ready-jsx-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 53
wave18-ready-jsx-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-ready-component.log: placement: compiled, exit 0, empty stderr; Go byte comparison catches byte 5981
wave18-ready-component.log: style: compiled, exit 0, empty stderr; Go byte comparison catches byte 321
wave18-ready-component.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-ready-component-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 49
wave18-ready-component-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-ready-partial.log: set_state_in_effect: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-ready-partial.log: set_state_in_render: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-ready-partial.log: static_components: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-ready-partial.log: base JSX parser: two nodes; compiling count mutant caught by expected output
wave18-ready-harness.log:     harness_test.go:54: ordinary comparison rejected clean-running emitted JavaScript mutant:
wave18-ready-harness.log:     harness_test.go:140: second suggestion edit mutant caught on Node: case 0 line 8: port "suggestion-edit\t9 11\t", Go "suggestion-edit\t9 10\t"
wave18-ready-harness.log:     harness_test.go:140: second suggestion edit mutant caught on emitted JavaScript: case 0 line 8: port "suggestion-edit\t9 11\t", Go "suggestion-edit\t9 10\t"
wave18-ready-harness.log:     harness_test.go:140: second suggestion edit mutant caught on native: case 0 line 8: port "suggestion-edit\t9 11\t", Go "suggestion-edit\t9 10\t"
wave18-ready-harness.log:     registration_test.go:288: ignored decoded-option mutant caught on Node: case 0 line 2: port "/tmp/adamic-gate/TestDecodedOptionsAndMutant526419110/001/catch.ts:1:26", Go "fixed\ttry { work(); } catch(e) {}\\u000a"
wave18-ready-harness.log:     registration_test.go:288: ignored decoded-option mutant caught on native: case 0 line 2: port "/tmp/adamic-gate/TestDecodedOptionsAndMutant526419110/001/catch.ts:1:26", Go "fixed\ttry { work(); } catch(e) {}\\u000a"
```
