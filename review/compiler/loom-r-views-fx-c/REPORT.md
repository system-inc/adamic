Rebased callable producers, view snapshots and callable union results for #aecx10a 143, 133, 135, 145 and 154.
Commit: the single commit on compiler/loom-r-views-fx-c; exact base dbdf977f5b8b0979d414ebe6f20e7aadd312e309.
Validation: focused lower/native/oracle checks, flow corpus, reader guard, vet, reader compilation and Linux counts pass.
Mutants: see mutant-results.json for the eleven corrected mutants and green isolation control rerun here.
Not covered: no full package suite, full gate, ahra, PR or Loom submission; standalone build-ignored metrics tool not run.

## Apply and resolution

Fetched e323da0f29a0e59ecfe14107f9f90916887271e9 and applied only that commit with
`git cherry-pick --no-commit e323da0f29a0e59ecfe14107f9f90916887271e9` onto the pinned
base. No diff from main to the old tip was used.

Only internal/oracle/counts.md conflicted. Its single conflict combines main's five
instantiation rows with the incoming undefined-read-write retain reduction. The exact
three-way hunk is in counts-conflicts.txt. Resolution: restore the whole pinned-main
table, then regenerate once on Linux clang 20.1.8 after compilation. All five main
instantiation rows survive; none of the old counts hunks was hand-merged.

An expanded native-agreement run exposed an integration error already present in the
supplied change: it registers p22 and p135_object_result with lowers=false, which the
shared oracle interprets as NotYet, although their ambiguous producers correctly return
Refused. The two cases failed the generic runner at oracle_test.go:704. In this change's
own callable_producers_test.go only, remove these two entries from the NotYet loop and
explain ownership in a comment. Dedicated lower tests retain their Node control and
located ambiguity/arrow-wrapper checks. The other three NotYet controls remain registered.
No shared harness or another package's tests changed. No refused top-level oracle
fixture was introduced; all four new top-level fixtures pass flow lowering.

Conservative assumption retained from the source: only unique assignable callable
producers whose native scalar ABI agrees are certified. Receiver, arity and unrecorded
result adapters remain located refusals/NotYet with the arrow-wrapper fix.

## Runtime clearance

No supplied runtime line changed. The automatically merged object.c preserves current
main's later storage cases and adds exactly these original lines (now at 217-218):

```c
        value.payload = (adamic_value){0};
        if (boolean.present) { value.payload.boolean = boolean.boolean; }
```

They replace `value.payload.boolean = boolean.boolean;`, exactly as in e323da0f's own
diff. runtime-lines.json records the byte comparison. No other runtime file changes.
No runtime feature set was added; TestRuntimeSetsAreEachUsed passes.

## Validation commands and results

Every test command redirected output to the corresponding log here; no test output was
piped. Shells source /workspace/adamic-tools/env.sh. Four-CPU cgroup; nproc=5.

- `timeout 180 bash cloud/setup.sh`: first attempt failed because setup overlapped the
  apply and saw method references before the newly added implementation files arrived.
  Exact diagnostics in setup.log. Rerun after the apply passed: setup-final.log.
- `timeout 120 go test ./internal/lower ./internal/native ./internal/oracle -run
  '^(TestCallableProducer|TestCallableUnionResult|TestViewSnapshot|TestLoomViewsC|TestCheckedViewV2CallableProducerMutant$|TestRuntimeSetsAreEachUsed$|TestNativeAgreementRowsNameRegisteredPrograms$)'
  -count=1 -v -timeout 90s -parallel 4`: outer exit 124 after lower passed (3.266s)
  and native passed (33.285s), before any oracle test started. No test leaf timed out.
- Oracle portion separately, same focused selector without the lower registration test,
  outer 120s / Go 90s: PASS 20.311s (oracle.log). Runtime set witness included.
- `timeout 90 go test ./internal/oracle -run
  '^TestNativeAgreesWithNode_[0-9]+$/^internal$/^lower$/^testdata$/^callable_producers$/'
  -timeout 90s -count=1 -parallel 4 -v`: first run caught the two registration bugs;
  after the own-registry correction, PASS 5.068s (producer-fixtures-final.log).
  This adds real native/sanitizer/Node agreement because main's lower agreement helper
  now covers JavaScript only. All eight admitted producer fixtures and three NotYet
  controls pass. The two Refused cases are held by the lower tests.
- `timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1
  -timeout 90s`: PASS 49.789s (reader-guard.log).
- `timeout 90 go vet ./internal/lower ./internal/native ./internal/oracle`: PASS, both
  initial and after registry correction (vet.log, vet-final.log).
