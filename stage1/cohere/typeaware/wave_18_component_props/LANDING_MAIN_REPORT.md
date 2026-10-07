Built: rebased 17 completed ports and four parked claims onto actual current main, retained lint-area ancestry and corrected the preceding ancestry report.
Commits: tested 199d55a5d10c9cfbfe716cf3edf58dd7ded99ee7; rebased implementation df398623b335261508f18a52d858256ba45af71c; main 4e0bfda50a19c705a1aac0d9932e08483806d61c and area bb2ece564842c4b2f909b9f75c27e74c2efa4f29 are both verified ancestors.
Checks: five suites pass in 513.179s, React in 606.632s; freshly rebuilt standalone groups, sanitizers, released handles, parked reporters, bridge, registry, seven shared harness tests, uncached Node and vet pass.
Mutants: 29 Go-byte mutations, six stale-registry checks, three removed-refusal checks, parser-count, Node one-byte, emitted-JavaScript mismatch and ignored decoded-option mutation caught.
Not covered: four HIR/SSA/capture claims parked; full gate and 17 external correctness checks unrun; full shared native JSX rule integration not claimed; zero unclaimed ranked rules.

Correction: LANDING_READY_REPORT.md and the preceding final response incorrectly said the lint area included current main 4e0bfda50. An audit showed main was not an ancestor; only the area ancestry passed. A shell sequence had hidden the first ancestry check's nonzero status behind the second successful command. The prior report's test observations describe the prior area-based tree, not main. This turn rebased all 48 replayed commits onto actual origin/main, then merged origin/area/stage1-lint into the owned branch to retain its ancestry. That merge changed no files compared with the rebased tree. Both ancestry checks now run as explicit required-success operations. Earlier test logs remain intact. No protected compiler, shared generator or shared harness file was edited; incoming main runtime/developer-tools changes were retained.

Setup on actual main: bash cloud/setup.sh succeeded. Node ready 0.032s, Go ready 0.038s, submodules 0.076s, clang ready 0.209s; Markdown dependencies installed with npm ci and integrity verified, step-duration 0.982s, ready 1.055s; Go build ready 29.307s; test binaries deferred 29.429s; cache warm 29.430s; total 29.455s. nproc=5, CPU quota=4, memory=17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Every test shell sourced /workspace/adamic-tools/env.sh. No cache clearing or build recovery was needed.

All six owned TestWave18 suites ran with -count=1 -timeout=30m -v. Five non-React suites ran in one process, React in a second with disjoint scratch directories. Both returned zero: five-suite result 513.179s, React 606.632s. Complete findings, automatic edits and suggestions match independent pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, normally and under sanitizers. Positive controls and per-rule mutations hold behavior where corpora produce zero findings. All byte-only mutants compile, exit zero and emit empty stderr before Go comparison catches them.

Inputs supplied explicitly: TypeScript source /workspace/wave-18-typescript; title parser slice /workspace/wave-18-jsx-slice; compiler config /workspace/wave-18-typescript/src/compiler/tsconfig.json with frozen /tmp/wave-18-compiler.manifest (77 roots); repository config /workspace/adamic/tsconfig.json with frozen /tmp/wave-18-repository.manifest (287 roots). Both standalone validators also received corpus config/manifest environment variables. Manifests, journals and complete outputs are archived.

Both standalone groups were rebuilt on actual main by removing their native-asan bootstrap sentinels before validation. Their stage0, independent Go oracle, normal/sanitized bridge archives and native binaries were rebuilt. JSX validate.py and validate_bridge.py passed 130 valid controls/four profiles, three per-rule byte mutations, numeric-question mutation, released handles and both normal/sanitized corpora. Property pair validators passed 73 controls/six profiles, two per-rule byte mutations, numeric-question mutation, released handles and both corpora. Numeric mutations change Go bytes at 53 and 49; stale-registry mutations violate required exit 70.

Parked reporter validation passed four findings (2355 bytes), sanitizer, three reporter-byte mutations, three removed-refusal mutations and compiling parser-count mutation. set-state-in-effect, set-state-in-render, static-components and jsx-no-constructed-context-values remain parked on native HIR/SSA/capture and callback/return escape analysis. JSX syntax support is present; title retains its private native parser slice, and React cores use Go-prepared rows. Full production analysis and full shared native driver integration are not claimed.

Bridge/checker and registry checks passed (0.254s/0.153s). Shared TestEmittedJavaScriptMismatch, TestDotARename, TestCompleteSuggestionSerialization, TestSuggestionAlongsideAutomaticFix, TestWitnessScriptKind, TestOwnedWitnesses and TestDecodedOptionsAndMutant passed in 324.254s. Configured witness sidecars remain checked, and ignoring decoded options is caught on emitted JavaScript and native. The adapter guard was not weakened. Expected inner emitted-JavaScript failure is caught by the passing outer test.

