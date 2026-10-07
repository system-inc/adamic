Built: rebased all wave-19 commits onto current origin/main and rebuilt all six owned native rule oracles; no new rules claimed.
Commits: previous pushed tip 2eaa56ca355b1da696e8d2aae4f23e7a9f389968; rebased source tip a83d29a7c932e56aa3f714770cf83c8982596a11; base f8013f0baac41ddc340d76f83bddde38536a8f07; this evidence commit is pushed only to codex/typeaware-wave-19.
Commands and outputs: six-rule oracle PASS 566.705s; full bridge PASS 159.659s; checker PASS 0.276s; filtered Node PASS 22.460s; vet PASS. Setup 195s, nproc 5.
Mutants: ten byte-only, five raw-fact, three retained-registry, one listener-metadata, seven foundation bridge and the Node one-byte checks caught their mutations.
Uncovered: three production registrations, handed-node numeric execution, lint emitted-JavaScript comparison and full repository gate; no new claims.

The rebase completed without conflicts. range-diff.txt records all 20 commits as patch-equivalent. The source tree diff for stage1/cohere/typeaware and bridge/tsgo between the previous pushed tip and the rebased source tip is empty. The existing dependency commit porting ten inventory rules is retained because main still lacks those bridge modules. No main or area branch was pushed. Only this branch is ours in this session.

The six rule comparisons ran against unchanged production Go cohere on 77 TypeScript compiler roots and the frozen 287-root repository corpus, retaining declaration roots. Findings, fixes and suggestions match as complete canonical bytes. Original controls: 212 sources, 171 findings; isolated controls 133 findings; index controls 4 findings. Timeout: 49 controls, 24 findings. Process output: 58 controls, 32 findings. Blocking streams: 43 controls, 24 findings. Normal and ASAN/UBSAN/LSAN builds pass. The known let.if() parser refusal remains explicitly tested and excluded from agreement, as documented in the original report.

Every previous mutant was rerun. Byte receipts:

```text
    wave_19_stream_test.go:136: no_process_exit_after_output.a mutant exit 0, empty stderr, only Go bytes catch byte 58
    wave_19_stream_test.go:257: require_blocking_standard_streams.a mutant exit 0, empty stderr, only Go bytes catch byte 59
    wave_19_stream_test.go:258: require_blocking_standard_streams.a mutant exit 0, empty stderr, only Go bytes catch byte 14602
    wave_19_stream_test.go:387: retained-registry mutant exits 0 with empty stderr; all three required-panic checks kill it
    wave_19_test.go:192: generic mutant: exit 0, empty stderr, byte oracle catches byte 59
    wave_19_test.go:192: dot mutant: exit 0, empty stderr, byte oracle catches byte 6540
    wave_19_test.go:192: array mutant: exit 0, empty stderr, byte oracle catches byte 10957
    wave_19_test.go:220: isolated-option checker mutant: exit 0, empty stderr, byte oracle catches byte 59
    wave_19_test.go:220: index-key checker mutant: exit 0, empty stderr, byte oracle catches byte 1340
    wave_19_test.go:220: symbol-origin checker mutant: exit 0, empty stderr, byte oracle catches byte 2338
    wave_19_test.go:278: released-registry mutant: exit 0 caught by required panic 70
    wave_19_timeout_test.go:142: timeout mutant: exit 0, empty stderr, Go bytes catch byte 58
    wave_19_timeout_test.go:212: released-registry mutant: exit 0; required panic check kills it
```

Raw-fact mutations remove ancestry alias resolution, loosen the binding-name guard, hide global augmentation, zero resolved-callee return flags, and drop module targets. Each compiles and fails its direct checker contract. The listener source mutant changes process CallExpression 214 to NewExpression 215 and is caught by the independent enum check, with got [215], want [214]; it is not a findings mutant because today's driver ignores the declaration. Foundation bridge mutations corrupt input/output lengths (ASan), retain a released handle (stale-handle assertion), query the source-file position (independent byte oracle), remove link opt-in (refusal), omit output frees (LSan), and allocate a region result on the heap (LSan). The Node one-byte check also passes. Full logs are preserved.

