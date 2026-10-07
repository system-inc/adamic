Built: rebased all 17 completed ports and four parked claims onto current main; retained lint-area ancestry without changing owned rule code.
Commits: tested 63e185a6b0853a5ea829a5bc0c9340c5cebc9c0e; main 71d7e491b3c9724f7a0e2ee754592149e7f9790b and lint area bb2ece564842c4b2f909b9f75c27e74c2efa4f29 both explicitly verified ancestors.
Checks: five Go suites 497.723s, prepared React 583.686s, standalone native/Go comparisons, sanitizers, released handles, parked reporters, bridge/registry, seven harness tests 303.181s, uncached Node fixtures 4.152s and vet passed.
Mutants: all per-rule byte-only mutations and bridge, released-handle, refusal, parser-count, suggestion, emitted-JavaScript and decoded-option mutations caught by their intended comparisons.
Not covered: four native HIR/SSA/capture claims parked; full gate and 17 external checks unrun; full shared native JSX integration not claimed; 575 live origin heads, 42 claim blobs, 197 rules, zero unclaimed.

Main advanced only under stage3. The rebase retained all incoming changes. A no-content merge retains the divergent lint area's ancestry. Neither main nor an area branch is pushed. Only owned report/evidence and claim status are changed beyond the rebase.

Setup succeeded: Node ready 0.023s, Go 0.024s, submodules 0.074s, verified installed Markdown dependencies 0.075s (step 0.008s), clang 0.171s, Go build 24.000s, test binaries deferred 24.141s, cache warm 24.142s, total 24.172s. nproc=5, cgroup quota=4 CPUs, memory=17.6 GB. Every test shell sourced /workspace/adamic-tools/env.sh.

Commands: go test ./stage1/cohere/typeaware -run '^TestWave18(AgreementAndMutants|ConstructorAgreementAndMutants|CoreAgreementAndMutants|PreferenceAgreementAndMutants|TitleWithJsxSlice)$' -count=1 -timeout=30m -v; separate TestWave18ReactCores with the same flags. Both processes returned zero. Compiler and repository configs/manifests, TypeScript source and private JSX parser slice were supplied as in LANDING_MAIN_REPORT.md. Independent Go cohere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db. Complete findings, edits and suggestions match on controls and both corpora, normally and under sanitizers. Every owned per-rule byte mutant compiles, exits zero and emits empty stderr before comparison catches it.

Standalone wave_18_jsx/validate.py and validate_bridge.py and wave_18_component_props/validate.py and validate_bridge.py passed using their existing scratch trees and explicit compiler/repository config and manifest environment variables. Stage 3 changes do not alter their compiler/runtime inputs; existing bootstrap artifacts were reused. JSX covers 130 controls/four profiles; property rules cover 73 controls/six profiles. Both include complete records, corpora, sanitizers and released handles. Parked validate_partial.py passed reporting/sanitizer/refusal checks and parser-count mutation. Production native HIR/SSA/capture and callback-return escape analysis remain unavailable; Go-prepared React input is not native source analysis.

Bridge/checker and registry tests passed in 0.255s and 0.192s. An initial command named nonexistent ./stage1/cohere/typeaware/bridge and failed package setup; the corrected ./bridge/tsgo/checker run passed. go vet ./... returned zero with empty output. Seven shared lint harness tests passed, including configured witnesses and decoded options; the intentional inner emitted-JavaScript failure is caught by its passing outer test. No shared harness or guard was changed.

Filtered uncached Node fixture run passed with native misses=60, Node misses=41, zero hits. The first subtest selector omitted internal/oracle/testdata and ran only the one-byte mutant; the corrected selector includes fixture paths and passed. Both logs are retained. No selected final check skipped; full repository gate and its 17 external compiler/postcss/graphql/parser checks were not selected, not waived and are not claimed green.

Three alternating process medians under concurrent test load, seconds: JSX compiler native 1.747233 / Go 0.355648, repository 0.317912 / 0.215034; property compiler 4.183598 / 0.372585, repository 0.701001 / 0.223474. Native remains slower. Prepared React native 0.018016 versus full Go 0.384115 and Go preparation 0.393021 uses different pipelines and is not an end-to-end speed comparison. Other timings and mutation offsets are retained in complete compressed logs under validation-stage3-main.

Refreshed all-origin claims audit found zero eligible names. No new rule or claim was added. Four named analysis claims remain parked. Complete mutation observations follow.

```text
oracle:     wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1100
oracle:     wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11255
oracle:     wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13042
oracle:     wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
oracle:     wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 47
oracle:     wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 560
oracle:     wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 41361
oracle:     wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil> 
oracle:     wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1593
oracle:     wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 32342
oracle:     wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 21969
oracle:     wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 32254
oracle:     wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
oracle:     wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 47
oracle:     wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4344
oracle:     wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1662
oracle:     wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7707
oracle:     wave_18_test.go:196: released registry mutant caught: expected stale handle panic, got <nil> 
oracle:     wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 50
react:     wave_18_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3067
jsx: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
jsx-bridge: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
jsx-bridge: PASS: numeric query mutation and released-handle contract
component: PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
component-bridge: stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
component-bridge: PASS: numeric query mutation and released-handle contract
partial: set_state_in_effect: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
partial: set_state_in_render: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
partial: static_components: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
partial: base JSX parser: two nodes; compiling count mutant caught by expected output
harness:     harness_test.go:54: ordinary comparison rejected clean-running emitted JavaScript mutant:
harness:     harness_test.go:140: second suggestion edit mutant caught on Node: case 0 line 8: port "suggestion-edit\t9 11\t", Go "suggestion-edit\t9 10\t"
harness:     harness_test.go:140: second suggestion edit mutant caught on emitted JavaScript: case 0 line 8: port "suggestion-edit\t9 11\t", Go "suggestion-edit\t9 10\t"
harness:     harness_test.go:140: second suggestion edit mutant caught on native: case 0 line 8: port "suggestion-edit\t9 11\t", Go "suggestion-edit\t9 10\t"
harness:     registration_test.go:288: ignored decoded-option mutant caught on Node: case 0 line 2: port "/tmp/adamic-gate/TestDecodedOptionsAndMutant3180771881/001/catch.ts:1:26", Go "fixed\ttry { work(); } catch(e) {}\\u000a"
harness:     registration_test.go:288: ignored decoded-option mutant caught on native: case 0 line 2: port "/tmp/adamic-gate/TestDecodedOptionsAndMutant3180771881/001/catch.ts:1:26", Go "fixed\ttry { work(); } catch(e) {}\\u000a"
```
