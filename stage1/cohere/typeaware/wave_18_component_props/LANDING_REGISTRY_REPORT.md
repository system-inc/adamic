Built: rebased 17 completed ports and four parked analysis claims onto the lint registry migration; no rule implementation changes.
Commits: tested 9120b267227d8b5a92344d76b324db32d06ef547; area b46914832d70e00847d82d5d221ab7bb24040c53 includes fetched main c7991b900362796aefd111474e65eb5398e91953.
Checks: six owned suites pass together in 897.723s; rebuilt standalone groups, sanitizers, released handles, parked reporters, registry, shared harness, uncached Node and vet pass.
Mutants: 29 Go-byte mutations, six stale-registry checks, three removed-refusal checks, parser-count, Node one-byte and emitted-JavaScript mismatch checks caught their mutations.
Not covered: four native HIR/SSA/capture claims remain parked; full gate and its 17 external correctness checks not run; no full shared native JSX rule integration; zero unclaimed ranked rules.

The area advanced by merging lint-rules/legacy at 91879f86, making the registry the shared dispatch path. Rebased cleanly, retained all incoming shared changes, and edited only owned report/evidence/claim status. No protected compiler, shared generator or shared harness file was edited. Only codex/typeaware-wave-18 is pushed; main and area integration remain with integration.

Setup: bash cloud/setup.sh succeeded. Go, clang, Node, submodules ready in 0s each; cache warm in 58s; total 58s. nproc=5, CPU quota=4, memory=17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Every test shell sourced /workspace/adamic-tools/env.sh. Removed only 84 reproducible ELF/archive files totaling 2,248,888,000 bytes from explicitly named owned scratch groups, either completed or not yet started. Sources, finding output and logs were retained. This avoided sanitizer disk exhaustion.

`go test ./stage1/cohere/typeaware -run '^TestWave18' -count=1 -timeout=30m -v` passed all six together in 897.723s: constructor 91.07s, core 89.19s, preference 102.89s, React prepared cores 470.49s, original suite 91.38s, isolated title suite 52.69s. Findings, fixes and suggestions compare byte for byte against pinned independent Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Normal/sanitizer corpora and controls passed; each byte-only mutation compiled successfully, exited zero and had empty stderr.

Both standalone validators were bootstrapped again by removing native-asan before validation, rebuilding stage0, Go oracle, bridge archives and native programs. JSX validate.py/validate_bridge.py passed 130 valid controls, four option profiles, all three per-rule mutations, raw numeric-question mutation, released handles and both normal/sanitized corpora. Component-property validators passed 73 controls, six profiles, two per-rule mutations, numeric-question mutation, released handles and both corpora. The numeric questions differed from Go at bytes 53 and 49; stale registry mutants violated required exit 70. Corpora produce zero findings for these groups, so positive controls and compiling mutants remain essential.

Pinned inputs were supplied explicitly: TypeScript source /workspace/wave-18-typescript; title native parser slice /workspace/wave-18-jsx-slice; compiler config /workspace/wave-18-typescript/src/compiler/tsconfig.json and frozen /tmp/wave-18-compiler.manifest (77 roots); repository config /workspace/adamic/tsconfig.json and frozen /tmp/wave-18-repository.manifest (287 roots). Corpus environment variables were set for both standalone validators. Input manifests and complete output journals are archived.

`validate_partial.py --scratch /workspace/wave18-area/partial` passed reporter bytes, sanitizer, three byte mutations, three removed-refusal mutations and the compiling parser-count mutation. Production source-analysis claims remain parked: set-state-in-effect, set-state-in-render, static-components, jsx-no-constructed-context-values. Blockers remain native HIR/SSA/capture and callback/return escape analysis; JSX syntax support is present. React prepared cores use Go-prepared input; title retains its private native JSX parser slice. No claim of complete shared native driver integration or production source-analysis parity is made.

`go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1 -timeout=10m` passed (0.440s/0.437s), including descriptor validation. Five shared harness tests, TestEmittedJavaScriptMismatch, TestDotARename, TestCompleteSuggestionSerialization, TestSuggestionAlongsideAutomaticFix and TestWitnessScriptKind, passed in 166.190s. Their Go/Node/emitted-JavaScript/native complete-record paths remain green on the migrated registry. The expected inner emitted-JavaScript failure is caught by its passing outer test.

