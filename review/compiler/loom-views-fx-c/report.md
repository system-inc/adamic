Built C: snapshot payload ownership, closure-thunk producer certification, unique compatible assignability, and union-call result decoding.
Commit: this single commit on bd90a433; branch compiler/loom-views-fx-c, independent of A and B.
Validation: 51 reader packages checked with focused selectors; reader guard and vet pass; Linux counts regenerated once (149.224s).
Mutants: 11 valid mutants killed; invalid or masked attempts and the green isolation control are listed below.
Not covered: D belongs to its independent branch; 137, 139, 141, 147-150 and 153 remain deferred. No PR or Loom submission.

Baseline observations, before any production edits: 43 distinct programs across C and D were run with Node, generated JavaScript and native on bd90a433. Baseline sources, hashes, commands, stdout, stderr and exit codes are in baseline-manifest.json and main-results.json. The replay scripts retain exact commands. C has 20 probe programs.

| Item | Main differs / probes | C outcome |
|---|---:|---|
| 133 | 1/1 | Fixed through closure thunks |
| 135 | 5/9 | 5 agree; 4 narrowly refused with source path and arrow-wrapper fix |
| 143 | 2/3 | All 3 agree, including original true / undefined |
| 145 | 2/2 | Both agree |
| 154 | 4/5 | All 5 agree, including plain-number control |

Assumptions: assignability alone does not adapt native argument or result slots. Only a producer assignable to exactly one member with the supported scalar ABI is certified. Ambiguous producers, receiver/arity adapters and unrecorded object-result producers get located compile-time refusals and an arrow-wrapper fix. Shared member contracts stay conservative when another containing union is ambiguous. Tagged unread callable descendants remain lazy. No runtime feature set was added; TestRuntimeSetsAreEachUsed passes. No cohere code was copied.

The settled branch CLI replay agrees with Node for every accepted C probe; its four differences are the intended compile-time refusals. The earlier extra replay used 30-second compilation deadlines and hit cold archive-build timeouts under concurrent checks. It was rerun with 90-second deadlines and one worker; source and runtime execution are still bounded.

Exact final commands (all output redirected to the named log):

- `timeout 300 go test $(cat review/compiler/loom-views-fx-c/reader-packages.txt) -run "^(TestCallableProducer|TestCallableUnionResult|TestViewSnapshot|TestLoomViewsC|TestCheckedViewV2CallableProducerMutant$|TestRuntimeSetsAreEachUsed$)" -count=1 -v -timeout 90s` (readers-focused.log): pass. Packages without matching tests were compiled; no whole-package suite or full gate ran.
- `timeout 120 go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -v -timeout 90s` (reader-guard.log): pass, 48.58s.
- `timeout 120 go vet ./internal/lower ./internal/native ./internal/oracle` (vet-final.log): pass, no output.
- `timeout 360 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout 300s -args -update-counts` (counts.log): pass, 149.224s; 15 new rows and 3 existing rows changed.
- `timeout 240 python3 review/compiler/loom-views-fx-c/replay-branch.py` (replay-branch-settled.log): accepted probes agree, 4 located refusals.

Mutant evidence (each script saves and restores sources in finally; sources are .patch, compressed for storage):

| Mutant | Result | Catcher |
|---|---|---|
| packed-snapshot-stale | Killed | `^TestViewSnapshotPayloadOwnership$` |
| native-direct-code | Killed | `^TestCallableProducerRegistryFiltersDirectFunctions$` |
| lower-direct-code | Invalid: unused Go variable; corrected mutant killed | `^TestCallableProducerCertificatesUseClosureThunks$` |
| exact-producer-type | Killed | `^TestCheckedViewV2CallableProducerMutant$` |
| ambiguous-producer | Killed | `^TestCallableProducerAmbiguousRefused$` |
| missing-abi-refusal | Killed | `^TestCallableProducerMethodShorthandRefused$|^TestCallableProducerExtraOptionalRefused$` |
| unrecorded-result | Masked by ambiguity; unique object-result witness killed corrected mutant | `^TestCallableProducerDiscardedObjectResultRefused$` |
| raw-union-result | Killed | `^TestCallableUnionResultNumber$` |
| lower-direct-code-fixed-mutant | Killed | `^TestCallableProducerCertificatesUseClosureThunks$` |
| snapshot-retain-unmasked | Killed | `^TestLoomViewsCOriginal$` |
| tagged-lazy-check | Killed | `^TestCallableProducerTaggedUnreadMethod$` |
| unrecorded-unique-result | Killed | `^TestCallableProducerUniqueObjectResultRefused$` |
| packed-number-stale | Survived: unpacker already zeros absent numbers; redundant production assignment removed | `^TestViewSnapshotPayloadOwnership$` |
| unique-certificate | Masked by incompatible union-parameter ABI; scalar-compatible witness killed corrected mutant | `^TestCallableProducerAmbiguousCertificate$` |
| snapshot-retain-control | Green isolation control: old runtime plus fixed compiler agrees with Node | `^TestLoomViewsCOriginal$` |
| unique-certificate-fixed-witness | Killed | `^TestCallableProducerAmbiguousCertificate$` |

