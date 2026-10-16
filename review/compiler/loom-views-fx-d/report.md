Built D: checked-cast metadata, callable contract retention and argument adaptation, tuple recognition, recursive membership and diagnostic text.
Commit: this single commit on bd90a433; compiler/loom-views-fx-d, independent of A, B and C.
Validation: focused checks across 51 reader packages, reader guard and vet pass; Linux counts regenerated successfully once (57.155s) after a disk-full attempt.
Mutants: 16 valid mutants killed; two initially masked witnesses were replaced with isolated witnesses.
Not covered: 137, 139, 141, 147-150 and 153 remain deferred. Wrong-type view probes retain checked failures. No PR or Loom submission.

Observations: every requested D item reproduces on bd90a433, before production edits. Baseline source hashes and exact Node, JavaScript and native observations are in baseline-manifest.json and main-results.json. All 23 D probes were replayed on this branch. C baseline observations are also retained in the shared main-results.json; D source copies are under this branch evidence directory.

| Item | Main differs / probes | D outcome |
|---|---:|---|
| 146 | 9/9 | 8 valid probes agree with Node; invalid callable control stops before argument effects |
| 138 | 3/3 | All 3 agree; ordinary-array negative control remains checked |
| 151 | 5/5 | All 5 agree, including 65, 70, 200 and 2000-link witnesses |
| 152 | 3/3 | Checked failures retained; native and JavaScript full diagnostic bytes agree |
| 140 | 3/3 | Checked failures retained; declared write types and tuple failure texts agree |

Conservative choices: checked callable spread arguments remain a located compile-time refusal. Ordinary receiver/rest calls retain their existing path. Native argument slots follow the selected producer, not the synthesized union signature. The initial metadata-only implementation still crashed or printed a wrong value for the broader item-146 callable probes, so contract retention and the historical argument adapter are necessary parts of this fix. No A/B/C dependency was needed. Diagnostic probes are deliberately wrong-type assertions: this unit aligns their checked failure texts rather than admitting unchecked reads. No runtime feature set was added; TestRuntimeSetsAreEachUsed passes. No cohere code was copied.

Exact final commands and outcomes (all output redirected to logs):

- `timeout 300 go test $(cat review/compiler/loom-views-fx-d/reader-packages.txt) -run "^(TestFX7|TestLoomViewsD|TestCallableContract|TestViewCheckedCastMetadata$|TestRuntimeSetsAreEachUsed$)" -count=1 -v -timeout 90s` (readers-focused.log): 50 packages passed or compiled with no matching tests; outer compilation deadline expired before stage3/fixtures completed. All selected test leaves passed.
- `timeout 150 go test $(cat review/compiler/loom-views-fx-d/reader-packages-remaining.txt) -run "^(TestFX7|TestLoomViewsD|TestCallableContract|TestViewCheckedCastMetadata$|TestRuntimeSetsAreEachUsed$)" -count=1 -v -timeout 90s` (readers-remaining.log): remaining reader package checked. No whole-package suite or full gate ran.
- `timeout 120 go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -v -timeout 90s` (reader-guard-recovered.log): pass, 10.02s.
- `timeout 120 go vet ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle` (vet-recovered.log): pass, no output.
- `timeout 360 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout 300s -args -update-counts` (counts-recovered.log): pass, 57.155s, 18 new rows. The first attempt failed with no space left on device and did not update counts.md; only this successful run regenerated it.
- `timeout 180 python3 review/compiler/loom-views-fx-d/replay-branch.py` (replay-branch-final.log): all accepted positive probes agree; checked negative probes have identical backend failure bytes. The script now gives a conservative 90-second per-command deadline and one worker, as the C cold-runtime replay required.

Infrastructure recovery: regenerable Go export caches filled the 32 GB filesystem. The first reader-guard attempt hit its outer deadline, the warm attempt reported no space left on device, and the counts attempt failed for the same reason. Four named regenerable export-cache directories were removed, freeing about 6 GB. Repository sources and review evidence were preserved. Counts, reader guard and vet then passed. No new or touched test leaf exceeded 60 seconds; no over-budget leaf was rerun. The first green push target of 30 minutes was missed.

Mutants (scripts save and restore production sources in finally; mutant sources are compressed .patch files):