Filtered `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle` passed in 6.087s. Filter covered TestTheOracleCatchesOneByte, native-agrees-with-Node closures, method_closures, generic_functions, regions, regions_throw, inherited_static_field_read, runtime_last_index_of and five proven_* fixtures, plus TestRuntimeLast and TestProven. Prefix selection also includes regexp_cycle_closures. Native 43 misses, Node 29 misses, zero hits. `go vet ./...` passed with empty output. No selected test skipped. The full gate and its 17 external compiler/postcss/graphql/parser correctness checks were not run and are not claimed green; no skip or requirement was weakened or removed.

Three alternating process medians in seconds, under concurrent validation load: JSX compiler native 1.708158 / Go 0.343757, repository 0.281949 / 0.155983. Property pair compiler native 4.444437 / Go 0.396175, repository 0.653215 / 0.175569. All suite timing rows are archived. React prepared native aggregate 0.018880s versus full Go 0.289487s and Go preparation 0.280166s uses different input pipelines, so is not an end-to-end speed comparison. No isolated speed improvement claim is made.

Refreshed actual origin audit: 462 heads, 33 distinct claim Markdown blobs, 197 ranked rules, zero eligible unclaimed entries. No new claim made. Current main and area are included by ancestry; only the owned branch is updated remotely. Earlier reports describe the 17 ports and their source/integration limits. No new rule, listener descriptor or regex matcher was introduced this turn.

All mutation/refusal observations follow. Full subprocess logs, stdout/stderr and journals are compressed under validation-registry. Failed compilation is not counted as a successful mutation test.

```text
wave18-registry-oracle.log:     wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1100
wave18-registry-oracle.log:     wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11255
wave18-registry-oracle.log:     wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13042
wave18-registry-oracle.log:     wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
wave18-registry-oracle.log:     wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-registry-oracle.log:     wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 560
wave18-registry-oracle.log:     wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 41361
wave18-registry-oracle.log:     wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-registry-oracle.log:     wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1593
wave18-registry-oracle.log:     wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 32342
wave18-registry-oracle.log:     wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 21969
wave18-registry-oracle.log:     wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 32254
wave18-registry-oracle.log:     wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
wave18-registry-oracle.log:     wave_18_react_cores_test.go:176: render-unconditional: exit 0, empty stderr; independent complete-byte comparison catches byte 400
wave18-registry-oracle.log:     wave_18_react_cores_test.go:176: effect-setter: exit 0, empty stderr; independent complete-byte comparison catches byte 1507
wave18-registry-oracle.log:     wave_18_react_cores_test.go:176: static-creator: exit 0, empty stderr; independent complete-byte comparison catches byte 50
wave18-registry-oracle.log:     wave_18_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3067
wave18-registry-oracle.log:     wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-registry-oracle.log:     wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4344
wave18-registry-oracle.log:     wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1662
wave18-registry-oracle.log:     wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7707
wave18-registry-oracle.log:     wave_18_test.go:196: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-registry-oracle.log:     wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 50
wave18-registry-jsx.log: fragments: compiled, exit 0, empty stderr; Go byte comparison catches byte 446
wave18-registry-jsx.log: undef: compiled, exit 0, empty stderr; Go byte comparison catches byte 8288
wave18-registry-jsx.log: adjacent: compiled, exit 0, empty stderr; Go byte comparison catches byte 27811
wave18-registry-jsx.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-registry-jsx-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 53
wave18-registry-jsx-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-registry-component.log: placement: compiled, exit 0, empty stderr; Go byte comparison catches byte 5981
wave18-registry-component.log: style: compiled, exit 0, empty stderr; Go byte comparison catches byte 321
wave18-registry-component.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-registry-component-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 49
wave18-registry-component-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-registry-partial.log: set_state_in_effect: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-registry-partial.log: set_state_in_render: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-registry-partial.log: static_components: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-registry-partial.log: base JSX parser: two nodes; compiling count mutant caught by expected output
```
