Built: landing rebase of all wave-18 work onto main b8fb957aa, with all existing oracle groups re-green.
Commits: tested rebased tip ae1ceca5abf421d09d59f05656424707a9b9500b; previous pushed tip 2cfdc1e5c; evidence commit follows.
Checks: six wave-18 suites pass together; both freshly rebuilt standalone groups, sanitizers, bridge, parked reporters, production Go, vet and filtered uncached Node pass.
Mutants: 29 byte-only rule/fact/reporter mutations, six stale-registry mutations, three removed-refusal mutations and the Node one-byte mutation caught.
Not covered: four parked analysis rules, shared-driver/native-JSX-parser integration, full root gate or expanded populations. The refreshed audit finds no unclaimed ranked rule.

The landing-first cap makes this rebase the unit. Fetching all origin branches
found main advanced from c01907a70 to b8fb957aa. The 37-commit rebase completed
without conflicts and preserved incoming changes. Only codex/typeaware-wave-18
is a push target. An exact lease on 2cfdc1e5cd31ed9233917bb3cd79de787952b3be
protects the requested rebase push. No new rule claim was made.

The refreshed audit examines 587 origin refs and 33 distinct claim Markdown
blobs against the 197-rule VOLUME_REPORT ranking and baseline ports. No
unclaimed ranked rule remains. Full claim texts and origins, selection summary
and command output are preserved. Existing named rule.json declarations remain
consistent with Ahra's registry contract. Shared harness 41eb6eab2 preserves
the seven-argument report and wire protocol; the landing target here is main
b8fb957aa. No shared harness, generator or protected compiler edit was made.

## Commands and results

    git fetch origin '+refs/heads/*:refs/remotes/origin/*'
    git rebase origin/main
    bash cloud/setup.sh > /tmp/wave18-b8fb-setup.log 2>&1
    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/typeaware -run '^TestWave18' -count=1 -timeout=30m -v > /tmp/wave18-b8fb-oracle.log 2>&1
    python3 stage1/cohere/typeaware/wave_18_jsx/validate.py --scratch /workspace/wave18-jsx-reproduce
    python3 stage1/cohere/typeaware/wave_18_jsx/validate_bridge.py --scratch /workspace/wave18-jsx-reproduce
    python3 stage1/cohere/typeaware/wave_18_component_props/validate.py --scratch /workspace/wave18-component
    python3 stage1/cohere/typeaware/wave_18_component_props/validate_bridge.py --scratch /workspace/wave18-component
    python3 stage1/cohere/typeaware/wave_18_react_partial/validate_partial.py --scratch /workspace/wave18-b8fb/partial
    go test ./bridge/tsgo/checker -count=1 -timeout=10m
    go -C cohere test ./internal/lint/rules/react -run '^(TestJsxFragments|TestJsxNoUndef|TestNoAdjacentInlineElements|TestStaticPropertyPlacement|TestStylePropObject)' -count=1 -timeout=10m -v
    go vet ./...
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' -count=1 -timeout=15m -v

All test stdout and stderr were redirected directly to the archived log files.
Environment: ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript,
ADAMIC_WAVE18_JSX_SLICE=/workspace/wave-18-jsx-slice, compiler config
/workspace/wave-18-typescript/src/compiler/tsconfig.json and repository config
/workspace/adamic/tsconfig.json. Manifest variables point to
/tmp/wave-18-compiler.manifest and /tmp/wave-18-repository.manifest. Six suite
artifact variables point under /workspace/wave18-b8fb to base, core, constructor,
preference, react and title. Both standalone native-asan files were removed
before validation, forcing fresh stage0, Go oracle, bridge archive and native
normal/sanitized builds. Source and older validation logs were retained while
84 obsolete compiled c019 scratch files were removed to recover disk space.

| Suite | Seconds |
| --- | ---: |
| TestWave18ConstructorAgreementAndMutants | 79.76 |
| TestWave18CoreAgreementAndMutants | 84.58 |
| TestWave18PreferenceAgreementAndMutants | 101.29 |
| TestWave18ReactCores | 457.19 |
| TestWave18AgreementAndMutants | 98.92 |
| TestWave18TitleWithJsxSlice | 51.39 |

Both frozen populations, 77 compiler and 287 repository sources, agree on full
findings/fixes/suggestions normally and under sanitizers. JSX retains 130 valid
controls and 110 default findings, with option counts 96/107/93. The final
property pair retains 73 valid controls, 45 default findings and option counts
46/43/25/26/46. All standalone native/sanitizer stderr is empty. Prepared-input
React cores retain 91 valid controls, 12 batches and 48 findings; those tests
do not establish native source-to-HIR lowering. Parked reporters retain their
four-finding comparison and explicit refusal contracts.