| Mutant | Result | Catcher |
|---|---|---|
| cast-metadata | Masked by retained checked calls; isolated metadata witness killed the same mutation | `^TestLoomViewsDOriginal$` |
| clear-callable-contract | P20 output masked the missing check; bad-callee witness killed the same mutation | `^TestLoomViewsDCallableUnion$` |
| source-argument-layout | Killed | `^TestLoomViewsDWideCallable$` |
| native-depth-128 | Killed | `^TestFX7Depth65$` |
| javascript-depth-128 | Killed | `^TestFX7Depth65$` |
| native-write-type | Killed | `^TestFX7TextP32$` |
| javascript-write-type | Killed | `^TestFX7TextP32$` |
| native-object-kind | Killed | `^TestFX7TextScalarTuple$` |
| native-union-kind | Killed | `^TestFX7TextP72$` |
| tuple-unmarked | Killed | `^TestLoomViewsDTupleUnion$` |
| tuple-rejected-object | Killed | `^TestLoomViewsDTupleObject$` |
| ordinary-array-admitted | Killed | `^TestFX7TupleMarkerRejectsOrdinaryArray$` |
| cast-metadata-isolated | Killed | `^TestViewCheckedCastMetadata$` |
| clear-contract-bad-callee | Killed | `^TestLoomViewsDBadCallable$` |
| native-static-kind | Killed | `^TestFX7TextP67$` |
| callable-spread | Killed | `^TestCallableContractSpreadRefused$` |
| producer-mask | Killed | `^TestLoomViewsDWideCallable$` |
| native-scalar-static-kind | Killed | `^TestFX7TextScalarStatic$` |

The metadata invariant deliberately isolates CheckedCast from read/call metadata demands. The ordinary-array mutant was tested with both tuple-brand guards removed and caught by the negative control; source guards were restored before counts and final checks. The runner now preserves the first saved original when multiple mutations target one file.

Measured final top-level test seconds:

| Test | Seconds |
|---|---:|
| TestFX7TupleMarkerRejectsOrdinaryArray | 0.18 |
| TestFX7ViewCycle | 0.23 |
| TestCallableContractPlainSpread | 1.03 |
| TestCallableContractP20 | 1.25 |
| TestCallableContractSpreadRefused | 0.23 |
| TestCallableContractReceiverEvaluatedOnce | 1.36 |
| TestCallableContractP18 | 1.38 |
| TestCallableContractWideProducer | 1.30 |
| TestCallableContractReceiverBinding | 0.89 |
| TestViewCheckedCastMetadata | 0.00 |
| TestFX7ViewCycleAndStackGuard | 7.44 |
| TestFX7TextScalarStatic | 10.68 |
| TestFX7Depth65 | 11.10 |
| TestFX7TextP33 | 11.21 |
| TestLoomViewsDWideCallable | 12.50 |
| TestLoomViewsDBadCallable | 12.70 |
| TestLoomViewsDTupleObject | 2.22 |
| TestLoomViewsDTupleUnion | 2.37 |
| TestFX7TextP09 | 1.66 |
| TestFX7TextP32 | 1.48 |
| TestFX7TextP72 | 1.65 |
| TestLoomViewsDOriginal | 3.09 |
| TestFX7TextScalarTuple | 1.64 |
| TestRuntimeSetsAreEachUsed | 6.70 |
| TestFX7TextP67 | 2.14 |
| TestLoomViewsDCallableUnion | 3.31 |
| TestFX7Depth200 | 2.98 |
| TestFX7Depth2000 | 3.88 |

Runtime C changes against bd90a433. Every added/replacement line follows; there are no other runtime C changes.

`internal/native/runtime/object.c`, lines 233-237:

```c
	if ((actual == adamic_rep_object || actual == adamic_rep_union || actual == adamic_rep_weak) && slot->reference != NULL && ((const adamic_heap *)slot->reference)->kind == adamic_kind_object) {
		const adamic_object *value = slot->reference;
		if (value->tuple) { found = "array"; }
		else if (value->class != NULL && value->class->is_static) { found = "function"; }
	}
```

`internal/native/runtime/view_unions_mixed.c`, lines 69-73:

```c
    if (value->kind == adamic_view_union_object && value->payload.reference != NULL) {
        const adamic_object *object = value->payload.reference;
        if (object->tuple) { found = "array"; }
        else if (object->class != NULL && object->class->is_static) { found = "function"; }
    }
```

`internal/native/runtime/view_unions_untagged.c`, lines 94-94:

```c
    ADAMIC_CHECK_STACK();
```

The untagged line replaces `if (depth > 128) { return false; }`. Object and mixed-union lines are insertions only. Runtime owners must clear these three files before submission.

Toolchain setup was shared for this unit. GOPROXY=https://proxy.golang.org|direct, nproc=5, cgroup quota=4 CPUs, Linux amd64. The emitted env file was sourced. Full timing output:

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

Automatic approval review rejected moving a local baseline copy as potentially invalidating C evidence. The safer requested alternative, copying into D while preserving the original, succeeded. No action remains blocked.
