Built: wave 19 rebased onto current main; six implemented rule algorithms revalidated, no new claims.
Commits: tested source e15f1df46955cac014ea6b1281f0751658527c7c, based on main e8ba3d5d81de4d3773c723914fccd4c76248b965; final evidence commit is the pushed branch tip.
Commands and outputs: six-rule oracle PASS 549.236s, bridge PASS 135.178s, checker PASS 0.252s, filtered Node PASS 4.545s, vet PASS; nproc 5.
Mutants: ten clean-exit byte mutants, five fact-contract mutants and three retained-handle mutants caught again; bridge and Node mutants also caught.
Not covered: three shared production bridge registrations, lint emitted-JavaScript comparison, complete upstream multifile/options matrices and full repository gate; no further rules claimed.

## Landing state and dependency

This is the only branch pushed by this unit: codex/typeaware-wave-19. It was not an ancestor of main. The first clean rebase was onto e011f8f6 and its complete six-rule run passed in 588.850s. Main advanced to e8ba3d5d with call-target and devirtualization changes, so a second clean rebase and a complete fresh validation were required. The final tested base is e8ba3d5d. The previous remote tip is 165964a00293a041e36aab15d3fd26dc45c3f2b0. The requested history update uses an exact lease on that tip, preserving any unexpected remote update rather than overwriting it.

The original 0d540f41 ten-rule bridge coverage commit remains a dependency in this branch because main does not contain it. This rebase retains that dependency; it is not ten newly claimed wave 19 rules. Historical reports retain their original SHAs. evidence/range-diff.log maps the pre-rebase patches to this history. All 33 owned native rule/decoder/runner, Go fact and test source blobs are identical before and after the rebase, as recorded in evidence/source-blobs.json. The sole source-discovery fix renames saved Go oracle evidence to oracle.go.txt: the old .go filename outside testdata caused go vet to treat the copied internal-cohere oracle as an Adamic package. After the rename, go vet ./... passes. The saved oracle text is unchanged.

No protected emitter, lowering, native driver or oracle test file was edited. The current-main compiler changes were inherited through the rebase. Original setup remains Go/clang/Node 0s, submodules 1s, build cache and total 137s; nproc 5, cgroup quota 4 cores. The retained /workspace/adamic-tools/env.sh was sourced for every build/test command.

## Final oracle checks

The original rules are consistent-generic-constructors, dot-notation and no-array-constructor. The continuation rules are nexus/correctness-no-uncleared-race-timeout, nexus/correctness-no-process-exit-after-output and nexus/correctness-require-blocking-standard-streams.

Complete finding, fix and suggestion bytes match independent production Go on 212 original controls (171 findings, 79740 bytes), 49 timeout controls (24 findings, 15809 bytes), 58 process-output controls (32 findings, 19954 bytes) and 43 blocking-stream controls (24 findings, 17421 bytes). Original isolated-declaration controls give 133 findings / 72329 bytes; index-signature controls give 4 / 1773 bytes. All normal/sanitized comparisons pass. For each runner, the frozen compiler corpus has 77 roots, zero findings and 5241 canonical bytes; repository has 287 roots, zero findings and 18485 bytes. Both corpora also match under ASan, UBSan and LeakSanitizer. The known let.if() shared-parser gap remains an explicit exit-70 refusal and is excluded from agreement with its own assertion.

All continuation comparisons use an isolated Go registration overlay. The actual production archive still exits 70 for declaration-ancestry, resolved-callee and program-modules. wave_19_next_registration.patch remains unapplied and git apply --check passes. This is the outstanding integration boundary, not a production-ready registration claim. The raw facts, native decoders, native analyses and owned runners are implemented. Ahra forbids shared-file edits; those files remain untouched. React overlap was already skipped in favor of wave 16, and no new reservation was made under the landing cap.

The full bridge package independently checks 1600 positions over four files, C ABI ownership/lengths, link opt-in and released handles under sanitizers. The filtered compiler oracle checks source Node, sanitized native and emitted JavaScript on 12 fixtures, including new call-target/devirtualization cases, plus its one-byte output mutant. This does not supply an emitted-JavaScript comparison for the lint runners. Full logs are in evidence; 324 source-manifest/output/error streams with SHA-256 hashes are preserved in streams.tar.gz and hashes.json. Reproducible compiled ELF/archive files from the intermediate run were removed to avoid scratch exhaustion; their sources and logs were retained.

## Every mutant rerun

The ten rule/fact byte mutants all compile, exit 0 and produce empty stderr; only independent Go finding bytes kill them. The following log lines also include the three distinct retained-registry mutants, killed by required released-handle refusal checks:

```text
    wave_19_stream_test.go:136: no_process_exit_after_output.a mutant exit 0, empty stderr, only Go bytes catch byte 53
    wave_19_stream_test.go:257: require_blocking_standard_streams.a mutant exit 0, empty stderr, only Go bytes catch byte 54
    wave_19_stream_test.go:258: require_blocking_standard_streams.a mutant exit 0, empty stderr, only Go bytes catch byte 14432
    wave_19_stream_test.go:387: retained-registry mutant exits 0 with empty stderr; all three required-panic checks kill it
    wave_19_test.go:192: generic mutant: exit 0, empty stderr, byte oracle catches byte 54
    wave_19_test.go:192: dot mutant: exit 0, empty stderr, byte oracle catches byte 6490
    wave_19_test.go:192: array mutant: exit 0, empty stderr, byte oracle catches byte 10897
    wave_19_test.go:220: isolated-option checker mutant: exit 0, empty stderr, byte oracle catches byte 54
    wave_19_test.go:220: index-key checker mutant: exit 0, empty stderr, byte oracle catches byte 1335
    wave_19_test.go:220: symbol-origin checker mutant: exit 0, empty stderr, byte oracle catches byte 2308
    wave_19_test.go:278: released-registry mutant: exit 0 caught by required panic 70
    wave_19_timeout_test.go:142: timeout mutant: exit 0, empty stderr, Go bytes catch byte 53
    wave_19_timeout_test.go:212: released-registry mutant: exit 0; required panic check kills it
```

The five raw fact mutants compile and fail their direct checker contract, not a rule verdict comparison:

```text
ancestry-alias: checker contract caught the deliberately wrong fact; compile succeeded
ancestry-name-guard: checker contract caught the deliberately wrong fact; compile succeeded
ancestry-global: checker contract caught the deliberately wrong fact; compile succeeded
callee-return: checker contract caught the deliberately wrong fact; compile succeeded
module-target: checker contract caught the deliberately wrong fact; compile succeeded

```

Bridge foundation mutants: input length +1 and output string length +1 are caught by ASan heap-buffer-overflow; keeping a released handle is caught by the stale-handle assertion; using a source-file type is caught at byte 6; removing link opt-in is caught by refusal; removing C output frees and allocating a region value on the heap are caught by LeakSanitizer. TestTheOracleCatchesOneByte separately proves compiler output comparison can fail. Complete mutation receipts are in evidence/bridge.log and evidence/node.log.

## Native against Go

After all builds and tests finished, three alternating process rounds per runner/corpus were measured with the existing volume_bench.py. The count-only benchmark output matches Go in every round; full-field agreement is established separately above. These are whole-process medians in seconds, including loading and release. All benchmark corpora have zero findings, so no findings-per-second claim is made. Native remains slower than Go.

| Runner | Corpus | Go | Native | Native / Go |
| --- | --- | ---: | ---: | ---: |
| original | compiler | 0.519623 | 2.243701 | 4.318x |
| original | repository | 0.194595 | 0.377188 | 1.938x |
| timeout | compiler | 0.427880 | 2.495565 | 5.832x |
| timeout | repository | 0.222236 | 0.375773 | 1.691x |
| process | compiler | 0.457771 | 2.230824 | 4.873x |
| process | repository | 0.192319 | 0.323866 | 1.684x |
| blocking | compiler | 0.456104 | 2.189306 | 4.800x |
| blocking | repository | 0.225854 | 0.344544 | 1.526x |

Raw timing phases, output, stderr and JSON medians are in evidence/bench.

## Commands

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE19_ARTIFACTS=/tmp/wave19-landing-final/original ADAMIC_WAVE19_TIMEOUT_ARTIFACTS=/tmp/wave19-landing-final/timeout ADAMIC_WAVE19_STREAM_ARTIFACTS=/tmp/wave19-landing-final/process ADAMIC_WAVE19_BLOCKING_ARTIFACTS=/tmp/wave19-landing-final/blocking ADAMIC_WAVE19_RELEASE_ARTIFACTS=/tmp/wave19-landing-final/released ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./stage1/cohere/typeaware -run '^TestWave19(AgreementAndMutants|TimeoutPendingRegistration|ProcessPendingRegistration|BlockingPendingRegistration|StreamReleasedHandles)$' -v -count=1 -timeout 30m > /tmp/wave19-landing-final-agreement.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /tmp/wave19-landing-final-checker.log 2>&1
python3 stage1/cohere/typeaware/testdata/prove_wave_19_next_facts.py /tmp/wave19-landing-final/fact-mutants > /tmp/wave19-landing-final-fact-mutants.log 2>&1
go vet ./... > /tmp/wave19-landing-final-vet.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave19-typescript go test ./bridge/tsgo/... -count=1 -timeout 20m -v > /tmp/wave19-landing-final-bridge.log 2>&1
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw|call_targets_.*|devirtualize)\.a$' -count=1 -timeout 15m -v > /tmp/wave19-landing-final-node.log 2>&1
```

Each of the four benchmark invocations uses its owned native runner and independent Go oracle, both compiler/repository manifests, three rounds, and an output log. No test output was piped. The full Go gate was not run; touched checker/bridge packages, all six owned rule oracles, and the expanded filtered Node oracle are the checks performed.
