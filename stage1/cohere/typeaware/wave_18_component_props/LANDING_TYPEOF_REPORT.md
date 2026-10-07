Built: rebased all 17 completed ports and four parked claims onto current main's typeof/null fix through the lint area; no rule implementation changes.
Commits: tested 01123ae34c26328d7aa6c4e2383f5410be03a1ec; area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 includes fetched main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06.
Checks: six owned suites green across initial run and targeted retries; rebuilt standalone groups, normal/sanitized complete Go bytes, released handles, parked reporters, registry, shared harness, uncached Node typeof witnesses and vet pass.
Mutants: 29 Go-byte mutations, six stale-registry checks, three removed-refusal checks, parser-count, Node one-byte and emitted-JavaScript mismatch checks caught their mutations.
Not covered: four native HIR/SSA/capture claims parked; full gate and 17 external correctness checks not run; full shared native JSX rule integration not claimed; zero unclaimed ranked rules.

The incoming main change preserves null and lookup presence in native typeof classification and slots. Rebased cleanly onto the updated lint area, retaining its shared registry migration and incoming compiler/runtime changes. No protected compiler, shared generator or shared harness file was edited. Only owned report/evidence/claim status is added after rebase. No main or area push was attempted.

Setup: bash cloud/setup.sh succeeded with Go/clang/Node/submodules ready in 0s each, cache warm in 139s, total 139s. nproc=5, CPU quota=4, memory=17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Every test shell sourced /workspace/adamic-tools/env.sh. Removed 56 previous-run compiled ELF/archive artifacts (1,093,839,896 bytes) from named owned scratch groups, retaining sources and logs. Concurrent compilation still exhausted disk. Clearing Go's regenerable build cache recovered about 25 GB, but doing so while builds were linking interrupted constructor and React builds with missing cached objects. That was an avoidable build-coordination mistake, not a passing oracle result. Retained those failed logs and reran both affected suites with the cache stable. No source change was needed. Future cache clearing should happen before active builds.

Owned tests used two disjoint artifact groups concurrently: the five non-React suites in one Go process, React in another. The first five-suite invocation returned failure solely for constructor's interrupted build; core passed (641.58s), preferences passed (127.77s), original suite passed (106.40s) and title passed (78.77s). Its overall failed duration was 960.500s. The first React invocation failed at build (4.110s). Targeted constructor retry `go test ./stage1/cohere/typeaware -run '^TestWave18ConstructorAgreementAndMutants$' -count=1 -timeout=30m -v` passed in 315.151s. Targeted React retry with `-run '^TestWave18ReactCores$'` passed in 778.342s. All six therefore have green results on the same source tree; the initial combined run is not described as wholly green. Build/cache recovery and concurrent load inflate these package durations.

All complete findings, automatic edits and suggestions compare byte for byte against independent pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Pinned inputs supplied explicitly: TypeScript source /workspace/wave-18-typescript; title parser slice /workspace/wave-18-jsx-slice; compiler config /workspace/wave-18-typescript/src/compiler/tsconfig.json and frozen /tmp/wave-18-compiler.manifest (77 roots); repository config /workspace/adamic/tsconfig.json and frozen /tmp/wave-18-repository.manifest (287 roots). Both standalone validators received the corpus config/manifest environment variables. Corpora have zero findings for these groups, so positive controls and compiling per-rule mutants remain essential.

Both standalone groups rebuilt because their previous compiled outputs had been cleared. JSX validate.py and validate_bridge.py passed 130 valid controls, four option profiles, three rule mutations, numeric question mutation, released handles and both normal/sanitized corpora. Component-property validators passed 73 controls, six option profiles, two rule mutations, numeric question mutation, released handles and both corpora. Numeric mutations change Go findings at bytes 53 and 49; stale registry mutations incorrectly permit a released handle and violate required exit 70. Successful byte-only mutants compile, exit zero and emit empty stderr.

Parked validate_partial.py passed four reporter findings (2355 bytes), sanitizers, three reporter-byte mutations, three removed-refusal mutations and the compiling parser-count mutation. Production claims set-state-in-effect, set-state-in-render, static-components and jsx-no-constructed-context-values remain parked on native HIR/SSA/capture and callback/return escape analysis. JSX syntax support is present. React cores use Go-prepared input; title uses its private native parser slice. Full production source analysis and complete shared native driver integration are not claimed.