The retain mutant disables the runtime boolean canonicalization too, so it isolates the emitter guard. The matching green control restores only the old runtime and passes the original fixture. The direct-code emitter invariant inspects generated comparisons before clang; no mutant is counted as killed solely by a C warning.

Measured top-level test seconds from final 51-package run (new and touched tests plus runtime-set witness):

| Test | Seconds |
|---|---:|
| TestCallableProducerMethodShorthandAdapted | 0.63 |
| TestCallableProducerExtraOptionalRefused | 0.80 |
| TestCallableProducerUniqueObjectResultRefused | 0.88 |
| TestCallableProducerLiteralReturn | 0.99 |
| TestCallableProducerCertificatesUseClosureThunks | 1.00 |
| TestCallableProducerAmbiguousRefused | 0.23 |
| TestCallableProducerMethodShorthandRefused | 0.69 |
| TestCallableProducerFewerParametersRefused | 0.80 |
| TestCallableProducerDiscardedObjectResultRefused | 0.79 |
| TestCallableProducerFewerParametersAdapted | 0.99 |
| TestCallableProducerP60 | 1.07 |
| TestCallableProducerTaggedUnreadMethod | 1.29 |
| TestCallableProducerP61 | 1.09 |
| TestCallableProducerP21 | 1.02 |
| TestCallableProducerAmbiguousCertificate | 0.92 |
| TestCallableProducerExtraOptionalAdapted | 0.91 |
| TestViewSnapshotPayloadOwnership | 3.00 |
| TestCallableProducerRegistryFiltersDirectFunctions | 0.35 |
| TestRuntimeSetsAreEachUsed | 2.37 |
| TestViewSnapshotMaybeNumber | 5.31 |
| TestLoomViewsCOriginal | 8.15 |
| TestViewSnapshotTypedArrayReference | 8.02 |
| TestCheckedViewV2CallableProducerMutant | 8.90 |
| TestViewSnapshotMaybeBoolean | 2.03 |
| TestCallableUnionResultString | 1.79 |
| TestCallableUnionResultNumber | 1.59 |
| TestCallableUnionResultZero | 2.35 |
| TestCallableUnionResultBoth | 41.45 |

C runtime changes against bd90a433, object.c lines 177-178:

```c
        value.payload = (adamic_value){0};
        if (boolean.present) { value.payload.boolean = boolean.boolean; }
```

Replaces: `value.payload.boolean = boolean.boolean;`
No other runtime C lines change.

Toolchain setup was shared for this unit and sourced from the emitted env file. GOPROXY=https://proxy.golang.org|direct. nproc=5; cgroup quota=4 CPUs; Linux amd64. Full timing output:

```text
submodule cache: off
v24.19.0
go version go1.27.1 linux/amd64
setup: node ready (0.021s)
setup: go ready (0.022s)
setup: submodules ready (0.055s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.010s
setup: markdown dependencies ready (0.070s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.165s)
setup: shared cache ready (0.932s)
setup: go build ready (24.311s)
setup: test binaries deferred (use --warm-tests) (24.389s)
setup: build cache warm (24.391s)
setup: build-flags commit=bd90a433bf636b8325a591ab3e7487d457a69ab4 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 1.20 3.81 1/245 62756 load-after=2.10 1.60 3.88 1/250 63400
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (24.418s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.6YNMjh
```

Automatic approval review initially rejected a combined cleanup/mutant command as an unrestored mutation. Read-only evidence from runtime/maybe.c:41-47 proved the number assignment redundant; a separate, explicitly described final implementation cleanup was approved. No action remains blocked.

The first green push exceeded the requested 30-minute target while baseline replay, mutant isolation, final reader compilation and shared disk-space recovery completed.