The production archive still refuses declaration-ancestry, wave19-resolved-callee and wave19-program-modules with panic 70. Their prepared implementations and decoders are held by scratch registration overlays. The three-line wave_19_next_registration.patch remains unapplied under Ahra's shared-file restriction.

Numeric listener declarations remain intact in all six classes. Main's shared ParseNode still exposes kind as a string, and the owned rule run API receives an entire file, not a handed node. Existing rules therefore still scan nodes and read string kinds. The pending shared parser/driver is needed for the execution part of the speed requirement. New rule.json/kinds and per-node ports have not been started. The batch-8 Diagnostic landing SHA was not provided in this message. No shared parser, generator, harness or protected compiler file was edited.

Commands, with all output redirected to logs:

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/wave19-f801-scratch
ADAMIC_WAVE19_ARTIFACTS=$TMPDIR/original ADAMIC_WAVE19_TIMEOUT_ARTIFACTS=$TMPDIR/timeout ADAMIC_WAVE19_STREAM_ARTIFACTS=$TMPDIR/process ADAMIC_WAVE19_BLOCKING_ARTIFACTS=$TMPDIR/blocking ADAMIC_WAVE19_RELEASE_ARTIFACTS=$TMPDIR/released ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./stage1/cohere/typeaware -run '^TestWave19(NumericListenerDeclarations|AgreementAndMutants|TimeoutPendingRegistration|ProcessPendingRegistration|BlockingPendingRegistration|StreamReleasedHandles)$' -v -count=1 -timeout 30m
python3 stage1/cohere/typeaware/testdata/prove_wave_19_next_facts.py $TMPDIR/fact-mutants
ADAMIC_TSGO_CORPUS=/workspace/wave19-typescript go test ./bridge/tsgo -count=1 -timeout 15m -v
go test ./bridge/tsgo/checker -count=1 -v
go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
ADAMIC_WAVE19_LISTENER_SOURCE=/tmp/wave19-listeners/metadata-mutant go test ./stage1/cohere/typeaware -run '^TestWave19NumericListenerDeclarations$' -count=1 -v
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw|call_targets_.*|devirtualize)\.a$' -count=1 -timeout 15m -v
```

Setup readiness: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 195s, total 195s; 5 reported processors, cgroup quota 4, memory 17.6 GB. Filtered Node tests may use the supported worker cache and report native hits 12/misses 25, Node hits 0/misses 25. The complete repository gate and baseline 26-rule corpus suite were not rerun; the six owned rules and full bridge gate were.

Fresh three-round alternating count-only whole-process medians, after all builds/tests finished:

| Runner | Corpus | Go seconds | Native seconds | Native / Go |
| --- | --- | --- | --- | --- |
| original | compiler | 0.461923 | 2.250732 | 4.87 |
| original | repository | 0.191242 | 0.374788 | 1.96 |
| timeout | compiler | 0.502589 | 2.534513 | 5.04 |
| timeout | repository | 0.219312 | 0.379422 | 1.73 |
| process | compiler | 0.445903 | 1.945638 | 4.36 |
| process | repository | 0.201163 | 0.327440 | 1.63 |
| blocking | compiler | 0.459522 | 2.227425 | 4.85 |
| blocking | repository | 0.202327 | 0.359399 | 1.78 |

Native remains slower, 1.63 to 5.04 times Go here. These are observations of complete processes, not isolated checker crossing costs or dispatch improvements. The archive contains 456 stdout/stderr/manifest/config/contract/benchmark files with SHA-256 hashes; paths are original-run provenance. Older evidence remains historical, with its own base and timing. This unit handles the landing-first requirement and stops at the existing shared integration blockers instead of adding claims.