`go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1 -timeout=10m` passed (0.415s/0.457s). Shared harness filter TestEmittedJavaScriptMismatch, TestDotARename, TestCompleteSuggestionSerialization, TestSuggestionAlongsideAutomaticFix, TestWitnessScriptKind passed in 714.893s, including cache rebuild time. The emitted-JavaScript mutant's expected inner failure is caught by the passing outer test. The .ts/.a rename comparison and full suggestions/automatic-fix serialization agree across Go, Node, emitted JavaScript and native.

Filtered `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle` passed in 5.449s: one-byte mutation and native-agrees-with-Node closure/method/generic/region witnesses plus typeof_dispatch, typeof_null, typeof_null_compare, typeof_null_roll, typeof_null_slots, typeof_null_switch and typeof_string_literal. Prefix selection also includes regexp_cycle_closures. Native 40 misses, Node 27 misses, zero hits. The filter also included ^TestTypeof but no extra test of that name is claimed; new typeof source witnesses were selected explicitly. `go vet ./...` passed with empty output.

No selected final test skipped. Full gate and the 17 external compiler/postcss/graphql/parser correctness checks were not run and are not claimed green. No check, skip or input requirement was weakened or deleted. Earlier failed builds were preserved and did not count as successful mutation evidence.

Three alternating process medians in seconds under concurrent cache/build load: JSX compiler native 2.732981 / Go 0.368226, repository 1.374074 / 1.156821. Property pair compiler native 6.095863 / Go 0.619824, repository 0.902056 / 0.290823. Every individual suite timing row is archived. These are load-contaminated observations, not isolated speed measurements or a speed improvement claim. Prepared React native/full-Go/preparation timings use different pipelines and cannot establish end-to-end native speed.

Audit: 477 actual origin branches, 33 distinct claim blobs, 197 ranked rules, zero eligible unclaimed entries. Selection JSON, actual origin heads and manifests are archived. No new claim, listener descriptor or regex matcher was introduced. Only codex/typeaware-wave-18 is pushed.

All observed mutation/refusal rows follow. Complete logs and subprocess stdout/stderr/journals are compressed under validation-typeof. Initial build failures and passing retries are both retained.

```text
wave18-typeof-oracle.log:     wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-typeof-oracle.log:     wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 560
wave18-typeof-oracle.log:     wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 41361
wave18-typeof-oracle.log:     wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-typeof-oracle.log:     wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1593
wave18-typeof-oracle.log:     wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 32342
wave18-typeof-oracle.log:     wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 21969
wave18-typeof-oracle.log:     wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 32254
wave18-typeof-oracle.log:     wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
wave18-typeof-oracle.log:     wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-typeof-oracle.log:     wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4344
wave18-typeof-oracle.log:     wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1662
wave18-typeof-oracle.log:     wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7707
wave18-typeof-oracle.log:     wave_18_test.go:196: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-typeof-oracle.log:     wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 50
wave18-typeof-constructor-final.log:     wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1100
wave18-typeof-constructor-final.log:     wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11255
wave18-typeof-constructor-final.log:     wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13042
wave18-typeof-constructor-final.log:     wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
wave18-typeof-react-final.log:     wave_18_react_cores_test.go:176: render-unconditional: exit 0, empty stderr; independent complete-byte comparison catches byte 400
wave18-typeof-react-final.log:     wave_18_react_cores_test.go:176: effect-setter: exit 0, empty stderr; independent complete-byte comparison catches byte 1507
wave18-typeof-react-final.log:     wave_18_react_cores_test.go:176: static-creator: exit 0, empty stderr; independent complete-byte comparison catches byte 50
wave18-typeof-react-final.log:     wave_18_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3067
wave18-typeof-jsx.log: fragments: compiled, exit 0, empty stderr; Go byte comparison catches byte 446
wave18-typeof-jsx.log: undef: compiled, exit 0, empty stderr; Go byte comparison catches byte 8288
wave18-typeof-jsx.log: adjacent: compiled, exit 0, empty stderr; Go byte comparison catches byte 27811
wave18-typeof-jsx.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-typeof-jsx-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 53
wave18-typeof-jsx-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-typeof-component.log: placement: compiled, exit 0, empty stderr; Go byte comparison catches byte 5981
wave18-typeof-component.log: style: compiled, exit 0, empty stderr; Go byte comparison catches byte 321
wave18-typeof-component.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-typeof-component-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 49
wave18-typeof-component-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-typeof-partial.log: set_state_in_effect: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-typeof-partial.log: set_state_in_render: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-typeof-partial.log: static_components: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-typeof-partial.log: base JSX parser: two nodes; compiling count mutant caught by expected output
```