- `timeout 120 go test ./internal/flow -run
  '^TestFlowCorpusUnitsCoverEveryProgram$|^TestFlowCorpusRemainder$/.*callable_union_result'
  -count=1 -v -timeout 90s -parallel 4`: PASS coverage, but slash selection did not
  execute the four fixture subtests. Corrected selector below actually ran each.
- `timeout 90 go test ./internal/flow -run
  '^TestFlowCorpusRemainder$/^..$/^oracle$/^testdata$/^callable_union_result_'
  -count=1 -v -timeout 90s -parallel 4`: PASS 1.254s; all four analyses for each new
  top-level fixture (flow-fixtures.log).
- `timeout 300 go test $(cat review/compiler/loom-views-fx-c/reader-packages.txt)
  -run '^$' -count=1 -timeout 90s -p 2`: PASS all 51 supplied reader packages, compilation
  only (readers.log). A current source scan additionally found internal/leakcheck (PASS),
  stage3/drivers/parser/build-metrics (requires stage3_split_metrics, PASS with that tag),
  and stage3/drivers/scanner (only a //go:build ignore standalone tool, excluded).
  The initial untagged supplemental invocation reports the two build exclusions;
  additional-readers.log and tagged-reader.log distinguish them from passing packages.
- `timeout 360 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v
  -timeout 300s -parallel 4 -args -update-counts`: PASS 165.266s. Exactly one regeneration.
  Later registry correction removes only unmeasured false entries and changes no count row.

All new/touched top-level test measurements are in test-seconds.json, reproduced below.
No test leaf exceeded 60s and none was rerun after exceeding that limit.

| Test | Seconds |
|---|---:|
| TestNativeAgreementRowsNameRegisteredPrograms | 0.15 |
| TestCallableProducerDiscardedObjectResultRefused | 0.79 |
| TestCallableProducerExtraOptionalAdapted | 0.84 |
| TestCallableProducerFewerParametersAdapted | 0.95 |
| TestCallableProducerTaggedUnreadMethod | 1.06 |
| TestCallableProducerMethodShorthandAdapted | 0.69 |
| TestCallableProducerP21 | 0.69 |
| TestCallableProducerCertificatesUseClosureThunks | 0.74 |
| TestCallableProducerP61 | 0.72 |
| TestCallableProducerP60 | 0.72 |
| TestCallableProducerUniqueObjectResultRefused | 0.60 |
| TestCallableProducerMethodShorthandRefused | 0.72 |
| TestCallableProducerAmbiguousRefused | 0.25 |
| TestCallableProducerAmbiguousCertificate | 0.92 |
| TestCallableProducerExtraOptionalRefused | 0.59 |
| TestCallableProducerFewerParametersRefused | 0.73 |
| TestCallableProducerLiteralReturn | 0.71 |
| TestViewSnapshotPayloadOwnership | 33.27 |
| TestCallableProducerRegistryFiltersDirectFunctions | 0.20 |
| TestViewSnapshotMaybeBoolean | 7.19 |
| TestViewSnapshotTypedArrayReference | 7.42 |
| TestCheckedViewV2CallableProducerMutant | 7.67 |
| TestCallableUnionResultZero | 1.68 |
| TestCallableUnionResultString | 1.85 |
| TestCallableUnionResultNumber | 1.82 |
| TestViewSnapshotMaybeNumber | 1.79 |
| TestLoomViewsCOriginal | 2.06 |
| TestRuntimeSetsAreEachUsed | 4.32 |
| TestCallableUnionResultBoth | 20.06 |

## Counts delta

15 new rows and 3 changed existing rows. Main's unrelated rows are preserved.
counts-rows.json contains both before and after values. Complete resulting changed rows:

| Fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/lower/testdata/callable_producers/p21.a | 3 | 3 | 7 | 13 | 3 | 0 |
| internal/lower/testdata/callable_producers/p60.a | 3 | 3 | 7 | 13 | 3 | 0 |
| internal/lower/testdata/callable_producers/p61.a | 4 | 4 | 7 | 14 | 4 | 0 |
| internal/lower/testdata/callable_producers/p23.a | 3 | 3 | 7 | 13 | 3 | 0 |
| internal/lower/testdata/callable_producers/p22_adapted.a | 3 | 3 | 6 | 11 | 3 | 0 |
| internal/lower/testdata/callable_producers/p59_adapted.a | 4 | 4 | 7 | 11 | 4 | 0 |
| internal/lower/testdata/callable_producers/p62_adapted.a | 4 | 4 | 7 | 11 | 4 | 0 |
| internal/lower/testdata/callable_producers/p135_tagged_unread.a | 4 | 4 | 7 | 11 | 4 | 0 |
| internal/oracle/testdata/callable_union_result_number.a | 4 | 4 | 1 | 6 | 4 | 0 |
| internal/oracle/testdata/callable_union_result_string.a | 4 | 4 | 2 | 7 | 3 | 0 |
| internal/oracle/testdata/callable_union_result_zero.a | 4 | 4 | 1 | 6 | 4 | 0 |
| internal/oracle/testdata/callable_union_result_both.a | 17 | 17 | 5 | 26 | 7 | 0 |
| internal/oracle/testdata/loom_views_c/snapshot-original.a | 3 | 3 | 8 | 15 | 3 | 0 |
| stage3/interface-downcasts/v2/snapshot-maybe-boolean.a | 6 | 6 | 9 | 24 | 2 | 0 |
| stage3/interface-downcasts/v2/snapshot-maybe-number.a | 10 | 10 | 9 | 21 | 4 | 0 |
| stage3/interface-downcasts/lane4b/fixtures/node-indicator-good.a | 5 | 5 | 14 | 22 | 4 | 0 |
| stage3/interface-downcasts/v2/callable-producer-wrong.a | 3 | 3 | 6 | 11 | 3 | 0 |
| stage3/interface-downcasts/v2/undefined-read-write.a | 2 | 2 | 4 | 9 | 2 | 0 |

## Mutants

run-mutants.py reconstructs the original report's valid corrected mutations, stores each
source mutation as .patch, uses 90s Go / 115s subprocess limits, and restores originals
in finally. Run with `timeout 660 python3 review/compiler/loom-r-views-fx-c/run-mutants.py`.
mutant-plan.json holds exact changes and selectors; mutant-results.json holds commands,
exit codes and seconds; each named .log holds the actual catcher.

Historical invalid unused-variable and masked attempts are not rerun as purported proofs.
The source's corrected lower-direct-code, unique-result and unique-certificate witnesses
replace them. packed-number-stale has no surviving production mutation site: the original
report removed the redundant assignment after that mutant survived. The matched old-runtime
snapshot-retain-control is expected green, not falsely counted as a killed mutant.

## Setup timing output

```
submodule cache: off
go version go1.27.1 linux/amd64
setup: go ready (0.061s)
v24.19.0
setup: node ready (0.077s)
setup: submodules ready (0.153s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.016s
setup: markdown dependencies ready (0.193s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.303s)
setup: shared cache ready (2.985s)
setup: go build ready (72.753s)
setup: test binaries deferred (use --warm-tests) (73.042s)
setup: build cache warm (73.043s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=2.42 0.63 0.22 12/204 1700 load-after=11.24 3.38 1.19 7/298 3944
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (73.074s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.dMu1SD
```

No cohere implementation code was copied. Original source review evidence is retained
under review/compiler/loom-views-fx-c; this rebase's new evidence is entirely here.

## Rerun mutant results

| Mutant | Result | Catcher | Command seconds |
|---|---|---|---:|
| packed-snapshot-stale | Killed | `^TestViewSnapshotPayloadOwnership$` | 19.128 |
| native-direct-code | Killed | `^TestCallableProducerRegistryFiltersDirectFunctions$` | 8.581 |
| exact-producer-type | Killed | `^TestCheckedViewV2CallableProducerMutant$` | 15.667 |
| ambiguous-producer | Killed | `^TestCallableProducerAmbiguousRefused$` | 9.087 |
| missing-abi-refusal | Killed | `^TestCallableProducerMethodShorthandRefused$|^TestCallableProducerExtraOptionalRefused$` | 8.853 |
| raw-union-result | Killed | `^TestCallableUnionResultNumber$` | 11.312 |
| lower-direct-code-fixed-mutant | Killed | `^TestCallableProducerCertificatesUseClosureThunks$` | 8.930 |
| snapshot-retain-unmasked | Killed | `^TestLoomViewsCOriginal$` | 24.106 |
| tagged-lazy-check | Killed | `^TestCallableProducerTaggedUnreadMethod$` | 8.944 |
| unrecorded-unique-result | Killed | `^TestCallableProducerUniqueObjectResultRefused$` | 3.031 |
| unique-certificate-fixed-witness | Killed | `^TestCallableProducerAmbiguousCertificate$` | 9.066 |
| snapshot-retain-control | Green control passed | `^TestLoomViewsCOriginal$` | 11.309 |

All eleven real mutants failed their intended tests; the green isolation control passed.
No mutant failed compilation or timed out. Every production mutation was restored.
The source change is unchanged except the own-registry correction; runtime additions
are byte-identical and the counts table is regenerated. gofmt reports no changed Go file.

Stored mutant patches use zero context and normalize added blank lines in evidence only, to avoid whitespace warnings on quoted Go tabs.
