Built persistent module units, selective declarations, stable shared identities and declaration ABI guards.
Measured before aca418919f345ecd41889e73832e6a93c95cd3db and after 36e332b885ea4c5780f8551cdb6d52a4d8b7e5a1.
Thirty interleaved timing samples passed; every after edit compiled exactly one object. Full oracle, native and cached lint suites passed; selected uncached lint parity checks passed in both modes.
Declaration disagreement compiles cleanly and fails at link; all eleven independent ownership, grouping, cache, declaration and counter mutants were caught.
No emitter or single-file output changes; Linux coverage only, and the full repository gate was not run.

Source functions and region variants follow the authoritative quoted module marker, decoded with strconv.Unquote. Filenames contain an escaped path hint and its full SHA-256 digest, so insertion cannot shift a numbered group. Source-less library functions use sixteen fixed buckets keyed by stable symbol. Constant descriptors and content-addressed strings, shapes and regex data retain independently owned definitions in thirty-two fixed buckets even when they currently have one consumer. This prevents a new consumer from moving an existing identity. Exclusive mutable caches follow their consumer module. Declaration-owned adapters follow source callees transitively, stopping at source functions. Unmarked forwarder globals stay with the coordinator; global declaration-ready writes identify and override ownership with the source module. Every shared object still has exactly one definition.

The common generated header shrinks from 898,912 bytes before to 112 bytes after; the runtime/type ABI remains in its included headers. Generated prototypes and externs are sorted and included directly in each unit's declaration preamble only when its identifier tokens reference them, including references in constant initializers and adapter bodies. Each defining unit includes its own declaration. Declaration hashes define guard symbols; constant pointer arrays retain relocations to every referenced guard. Different declarations in caller and owner therefore produce an unresolved guard at link time, without adding runtime calls.

Module initialization is extracted in the original order. The coordinator retains adamic_start, uncovered forwarder initialization, final global releases and return. Unbalanced marked scopes, cross-module locals, locals surviving into cleanup, repeated initialization modules and early returns fail explicitly. This postprocessor accepts the emitter's generated C contract, rather than arbitrary C.

The existing preprocessing snapshot cache remains keyed by exact preprocessed input, its filename, ordered Flags, compiler path/full version, platform and cache format v3. Transitive runtime/system headers and system-header line-marker flags remain in that snapshot. ADAMIC_GATE_UNCACHED=1 bypasses generated-object hits/publication; the existing runtime archive cache remains warm and correctly keyed. No additional production cache was added.

The expanded lint input has 56,904 lines and 6,841,150 bytes. There are 53 units before and 129 after. The largest after unit is parser/statements.ts, 1,238,829 bytes; parser/parser.ts is 1,017,423 bytes. Source modules remain the compilation granularity, so a change inside either large module rebuilds that module. The split is still off by default.

Elapsed time wraps native.Build on already emitted C, including splitting, preprocessing, object-cache lookup, compilation and linking. Loading/lowering/emission are outside the timer. Each edit uses a fresh generated-object cache primed only with its baseline. Edited objects are absent. Runtime archives are warm. Versions were interleaved B,A,A,B,B,A, three samples per edit/version, on the same box and unchanged implementation. Counts are actual newly published object entries; the after probe additionally asserts exactly one. Removed units and changed declaration inputs are included in the structural check.

Adding/removing/renaming uses the same unused isolatedProbe function in grammar.ts; rename changes it to renamedProbe. This isolates insertion/removal/identity changes from edits to callers. A rename with real callers must also rebuild the changed caller modules. The other two overlays are the handoff's unused [14].length temporary and return 14 to return [14].length. Behavior gates use the unedited program.

| Loop | Before units / s | After units / s | Before load | After load | Instrument |
|---|---:|---:|---|---|---|
| add-temporary | 1 / 1.515 | 1 / 2.880 | 1.38 to 1.67 | 3.14 to 2.97 | B / A |
| 14-to-array-length | 1 / 1.334 | 1 / 2.601 | 2.97 to 2.97 | 3.00 to 3.08 | B / A |
| add-function | 53 / 12.623 | 1 / 2.587 | 3.04 to 3.18 | 3.04 to 3.04 | B / A |
| remove-function | 53 / 12.999 | 1 / 2.729 | 2.54 to 2.70 | 3.25 to 3.07 | B / A |
| rename-function | 53 / 12.762 | 1 / 2.490 | 2.82 to 2.79 | 3.30 to 3.27 | B / A |

B is the exact before command, using a Go overlay of aca41891's units.go. A omits that overlay. Both run:

```text
ADAMIC_STABLE_LINT_PROBE=1 ADAMIC_STABLE_MEASURE=1 ADAMIC_STABLE_ONE_ROUND=1 ADAMIC_STABLE_READ_INPUTS=1 ADAMIC_STABLE_C_INPUTS=/tmp/stable-c-inputs ADAMIC_STABLE_BASELINE=<B:1,A:0> go test <B:-overlay=/tmp/stable-measure/before.json> ./internal/native -run '^TestStableLintUnitChanges$' -count=1 -v -timeout 30m
```

