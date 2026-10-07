Built: landing rebase of all wave-18 work onto main c01907a70, with existing oracles re-green.
Commits: tested rebased tip f619ea20dd1d3f4650532eb063effcc0ecd7b037; prior pushed tip 9f9330c9d; evidence commit follows.
Checks: six wave-18 suites pass in one run; rebuilt JSX parity, sanitizer, bridge, reporters, production Go, vet and filtered uncached Node pass.
Mutants: 26 byte-only rule/fact/reporter mutations, five stale-registry mutations, three removed-refusal mutations and the Node one-byte mutation caught.
Not covered: four parked analysis rules, native JSX parser/shared-driver integration, older-rule numeric dispatch conversion, new stage3 populations or the full root gate. No new claims.

The landing-first requirement makes this rebase the unit. The branch rebased
cleanly onto fetched origin/main c01907a70, preserving incoming main changes.
The existing implementation is held by fresh native builds and independent Go
comparisons. Only codex/typeaware-wave-18 is the push target. An exact lease on
9f9330c9de3dd6e2cc8e879ef87294e49348d5ee protects the requested rebase push.

## Reproduction and results

    git fetch origin
    git rebase origin/main
    bash cloud/setup.sh > /tmp/wave18-c019-setup.log 2>&1
    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/typeaware -run '^TestWave18' -count=1 -timeout=30m -v > /tmp/wave18-c019-oracle.log 2>&1
    python3 stage1/cohere/typeaware/wave_18_jsx/validate.py --scratch /workspace/wave18-jsx-reproduce > /tmp/wave18-c019-jsx-rebuilt.log 2>&1
    python3 stage1/cohere/typeaware/wave_18_jsx/validate_bridge.py --scratch /workspace/wave18-jsx-reproduce > /tmp/wave18-c019-jsx-bridge-rebuilt.log 2>&1
    python3 stage1/cohere/typeaware/wave_18_react_partial/validate_partial.py --scratch /workspace/wave18-c019/partial > /tmp/wave18-c019-partial.log 2>&1
    go test ./bridge/tsgo/checker -count=1 -timeout=10m
    go -C cohere test ./internal/lint/rules/react -run '^(TestJsxFragments|TestJsxNoUndef|TestNoAdjacentInlineElements)' -count=1 -timeout=10m -v
    go vet ./...
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' -count=1 -timeout=15m -v

Environment: ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript;
ADAMIC_WAVE18_JSX_SLICE=/workspace/wave-18-jsx-slice;
ADAMIC_WAVE18_COMPILER_CONFIG=/workspace/wave-18-typescript/src/compiler/tsconfig.json;
compiler/repository manifest variables point to /tmp/wave-18-compiler.manifest
and /tmp/wave-18-repository.manifest. The six artifact variables point to
/workspace/wave18-c019/base, core, constructor, preference, react and title.
JSX repository config is /workspace/adamic/tsconfig.json.

| Suite | Seconds |
| --- | ---: |
| TestWave18ConstructorAgreementAndMutants | 82.54 |
| TestWave18CoreAgreementAndMutants | 86.54 |
| TestWave18PreferenceAgreementAndMutants | 101.42 |
| TestWave18ReactCores | 465.41 |
| TestWave18AgreementAndMutants | 94.23 |
| TestWave18TitleWithJsxSlice | 51.06 |

All suites include full findings/fixes/suggestions over the frozen 77 compiler
and 287 repository sources, normally and under sanitizers. Prepared-input React
cores retain their limited scope: 91 valid controls, 12 batches and 48 findings.
The three JSX rules retain 130 valid controls, 110 default findings, and option
counts 96/107/93. Native and sanitized stderr is empty. Complete record comparison
also holds the zero fixes and suggestions of these three JSX rules.

The first JSX attempt reused earlier artifacts and is not the authoritative
landing check. Removing only /workspace/wave18-jsx-reproduce/native-asan forced
validate.py to rebuild stage0, the independent production Go oracle, the normal
and sanitized bridge archives, and both native executables. The rebuilt run and
subsequent bridge-mutant run pass. Their command journals are archived. Checker
PASS 0.130s, production Go PASS 0.091s, vet exit 0, filtered uncached Node
PASS 1.367s (native 19 misses, Node 13 misses, zero hits). Setup prints Go 0s,
clang 0s, Node 0s, submodules 0s, cache warm 40s, done 40s; nproc 5, CPU quota 4.
Obsolete compiled scratch artifacts were removed to free disk space; their
source and validation logs were retained.

## Every mutant observation

Every byte-only mutant compiles, exits 0 and has empty stderr. The stale-registry
and removed-refusal mutants violate the required released-handle/refusal
contracts. The independent Node oracle also proves its one-byte check fails.

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
- set_state_in_effect: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
- set_state_in_render: rendering mutant caught by Go bytes; removed refusal caught by required exit 70
- static_components: rendering mutant caught by Go bytes; removed refusal caught by required exit 70

## Native time against Go

Three alternating whole-process samples give these medians, in seconds.
JSX samples ran alongside the landing suites, so they are observations under
concurrent load. Native remains slower on the complete source pipelines.

| Suite/corpus/backend | Median seconds |
| --- | ---: |
| wave_18_constructor_test.go/compiler/go | 0.299979 |
| wave_18_constructor_test.go/compiler/native | 1.510739 |
| wave_18_constructor_test.go/repository/go | 0.132596 |
| wave_18_constructor_test.go/repository/native | 0.231182 |
| wave_18_core_test.go/compiler/go | 0.415984 |
| wave_18_core_test.go/compiler/native | 2.359140 |
| wave_18_core_test.go/repository/go | 0.185062 |
| wave_18_core_test.go/repository/native | 0.353990 |
| wave_18_preference_test.go/compiler/go | 0.403260 |
| wave_18_preference_test.go/compiler/native | 2.092742 |
| wave_18_preference_test.go/repository/go | 0.138831 |
| wave_18_preference_test.go/repository/native | 0.241365 |
| wave_18_test.go/compiler/go | 0.291391 |
| wave_18_test.go/compiler/native | 1.391978 |
| wave_18_test.go/repository/go | 0.122317 |
| wave_18_test.go/repository/native | 0.222911 |
| wave_18_title_test.go/compiler/go | 0.282987 |
| wave_18_title_test.go/compiler/native | 1.370941 |
| wave_18_title_test.go/repository/go | 0.133870 |
| wave_18_title_test.go/repository/native | 0.223118 |

Rebuilt JSX compiler median: native 1.821702s, Go 0.323028s.
Rebuilt JSX repository median: native 0.297296s, Go 0.169260s.
React prepared-input native 0.020121s, full Go source analysis 0.334142s and
fixture preparation 0.285700s perform different work and are not comparable
end-to-end timings.

set-state-in-effect, set-state-in-render and static-components remain parked on
native HIR/SSA/gates. jsx-no-constructed-context-values remains parked on
callee/callback return and capture/escape analysis. #dnv6f2c owns those analysis
ports; JSX integration is landing on area/stage1-lint. Raw JSX syntax and binding
facts still come from typescript-go through the C bridge. The three new JSX
rules retain numeric rule.json kinds and supplied-node callbacks. Older rules
retain their previously reported string-kind parser dependency. The announced
shared harness ab70f38d4 preserves the report wire contract; this landing target
is origin/main c01907a70. No full root gate or expanded stage3 corpus was run.