Checker PASS 0.131s, production Go PASS 0.138s, vet exit 0, filtered uncached
Node PASS 1.342s (native 19 misses, Node 13 misses, zero hits). Setup prints Go
1s, clang 1s, Node 1s, submodules 1s, cache warm 86s, done 86s; nproc 5,
CPU quota 4. Exact setup lines are authoritative if rounding differs.

## Every mutant observation

All 29 byte-only mutations compile, exit 0 and have empty stderr; only full
independent Go bytes catch them. Stale-registry and removed-refusal mutations
violate their required contracts. The Node suite includes its one-byte mutant.

- wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1100
- wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11255
- wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13042
- wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
- wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 47
- wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 560
- wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 41361
- wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil>
- wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1593
- wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 32342
- wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 21969
- wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 32254
- wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
- wave_18_react_cores_test.go:176: render-unconditional: exit 0, empty stderr; independent complete-byte comparison catches byte 400
- wave_18_react_cores_test.go:176: effect-setter: exit 0, empty stderr; independent complete-byte comparison catches byte 1507
- wave_18_react_cores_test.go:176: static-creator: exit 0, empty stderr; independent complete-byte comparison catches byte 50
- wave_18_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3067
- wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 47
- wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4344
- wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1662
- wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7707
- wave_18_test.go:195: released registry mutant caught: expected stale handle panic, got <nil>
- wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 50
- fragments: compiled, exit 0, empty stderr; Go byte comparison catches byte 446
- undef: compiled, exit 0, empty stderr; Go byte comparison catches byte 8288
- adjacent: compiled, exit 0, empty stderr; Go byte comparison catches byte 27811
- PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
- numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 53
- stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
- placement: compiled, exit 0, empty stderr; Go byte comparison catches byte 5981
- style: compiled, exit 0, empty stderr; Go byte comparison catches byte 321
- PASS: complete records, options, byte-only rule mutants, corpora and sanitizers
- numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte 49
- stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant
- set_state_in_effect: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
- set_state_in_render: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
- static_components: rendering mutant caught by Go bytes; removed refusal caught by required exit 70

## Native time against Go

Three alternating whole-process rounds per source population give these median
seconds. Standalone groups ran concurrently with the six-suite landing check,
so these are observations under concurrent load, not isolated attribution.

| Suite/corpus/backend | Median seconds |
| --- | ---: |
| wave_18_constructor_test.go/compiler/go | 0.322869 |
| wave_18_constructor_test.go/compiler/native | 1.654201 |
| wave_18_constructor_test.go/repository/go | 0.155295 |
| wave_18_constructor_test.go/repository/native | 0.261209 |
| wave_18_core_test.go/compiler/go | 0.373993 |
| wave_18_core_test.go/compiler/native | 2.141404 |
| wave_18_core_test.go/repository/go | 0.165141 |
| wave_18_core_test.go/repository/native | 0.368041 |
| wave_18_preference_test.go/compiler/go | 0.379050 |
| wave_18_preference_test.go/compiler/native | 2.020467 |
| wave_18_preference_test.go/repository/go | 0.129202 |
| wave_18_preference_test.go/repository/native | 0.238964 |
| wave_18_test.go/compiler/go | 0.310981 |
| wave_18_test.go/compiler/native | 1.496911 |
| wave_18_test.go/repository/go | 0.151037 |
| wave_18_test.go/repository/native | 0.252405 |
| wave_18_title_test.go/compiler/go | 0.305998 |
| wave_18_title_test.go/compiler/native | 1.502792 |
| wave_18_title_test.go/repository/go | 0.135088 |
| wave_18_title_test.go/repository/native | 0.226033 |

| Standalone group/corpus | Native | Go |
| --- | ---: | ---: |
| JSX/compiler | 1.871347 | 0.325685 |
| JSX/repository | 0.314448 | 0.152329 |
| Property/compiler | 4.049208 | 0.309935 |
| Property/repository | 0.658745 | 0.152978 |

Native remains slower. Prepared-input React native totals 0.018359s, full Go
source analysis 0.292612s and fixture preparation 0.268285s perform different
work and are not an end-to-end speed comparison.

The parked rules are set-state-in-effect, set-state-in-render and
static-components (native HIR/SSA/gates), plus jsx-no-constructed-context-values
(callback/callee return and capture/escape analysis). #dnv6f2c owns those analysis
ports; JSX integration is landing on area/stage1-lint. Raw JSX syntax and binding
facts continue to use typescript-go through the bridge. Older rules retain their
previously reported string-kind parser dependency. The full root gate, arbitrary
JSX projects, suppression/edit application and expanded populations were not run.
No further unclaimed ranked rule is available, so this unit stops after landing.