Build-flags for every row: commit=36e332b885ea4c5780f8551cdb6d52a4d8b7e5a1, effective-before=aca418919f345ecd41889e73832e6a93c95cd3db, nproc=5, cpu.max="400000 100000" (four CPU quota), go="go version go1.27.1 linux/amd64", clang="clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)", node=v24.19.0, jobs=4, cached=unchanged baseline only. Table loads are the one-minute averages before/after each winning sample; all three averages and exact commands/flags for all thirty samples are in stable_splitter_evidence/build-flags.txt and timings.jsonl.

```text
-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all
```

Observed insertion time improves 4.88 times, removal 4.76 times and rename 5.13 times. Small body edits regress 1.90 to 1.95 times despite retaining one-object rebuilds. More units require more preprocessing lookups; this is a plausible contributor, not a separately profiled attribution. The full cold-build duration was not benchmarked here.

Setup: source /workspace/adamic-tools/env.sh; Go ready 0.084s, clang ready 0.410s, Node ready 0.080s, submodules ready 0.147s, markdown dependencies ready 1.950s, Go build ready 54.551s, build cache warm 54.781s, done 54.931s. nproc=5, cpu.max=400000 100000. The setup log is archived with the evidence.

Final validation on the frozen ownership implementation, with complete output in archived logs:

| Command | Result |
|---|---|
| ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m | PASS, entire oracle, Node/JavaScript/native and recorded counts |
| ADAMIC_NATIVE_SPLIT=0 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m | PASS, entire single-file oracle; this path and its emitted C are unchanged |
| ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=1 ADAMIC_GATE_UNCACHED=0 go test ./stage1/cohere/lint -count=1 -timeout 30m | PASS, complete cached split lint suite |
| ADAMIC_GATE_UNCACHED=1 go test ./internal/native -count=1 -timeout 30m | PASS, complete native package including cache/header-provenance checks |
| go vet ./... | PASS |
| python3 internal/native/stable_splitter_evidence/mutants.py | All 11 caught |

The same exact uncached lint command was run with split=1 and split=0, both passing all selected subtests and their strict output comparisons against Go and Node:

```text
ADAMIC_GATE_UNCACHED=1 ADAMIC_NATIVE_SPLIT=<1,0> ADAMIC_NATIVE_JOBS=4 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestJsxLintTrees|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix)$' -count=1 -timeout 30m
```

The full uncached split lint package with one clang worker timed out at 30 minutes while compiling TestMutants/react-no-is-mounted_mutant_from_batch_8. This was a timeout, not a reported output disagreement, and is not counted as a pass. A full single-file lint run was interrupted by the workspace restart. The completed full cached suite and selected uncached comparisons are the claimed coverage. Optional environment-gated corpus/profile/throughput tests retain their usual skips. macOS, WASI splitting, the explicit TSGo split API and the full repository gate were not run.

| Mutant | Check that caught it |
|---|---|
| Restore sixteen-in-order grouping | TestStableUnitInsertion, unrelated unit inputs changed |
| Remove persistent content ownership | TestSharedIdentitiesDoNotMigrate |
| Remove persistent helper ownership in both assignments | TestSharedIdentitiesDoNotMigrate |
| Remove flags from object key | TestUnitCacheFlagsHoldSanitizer, release object omitted UBSan's overflow check |
| Remove declaration relocations | TestUnitDeclarationDisagreement, wrong declaration silently linked |
| Remove cross-module local rejection | TestModuleMainRejectsCrossRangeLocal |
| Remove constant descriptor ownership | TestConstantDescriptorDoesNotMigrate, three units changed |
| Remove source-callee adapter ownership | TestDeclarationAdapterFollowsCallee, nested adapters left their declaration's module |
| Remove coordinator ownership for forwarder globals | TestForwarderGlobalsDoNotMigrate |
| Drop declaration bytes from guard key | TestUnitDeclarationDisagreement, wrong declaration silently linked |
| Remove emitter temporary reset through a Go overlay | TestStableLintUnitChanges, unrelated grammar-edit inputs changed |

The declaration-disagreement fixture changes the caller's return declaration from int to long. Both units compile cleanly; the correct implementation rejects its missing ABI guard at link time. Its relocation/key mutants make the test fail because that wrong pair links successfully. The flags-key mutant fails only the sanitizer observation, with no compiler warning killing it. Mutants use overlays; no emitter production file is changed.
 Reproduce measurements with python3 internal/native/stable_splitter_evidence/measure.py and mutants with python3 internal/native/stable_splitter_evidence/mutants.py after sourcing the setup environment. Both write complete subprocess output to log files. Compressed emitted-C snapshots and SHA-256/size manifests preserve the exact measured inputs.