Filtered ADAMIC_GATE_UNCACHED=1 internal/oracle checks passed in 7.001s: TestTheOracleCatchesOneByte and native-agrees-with-Node closures/method_closures/generic_functions/regions/regions_throw and typeof_dispatch, typeof_null, typeof_null_compare, typeof_null_roll, typeof_null_slots, typeof_null_switch and typeof_string_literal. Prefix selection also selects regexp_cycle_closures. Native 52 misses, Node 35 misses, zero hits. go vet ./... passed with empty output. No selected test skipped. Full repository gate and its 17 external compiler/postcss/graphql/parser checks were not selected and are not claimed green. No skip or input requirement was weakened or deleted.

Three alternating process medians, seconds, under concurrent compilation/validation load: JSX compiler native 2.046194 / Go 0.418661, repository 0.336313 / 0.210578; property pair compiler native 4.492242 / Go 0.428482, repository 0.744805 / 0.216548. All individual timing rows are archived. React prepared-native aggregate 0.020607s versus full Go 0.374521s and preparation 0.353613s uses different pipelines; it is not an end-to-end speed comparison. Loaded-run timings do not establish isolated speed improvement.

Audit: 566 actual origin heads, 42 distinct claim blobs, 197 ranked rules, zero eligible unclaimed entries. Main and area ancestry were required to pass in the audit script. No new claim, rule or regex matcher was introduced. Only codex/typeaware-wave-18 is pushed, with an exact prior-tip lease; integration retains main and area ownership.

All mutation/refusal observations follow. Complete logs, outputs and journals are compressed under validation-main. Prior evidence is retained with its corrected scope.

```text
wave18-main-oracle.log:     wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1100
wave18-main-oracle.log:     wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11255
wave18-main-oracle.log:     wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13042
wave18-main-oracle.log:     wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
wave18-main-oracle.log:     wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-main-oracle.log:     wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 560
wave18-main-oracle.log:     wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 41361
wave18-main-oracle.log:     wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-main-oracle.log:     wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1593
wave18-main-oracle.log:     wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 32342
wave18-main-oracle.log:     wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 21969
wave18-main-oracle.log:     wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 32254
wave18-main-oracle.log:     wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
wave18-main-oracle.log:     wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-main-oracle.log:     wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4344
wave18-main-oracle.log:     wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1662
wave18-main-oracle.log:     wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7707
wave18-main-oracle.log:     wave_18_test.go:196: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-main-oracle.log:     wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 50
wave18-main-react.log:     wave_18_react_cores_test.go:176: render-unconditional: exit 0, empty stderr; independent complete-byte comparison catches byte 400
wave18-main-react.log:     wave_18_react_cores_test.go:176: effect-setter: exit 0, empty stderr; independent complete-byte comparison catches byte 1507
wave18-main-react.log:     wave_18_react_cores_test.go:176: static-creator: exit 0, empty stderr; independent complete-byte comparison catches byte 50
wave18-main-react.log:     wave_18_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3067
wave18-main-jsx.log: fragments: compiled, exit 0, empty stderr; Go byte comparison catches byte 446
wave18-main-jsx.log: undef: compiled, exit 0, empty stderr; Go byte comparison catches byte 8288
wave18-main-jsx.log: adjacent: compiled, exit 0, empty stderr; Go byte comparison catches byte 27811
wave18-main-jsx.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-main-jsx-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 53
wave18-main-jsx-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-main-component.log: placement: compiled, exit 0, empty stderr; Go byte comparison catches byte 5981
wave18-main-component.log: style: compiled, exit 0, empty stderr; Go byte comparison catches byte 321
wave18-main-component.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-main-component-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 49
wave18-main-component-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-main-partial.log: set_state_in_effect: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-main-partial.log: set_state_in_render: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-main-partial.log: static_components: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-main-partial.log: base JSX parser: two nodes; compiling count mutant caught by expected output
wave18-main-harness.log:     harness_test.go:54: ordinary comparison rejected clean-running emitted JavaScript mutant:
wave18-main-harness.log:     harness_test.go:140: second suggestion edit mutant caught on Node: case 0 line 8: port "suggestion-edit\t9 11\t", Go "suggestion-edit\t9 10\t"
wave18-main-harness.log:     harness_test.go:140: second suggestion edit mutant caught on emitted JavaScript: case 0 line 8: port "suggestion-edit\t9 11\t", Go "suggestion-edit\t9 10\t"
wave18-main-harness.log:     harness_test.go:140: second suggestion edit mutant caught on native: case 0 line 8: port "suggestion-edit\t9 11\t", Go "suggestion-edit\t9 10\t"
wave18-main-harness.log:     registration_test.go:288: ignored decoded-option mutant caught on Node: case 0 line 2: port "/tmp/adamic-gate/TestDecodedOptionsAndMutant640696883/001/catch.ts:1:26", Go "fixed\ttry { work(); } catch(e) {}\\u000a"
wave18-main-harness.log:     registration_test.go:288: ignored decoded-option mutant caught on native: case 0 line 2: port "/tmp/adamic-gate/TestDecodedOptionsAndMutant640696883/001/catch.ts:1:26", Go "fixed\ttry { work(); } catch(e) {}\\u000a"
```
