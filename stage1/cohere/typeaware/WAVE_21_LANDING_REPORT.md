Built: rebased the owned wave-21 branch onto main; retained ten source rule ports and three partial native React validator cores.
Commits: main e8ba3d5d; rebased implementation 85be14b8 and pre-report tip e28b55eb; old-to-new commit mapping is archived.
Checks: six wave-21 suites PASS 880.182s; five foundation suites PASS 579.690s; checker PASS 0.558s; Node PASS 36.207s; vet PASS.
Mutants: every rule and guard mutant reran; exact mutations and observed catches are in validation-wave-21-landing/mutants.md and the archived logs.
Not covered: full React source ports, native HIR lowering and graph transforms, compilation-unit/memo annotations, JSX/parser refusals, Go two-creator nondeterminism, optional inherited corpus runs, or the full repository gate. No new rules claimed.

The landing cap was applied to codex/typeaware-wave-21, the only branch pushed by this unit. Main was e8ba3d5d81de4d3773c723914fccd4c76248b965. All 21 commits rebased without conflicts; git range-diff reports unchanged patches. No main or area branch was updated. The explicit user instruction to rebase and push this owned branch authorizes the history rewrite; the push uses a lease fixed to its observed old tip e9ec024c3213ec7c9ad967d39029e4dd004abb9a.

The source ports are no-mixed-enums, collection-misuse, discarded-outcome, discarded-pure-result, process-exit-after-output, uncleared-race-timeout, require-blocking-standard-streams, no-obj-calls, no-object-constructor and no-promise-executor-return. Each has a qualifying native mutant and complete finding/fix/suggestion comparisons, normal and sanitizer executions, and released-handle checks. Wave-21 tests ran against the frozen 77-root compiler and 287-root repository corpora. Existing documented parser exclusions are retained and are not described as passing inputs.

The three React cores match independent Go on 91 valid prepared-HIR controls and both prepared corpora under normal and ASAN builds. Source lowering, compilation gates and memo annotations are still supplied by the test-only Go provider. They are not native source ports. The unchanged Go static-components oracle also has the separately archived two-creator phi ambiguity. Existing claims remain held; no new claim was made while this source-level work remains incomplete. See WAVE_21_REACT_CORE_REPORT.md for the exact boundary and original evidence.

Fresh quiet wall times, seconds (native / Go):

| Suite | Compiler | Repository |
| --- | --- | --- |
| Constructors and promise executor | 3.139418 / 0.354631 | 0.425405 / 0.128589 |
| Collection and discarded results | 2.617588 / 0.764111 | 0.343483 / 0.154500 |
| Process and timer rules | 3.104731 / 0.318935 | 0.364910 / 0.117028 |

React prepared-input native execution was 0.017485s versus 0.255472s for full Go execution; Go fixture preparation was 0.277096s. These different input pipelines do not establish an end-to-end native speed comparison.

Setup completed in 119s: Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warmup 119s. nproc was 5; cgroup quota was four CPUs. The tool environment is /workspace/adamic-tools/env.sh.

Commands (output redirected directly to archived logs):

```bash
bash cloud/setup.sh > /workspace/wave21-landing-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave21' -count=1 -timeout=30m -v > /workspace/wave21-landing-oracles.log 2>&1
go test ./stage1/cohere/typeaware -run '^Test(TypeAwareAgreementAndMutants|PinnedTypeFlags|VolumeAgreementAndMutants|VolumeConfigGuardAndMutant|CoverageAgreementAndMutants)$' -count=1 -timeout=20m -v > /workspace/wave21-landing-foundation.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /workspace/wave21-landing-checker.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave21-landing-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|method_closures|generic_functions|number_parsing)\.a$' -count=1 -timeout=10m -v > /workspace/wave21-landing-node.log 2>&1
```

The inherited foundation run exercised controls, ASAN, mutants and framing/configuration guards, but its optional compiler/repository environments were not enabled. Wave-21's own compiler/repository comparisons did run. The Node command exercised 31 native and 21 Node cache misses as recorded by its actual log. No full repository gate was run. All fresh logs and the unchanged-patch rebase map are in validation-wave-21-landing; the original per-rule reports and source artifacts remain available.
