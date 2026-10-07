Built: rebased all 17 completed native ports and four parked claims onto lint area d65a8f931; no rule implementation changes.
Commits: validated dc87cb74d197aaddeaa64104c7c9ac6e5b649edd on area d65a8f931c98655936ae04c6899f38f14862b73e, with fetched main 39638d9e278d38bb5aeae887f46d55a70e47aaad included.
Checks: six suites pass together in 862.181s; standalone groups, sanitizers, released handles, parked reporters, shared harness, uncached Node and vet pass.
Mutants: 29 Go-byte mutations, six stale-registry checks, three removed-refusal checks, parser-count, Node one-byte and emitted-JavaScript mismatch checks caught their mutations.
Not covered: four HIR/SSA/capture claims remain parked; full gate and full shared native JSX rule integration remain outside this unit; no unclaimed ranked rule remains.

The area advanced from 7481e032 to d65a8f931 by merging the lint runtime profile work. Rebased the same owned branch, retaining those heap release and string comparison/search changes. Main did not advance in this fetch. Nothing was pushed to main or any area branch. No shared harness, registration generator or protected compiler file was edited.

Setup: bash cloud/setup.sh succeeded, with Go/clang/Node/submodules each ready in 0s, cache warm in 114s, total 114s. nproc=5, cgroup CPU quota=4, memory=17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0; each execution sourced /workspace/adamic-tools/env.sh. Removed 102 obsolete compiled ELF/archive artifacts (2,086,943,575 bytes) from named old wave-18 scratch directories, preserving sources/logs, before sanitizer output ran out of space.

Commands and results:

- `go test ./stage1/cohere/typeaware -run '^TestWave18' -count=1 -timeout=30m -v`: all six pass together, 862.181s. Compiler config /workspace/wave-18-typescript/src/compiler/tsconfig.json, frozen compiler manifest /tmp/wave-18-compiler.manifest (77 roots); repository config /workspace/adamic/tsconfig.json, frozen repository manifest /tmp/wave-18-repository.manifest (287 roots). Dedicated title parser slice /workspace/wave-18-jsx-slice. Complete findings, fixes and suggestions match pinned independent Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, normally and under sanitizers.
- Rebuilt standalone JSX and property-pair groups after removing only their native-asan bootstrap sentinels; their validators rebuilt stage0, Go oracle, bridge archives and native binaries. The first control invocations omitted corpus environment variables, so they covered controls/options/mutations only. Repeated both validators with the compiler/repository config and manifest environment explicitly set: both corpora and sanitizer comparisons pass. No implementation changed between runs. JSX has 130 valid controls/four option profiles; property pair 73 controls/six profiles. Corpora produce zero findings, held by positive controls and per-rule compiling mutations.
- Both `validate_bridge.py` checks pass: numeric bridge mutations alter Go findings at bytes 53 and 49; keeping released handles incorrectly valid is caught by the required exit-70 refusal.
- `validate_partial.py --scratch /workspace/wave18-area/partial` passes reporting, sanitizer, three byte mutations, three removed-refusal mutations and the parser-count mutation. It explicitly reports production source-analysis ports blocked. JSX parser support is present; remaining blockers are native HIR/SSA/capture and callback/return escape analysis. Claims set-state-in-effect, set-state-in-render, static-components and jsx-no-constructed-context-values stay parked.
- `go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1 -timeout=10m`: pass.
- Shared harness tests `TestEmittedJavaScriptMismatch`, `TestDotARename`, `TestCompleteSuggestionSerialization`, `TestSuggestionAlongsideAutomaticFix`, `TestWitnessScriptKind`: pass in 152.022s. Expected inner failure from the emitted-JavaScript mutation is caught by the passing outer test.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle` with filtered `TestTheOracleCatchesOneByte`, native-agrees-with-Node closures/method_closures/generic_functions/regions/regions_throw/inherited_static_field_read/runtime_last_index_of, and `TestRuntimeLast`: pass in 15.080s, native 28 misses and Node 19 misses, zero hits. The new runtime search fixture and independent Node search check pass. Prefix selection also includes regexp_cycle_closures.
- `go vet ./...`: pass, empty output.

Three alternating process medians, seconds, under concurrent validation load: JSX compiler native 2.081060 / Go 0.401532; repository 0.364323 / 0.204683. Property pair compiler native 4.230093 / Go 0.349709; repository 0.637074 / 0.188853. Title compiler native 1.359226 / Go 0.300925. All individual timing rows are archived. React prepared-core aggregate native 0.019142s versus full Go 0.270206s and preparation 0.279998s uses different input pipelines and is not an end-to-end speed comparison. These are observed timings, not an isolated speed improvement claim.

The title still uses its private native JSX parser slice; prepared React cores still use Go-prepared rows. This run does not claim all 17 ports are integrated with the new shared native driver, nor full production analysis for the four parked claims. Older owned rules retain their pre-existing parser APIs. No new rule was added, so no new listener descriptor or regex translation was required. No hand-rolled regex matcher was introduced.

Actual origin heads refreshed explicitly because the default fetch tracks only a limited ref set: 433 heads, 33 distinct claims, 197 ranked rules, zero remaining unclaimed entries. The current ranking and baseline port flags are the preserved full selection population, checked against every actual head's claim blobs. Audit JSON and actual remote head list are archived. No new claim was made.

Every observed mutation and refusal row follows; full subprocess logs, canonical outputs and journals are compressed in validation-d65a. Successful byte mutations compile, exit 0 and have empty stderr; killed compiler errors do not count.

```text
wave18-refresh-oracle.log:     wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1100
wave18-refresh-oracle.log:     wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11255
wave18-refresh-oracle.log:     wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13042
wave18-refresh-oracle.log:     wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
wave18-refresh-oracle.log:     wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-refresh-oracle.log:     wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 560
wave18-refresh-oracle.log:     wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 41361
wave18-refresh-oracle.log:     wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-refresh-oracle.log:     wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1593
wave18-refresh-oracle.log:     wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 32342
wave18-refresh-oracle.log:     wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 21969
wave18-refresh-oracle.log:     wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 32254
wave18-refresh-oracle.log:     wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
wave18-refresh-oracle.log:     wave_18_react_cores_test.go:176: render-unconditional: exit 0, empty stderr; independent complete-byte comparison catches byte 400
wave18-refresh-oracle.log:     wave_18_react_cores_test.go:176: effect-setter: exit 0, empty stderr; independent complete-byte comparison catches byte 1507
wave18-refresh-oracle.log:     wave_18_react_cores_test.go:176: static-creator: exit 0, empty stderr; independent complete-byte comparison catches byte 50
wave18-refresh-oracle.log:     wave_18_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3067
wave18-refresh-oracle.log:     wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 47
wave18-refresh-oracle.log:     wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4344
wave18-refresh-oracle.log:     wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1662
wave18-refresh-oracle.log:     wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7707
wave18-refresh-oracle.log:     wave_18_test.go:196: released registry mutant caught: expected stale handle panic, got <nil> 
wave18-refresh-oracle.log:     wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 50
wave18-refresh-jsx-corpora.log: fragments: compiled, exit 0, empty stderr; Go byte comparison catches byte 446
wave18-refresh-jsx-corpora.log: undef: compiled, exit 0, empty stderr; Go byte comparison catches byte 8288
wave18-refresh-jsx-corpora.log: adjacent: compiled, exit 0, empty stderr; Go byte comparison catches byte 27811
wave18-refresh-jsx-corpora.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-refresh-jsx-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 53
wave18-refresh-jsx-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-refresh-component-corpora.log: placement: compiled, exit 0, empty stderr; Go byte comparison catches byte 5981
wave18-refresh-component-corpora.log: style: compiled, exit 0, empty stderr; Go byte comparison catches byte 321
wave18-refresh-component-corpora.log: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
wave18-refresh-component-bridge.log: numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 49
wave18-refresh-component-bridge.log: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
wave18-refresh-partial.log: set_state_in_effect: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-refresh-partial.log: set_state_in_render: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-refresh-partial.log: static_components: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
wave18-refresh-partial.log: base JSX parser: two nodes; compiling count mutant caught by expected output
```
