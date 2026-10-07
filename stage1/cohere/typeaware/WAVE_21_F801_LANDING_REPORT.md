Built: rebased wave-21 onto current main f8013f0ba and re-greened all owned rule/listener oracles plus the inherited bridge foundation.
Commits: pre-rebase remote 078b38c96; rebased tip 6caedccc5; React core implementation maps to c51ec0274; all 24 patches are unchanged.
Checks: seven wave-21 suites PASS 939.403s; five foundation suites PASS 596.867s; checker PASS 0.562s; expanded Node oracle PASS 53.819s; vet PASS.
Mutants: all thirteen native rule mutants, numeric listener and adjacency-row mutants, question/handle/framing guards and prerequisite guards reran and were caught; complete ledger archived.
Not covered: full native React source ports, native HIR lowering/graph transforms, numeric handed-node integration, the future shared Diagnostic, existing parser refusals, Go phi ambiguity, optional inherited corpus runs or the full repository gate. No new claims.

The landing cap made this rebase the current unit. Main advanced from e8ba3d5d to f8013f0baac41ddc340d76f83bddde38536a8f07, bringing compiler narrowing/inheritance and map/set iterator fixes. The owned branch rebased without conflicts. git range-diff shows all 24 carried patches unchanged. This unit has pushed only codex/typeaware-wave-21. No main or area branch is updated. The explicit instruction to rebase and push the owned branch authorizes the history rewrite; the push uses a lease fixed to its observed old tip 078b38c96a653afd12b3ace4405e4f1a5cb1029e.

Fresh verification covers ten native source rule ports, three partial React validator cores, all thirteen numeric syntaxKinds exports and rule.json declarations, and the inherited 26-rule bridge foundation. Wave-21 ran on both frozen corpora, with 77 compiler roots and 287 repository roots. Findings, fixes and suggestions are compared in their complete canonical representation. Normal, ASAN/UBSAN/leak checks, native rule mutants and released-handle guards passed. Existing documented object-constructor parser exclusions are preserved rather than counted as passing inputs.

The React result is still prepared-HIR coverage: 91 valid controls, twelve batches, 48 findings, plus both prepared corpora. The test-only Go provider supplies source lowering, compilation gates and memo annotations. The native validator cores do not implement that source pipeline. The separately archived production Go two-creator phi ambiguity remains; the pinned Go source did not change in this main update. Shared numeric ParseNode/handed-node integration and the future Diagnostic model also remain outside this implementation. rule.json files remain declaration-only metadata, not registered source callbacks. No source decision logic or shared harness/generator/compiler files were edited during this landing unit.

| Wave-21 suite | Seconds | Result |
| --- | ---: | --- |
| Constructors and promise executor | 160.33 | PASS |
| Numeric/JSON listener declarations | 19.67 | PASS |
| Collection and discarded results | 80.06 | PASS |
| Process and timer rules | 93.50 | PASS |
| React prepared-HIR cores | 499.95 | PASS |
| React prerequisites | 39.18 | PASS, documented JSX refusals remain |
| Mixed enums | 46.70 | PASS |

Fresh quiet wall measurements, native / Go seconds:

| Suite | Compiler | Repository |
| --- | --- | --- |
| Constructors and promise executor | 3.187766 / 0.404857 | 0.376439 / 0.136078 |
| Collection and discarded results | 3.376900 / 0.800035 | 0.382773 / 0.163162 |
| Process and timer rules | 2.937038 / 0.310111 | 0.366844 / 0.214991 |

React prepared-input native execution was 0.017095s, full Go execution 0.314936s and Go fixture preparation 0.289364s. Different input pipelines do not establish an end-to-end native speed comparison. Measurements were taken within a gate where the independent foundation suite ran concurrently; no new throughput claim is made.

Setup completed in 119s: Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warming 119s. nproc was 5 and the cgroup quota four CPUs. Source /workspace/adamic-tools/env.sh for every shell. The new environment and configured network readiness were inspected through the cloud runtime skill; no credential values were read or printed.

Commands, each writing output directly to a log:

```bash
bash cloud/setup.sh > /workspace/wave21-f801-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave21' -count=1 -timeout=30m -v > /workspace/wave21-f801-oracles.log 2>&1
go test ./stage1/cohere/typeaware -run '^Test(TypeAwareAgreementAndMutants|PinnedTypeFlags|VolumeAgreementAndMutants|VolumeConfigGuardAndMutant|CoverageAgreementAndMutants)$' -count=1 -timeout=20m -v > /workspace/wave21-f801-foundation.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /workspace/wave21-f801-checker.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave21-f801-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestLibraryMapSetIteratorCopiesRefused$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|method_closures|generic_functions|number_parsing|library_map_set_iterator_exhausted|library_map_set_iterator_number_hash|override_same_representation|047cb0d_n_)' -count=1 -timeout=10m -v > /workspace/wave21-f801-node.log 2>&1
```

The expanded Node command passed with 77 native and 67 Node cache misses and included current-main iterator, narrowing and override fixtures plus the iterator-copy refusal guard. Inherited foundation controls, sanitizers and mutants ran, but its optional corpus environments were not enabled; wave-21's own corpus comparisons did run. No full repository gate was run. The logs, full mutant ledger and old-to-new commit map are in validation-wave-21-f801. No new claim was made while landing or the partial React source work remains unfinished.
