Audited u048 at b902a0ccc09e97940571388a1634450da3383559.
13 requested tests remain 13 rows; no shared-input-only family wrappers were found.
Bounded verdicts: 5 sacred, 2 subsumed, 4 overlapping, 1 untrue within its mutant set, 1 witness.
20 production mutants, 2 survivors, 8 separate empty-answer probes; nproc=5.
Sources restored; evidence branch test-audit/internal-native-library.

```json
[
  {
    "test": "TestRuntimeKeyIncludesEveryInput",
    "package": "internal/native",
    "file": "internal/native/library_test.go",
    "seconds": 0.076,
    "oracle": "Handwritten metamorphic key changes and flag order checks.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M05"
    ],
    "unique_kills": [
      "M01",
      "M02",
      "M03"
    ],
    "last_proven_fail": "M05: library_test.go:34: changing adamic.c did not change the key",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRuntimeKeyIncludesEveryInput",
      "TestRuntimeCacheKeepsCountFlags",
      "TestRuntimeCacheRebuildsChangedSources",
      "TestRuntimeCacheConcurrentBuilders",
      "TestRuntimeCacheConcurrentProcesses"
    ],
    "evidence": "ADAMIC_MUTANT=M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestRuntimeKeyIncludesEveryInput|TestRuntimeCacheKeepsCountFlags|TestRuntimeCacheRebuildsChangedSources|TestRuntimeCacheConcurrentBuilders|TestRuntimeCacheConcurrentProcesses)$ > M05.log 2>&1; library_test.go:34: changing adamic.c did not change the key"
  },
  {
    "test": "TestRuntimeCacheKeepsCountFlags",
    "package": "internal/native",
    "file": "internal/native/library_test.go",
    "seconds": 0.187,
    "oracle": "Handwritten empty-program count report, run from the built executable.",
    "oracle_kind": "self",
    "kills": [
      "M06",
      "M08",
      "M09"
    ],
    "unique_kills": [
      "M09"
    ],
    "last_proven_fail": "M09: library_test.go:69: count true: got \"\", want \"adamic: counts: allocations 0 frees 0 retains 0 releases 0 peak 0 regions 0\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P03"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRuntimeKeyIncludesEveryInput",
      "TestRuntimeCacheKeepsCountFlags",
      "TestRuntimeCacheRebuildsChangedSources",
      "TestRuntimeCacheConcurrentBuilders",
      "TestRuntimeCacheConcurrentProcesses"
    ],
    "evidence": "ADAMIC_MUTANT=M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestRuntimeKeyIncludesEveryInput|TestRuntimeCacheKeepsCountFlags|TestRuntimeCacheRebuildsChangedSources|TestRuntimeCacheConcurrentBuilders|TestRuntimeCacheConcurrentProcesses)$ > M09.log 2>&1; library_test.go:69: count true: got \"\", want \"adamic: counts: allocations 0 frees 0 retains 0 releases 0 peak 0 regions 0\\n\""
  },
  {
    "test": "TestRuntimeCacheRebuildsChangedSources",
    "package": "internal/native",
    "file": "internal/native/library_test.go",
    "seconds": 0.199,
    "oracle": "Handwritten C fixture outputs 1, 2, 4 and changed archive identities.",
    "oracle_kind": "self",
    "kills": [
      "M05",
      "M06",
      "M08"
    ],
    "unique_kills": [],
    "last_proven_fail": "M08: library_test.go:117: native: publishing runtime: rename /workspace/u048-tmp/TestRuntimeCacheRebuildsChangedSources2846577434/001/3403f3091932ef2ce58a41fb6cc163b3a9d501860de90dc6b482f023e77562e5 /workspace/u048-tmp/TestRuntimeCacheRebuildsChangedSources2846577434/001/.build-4070347847: no such file or directory",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestRuntimeCacheKeepsCountFlags",
      "TestRuntimeKeyIncludesEveryInput"
    ],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRuntimeKeyIncludesEveryInput",
      "TestRuntimeCacheKeepsCountFlags",
      "TestRuntimeCacheRebuildsChangedSources",
      "TestRuntimeCacheConcurrentBuilders",
      "TestRuntimeCacheConcurrentProcesses"
    ],
    "evidence": "ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestRuntimeKeyIncludesEveryInput|TestRuntimeCacheKeepsCountFlags|TestRuntimeCacheRebuildsChangedSources|TestRuntimeCacheConcurrentBuilders|TestRuntimeCacheConcurrentProcesses)$ > M08.log 2>&1; library_test.go:117: native: publishing runtime: rename /workspace/u048-tmp/TestRuntimeCacheRebuildsChangedSources2846577434/001/3403f3091932ef2ce58a41fb6cc163b3a9d501860de90dc6b482f023e77562e5 /workspace/u048-tmp/TestRuntimeCacheRebuildsChangedSources2846577434/001/.build-4070347847: no such file or directory"
  },
  {
    "test": "TestRuntimeCacheConcurrentBuilders",
    "package": "internal/native",
    "file": "internal/native/library_test.go",
    "seconds": 1.293,
    "oracle": "Handwritten one-compilation, identical archives, output 42 and one cache entry.",
    "oracle_kind": "self",
    "kills": [
      "M07",
      "M08"
    ],
    "unique_kills": [
      "M07"
    ],
    "last_proven_fail": "M08: library_test.go:167: caller 0: native: publishing runtime: rename /workspace/u048-tmp/TestRuntimeCacheConcurrentBuilders1944570777/001/cache/eb4becf7cbff6b19bab0600055fdc60ee922b74dd3b3c55631e62bae318d532d /workspace/u048-tmp/TestRuntimeCacheConcurrentBuilders1944570777/001/cache/.build-3253343070: no such file or directory",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRuntimeKeyIncludesEveryInput",
      "TestRuntimeCacheKeepsCountFlags",
      "TestRuntimeCacheRebuildsChangedSources",
      "TestRuntimeCacheConcurrentBuilders",
      "TestRuntimeCacheConcurrentProcesses"
    ],
    "evidence": "ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestRuntimeKeyIncludesEveryInput|TestRuntimeCacheKeepsCountFlags|TestRuntimeCacheRebuildsChangedSources|TestRuntimeCacheConcurrentBuilders|TestRuntimeCacheConcurrentProcesses)$ > M08.log 2>&1; library_test.go:167: caller 0: native: publishing runtime: rename /workspace/u048-tmp/TestRuntimeCacheConcurrentBuilders1944570777/001/cache/eb4becf7cbff6b19bab0600055fdc60ee922b74dd3b3c55631e62bae318d532d /workspace/u048-tmp/TestRuntimeCacheConcurrentBuilders1944570777/001/cache/.build-3253343070: no such file or directory"
  },
  {
    "test": "TestRuntimeCacheConcurrentProcesses",
    "package": "internal/native",
    "file": "internal/native/library_test.go",
    "seconds": 0.173,
    "oracle": "Handwritten subprocess header, output 42, successful child checks and one cache entry; parent is not merely a helper.",
    "oracle_kind": "self",
    "kills": [
      "M08"
    ],
    "unique_kills": [],
    "last_proven_fail": "M08: library_test.go:232: process 0: exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRuntimeCacheKeepsCountFlags"
    ],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": 0.187,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRuntimeKeyIncludesEveryInput",
      "TestRuntimeCacheKeepsCountFlags",
      "TestRuntimeCacheRebuildsChangedSources",
      "TestRuntimeCacheConcurrentBuilders",
      "TestRuntimeCacheConcurrentProcesses"
    ],
    "evidence": "ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestRuntimeKeyIncludesEveryInput|TestRuntimeCacheKeepsCountFlags|TestRuntimeCacheRebuildsChangedSources|TestRuntimeCacheConcurrentBuilders|TestRuntimeCacheConcurrentProcesses)$ > M08.log 2>&1; library_test.go:232: process 0: exit status 1"
  },
  {
    "test": "TestStringLengthLimit",
    "package": "internal/native",
    "file": "internal/native/limit_test.go",
    "seconds": 0.184,
    "oracle": "V8 13.6.233.17 include/v8-primitive.h:126 64-bit kMaxLength formula checked: (1<<29)-24=536870888. Adamic panic prefix and exit 70 are self-written.",
    "oracle_kind": [
      "external-authority",
      "self"
    ],
    "kills": [
      "M18",
      "M19"
    ],
    "unique_kills": [
      "M18",
      "M19"
    ],
    "last_proven_fail": "M19: limit_test.go:48: 536870889 units: exit 0, \"allowed\\n\"; want exit 70, \"adamic: panic: RangeError: Invalid string length\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 2,
    "probe_kills": [
      "P07"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestStringLengthLimit"
    ],
    "evidence": "ADAMIC_MUTANT=M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestStringLengthLimit)$ > M19.log 2>&1; limit_test.go:48: 536870889 units: exit 0, \"allowed\\n\"; want exit 70, \"adamic: panic: RangeError: Invalid string length\\n\"",
    "vacuous_subcases": [
      "0 units",
      "536870888 units"
    ]
  },
  {
    "test": "TestLoopBorrowPlan",
    "package": "internal/native",
    "file": "internal/native/loop_borrow_test.go",
    "seconds": 0.052,
    "oracle": "Handwritten borrowing, local ownership, lending and function/loop inventory.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M11",
      "M12",
      "M13",
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: loop_borrow_test.go:44: closureReassignBody: borrowed true, want false",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestLoopCallCoverage",
      "TestLoopArrayHoldC"
    ],
    "mutants_in_matrix": 8,
    "probe_kills": [
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLoopBorrowPlan",
      "TestNbodyBorrowedLoopC",
      "TestGlobalArgumentLending",
      "TestLoopArrayHoldC",
      "TestLoopCallCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestLoopBorrowPlan|TestNbodyBorrowedLoopC|TestGlobalArgumentLending|TestLoopArrayHoldC|TestLoopCallCoverage)$ > M14.log 2>&1; loop_borrow_test.go:44: closureReassignBody: borrowed true, want false"
  },
  {
    "test": "TestNbodyBorrowedLoopC",
    "package": "internal/native",
    "file": "internal/native/loop_borrow_test.go",
    "seconds": 0.086,
    "oracle": "Handwritten IR Borrowed flags and absence of named binding retain/release in emitted C; no Node execution.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M12"
    ],
    "unique_kills": [],
    "last_proven_fail": "M12: loop_borrow_test.go:79: function 1 offsetMomentum still counts binding adamic_local_18_body",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestLoopBorrowPlan"
    ],
    "mutants_in_matrix": 8,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": 0.052,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLoopBorrowPlan",
      "TestNbodyBorrowedLoopC",
      "TestGlobalArgumentLending",
      "TestLoopArrayHoldC",
      "TestLoopCallCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestLoopBorrowPlan|TestNbodyBorrowedLoopC|TestGlobalArgumentLending|TestLoopArrayHoldC|TestLoopCallCoverage)$ > M12.log 2>&1; loop_borrow_test.go:79: function 1 offsetMomentum still counts binding adamic_local_18_body"
  },
  {
    "test": "TestGlobalArgumentLending",
    "package": "internal/native",
    "file": "internal/native/loop_borrow_test.go",
    "seconds": 0.056,
    "oracle": "Handwritten eligible-global argument predicate and minimum coverage count.",
    "oracle_kind": "self",
    "kills": [
      "M15"
    ],
    "unique_kills": [
      "M15"
    ],
    "last_proven_fail": "M15: loop_borrow_test.go:110: read virtual 0 argument 0 lent false, want true",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 8,
    "probe_kills": [
      "P06"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLoopBorrowPlan",
      "TestNbodyBorrowedLoopC",
      "TestGlobalArgumentLending",
      "TestLoopArrayHoldC",
      "TestLoopCallCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestLoopBorrowPlan|TestNbodyBorrowedLoopC|TestGlobalArgumentLending|TestLoopArrayHoldC|TestLoopCallCoverage)$ > M15.log 2>&1; loop_borrow_test.go:110: read virtual 0 argument 0 lent false, want true"
  },
  {
    "test": "TestLoopArrayHoldC",
    "package": "internal/native",
    "file": "internal/native/loop_borrow_test.go",
    "seconds": 0.121,
    "oracle": "Handwritten iterator retain/release expectations, emitted definitions and coverage counts.",
    "oracle_kind": "self",
    "kills": [
      "M12",
      "M13",
      "M14",
      "M16",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: loop_borrow_test.go:157: pushBody: redundant iterator retain",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestLoopCallCoverage",
      "TestLoopBorrowPlan"
    ],
    "mutants_in_matrix": 8,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLoopBorrowPlan",
      "TestNbodyBorrowedLoopC",
      "TestGlobalArgumentLending",
      "TestLoopArrayHoldC",
      "TestLoopCallCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestLoopBorrowPlan|TestNbodyBorrowedLoopC|TestGlobalArgumentLending|TestLoopArrayHoldC|TestLoopCallCoverage)$ > M17.log 2>&1; loop_borrow_test.go:157: pushBody: redundant iterator retain"
  },
  {
    "test": "TestLoopCallCoverage",
    "package": "internal/native",
    "file": "internal/native/loop_borrow_test.go",
    "seconds": 0.317,
    "oracle": "Handwritten borrow/lending/iterator expectations, global-call checks and 26-function inventory.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M11",
      "M12",
      "M13",
      "M16",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: loop_borrow_test.go:271: readLocal: redundant iterator retain",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestLoopBorrowPlan",
      "TestLoopArrayHoldC"
    ],
    "mutants_in_matrix": 8,
    "probe_kills": [
      "P04",
      "P05",
      "P06"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLoopBorrowPlan",
      "TestNbodyBorrowedLoopC",
      "TestGlobalArgumentLending",
      "TestLoopArrayHoldC",
      "TestLoopCallCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestLoopBorrowPlan|TestNbodyBorrowedLoopC|TestGlobalArgumentLending|TestLoopArrayHoldC|TestLoopCallCoverage)$ > M17.log 2>&1; loop_borrow_test.go:271: readLocal: redundant iterator retain"
  },
  {
    "test": "TestMapHashProbeBound",
    "package": "internal/native",
    "file": "internal/native/map_hash_test.go",
    "seconds": 12.422,
    "oracle": "C harness asserts SameValueZero, insertion/lookups and probe bound 64; Go checks exit status only. M20 changes hash(1) while this corpus still passes.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P08"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMapHashProbeBound"
    ],
    "evidence": "ADAMIC_MUTANT=P08 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestMapHashProbeBound)$ > P08.log 2>&1; map_hash_test.go:91: probe bound 64 exceeded: integers entries=64 hit=64 miss=65"
  },
  {
    "test": "TestMapHashProbeCatchesMutants",
    "package": "internal/native",
    "file": "internal/native/map_hash_test.go",
    "seconds": 20.233,
    "oracle": "Witness expects three built-in bad hashes to fail with exit 1 and specific diagnostic text, without sanitizer failures.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: map_hash_test.go:131: mutant escaped or failed for another reason: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMapHashProbeCatchesMutants"
    ],
    "evidence": "W01 checker diff applied; timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^TestMapHashProbeCatchesMutants$ > W01.log 2>&1; map_hash_test.go:131: mutant escaped or failed for another reason: <nil>"
  }
]
```

CODE UNDER TEST, named before mutations: native runtime-cache construction in library.go/native.go; borrow planning and C emission in element_borrow.go/borrow.go/emit_statements.go/emit.go; C string-length and map hash functions. ORACLE: handwritten metamorphic, fixture-output, IR-plan and emitted-C assertions; V8's checked kMaxLength value plus self-written panic/exit expectations; the map regression probe's own checker. No row runs Node as an external execution oracle.

Scope: all 13 names exist in list.log at the starting commit; none moved or vanished. All four requested test files were read whole. Each asserts additional properties compared with its siblings, so none was grouped as a family. TestRuntimeCacheConcurrentProcesses has an environment-gated child path but also a real asserting parent path; it is not merely a helper. The map mutant test is a witness and was excluded from production matrices.

Reach: reached-go-functions.txt lists the covered Go production functions with origin locations. test-callers.txt records the broader caller search. Go coverage for the full assigned slice was 41.3 percent. C roots were read directly. The attempted dynamic C inventory failed to start a binary before its outer compilation backstop; c-reach.txt records that limitation. A complete dynamic C function inventory was not obtained.

The clean full package cooked at 90.020 seconds, with no prior assertion failures. The clean assigned slice passed at 21.208 seconds. Matrices were narrowed by main code under test: five cache rows for M01-M09, five borrow/emission rows for M10-M17, StringLengthLimit for M18-M19 and MapHashProbeBound for M20. The exact rows per mode are in matrix.json and each row's matrix_rows. Missing rows and kills outside each group are unknown. Therefore every sacred/unique result is bounded, not proven package-wide uniqueness.

| ID | Starting-commit file:line | Fixed change | Observed failed rows |
|---|---|---|---|
| M01 | internal/native/library.go:129 | part(version) -> drop statement/whole loop | TestRuntimeKeyIncludesEveryInput |
| M02 | internal/native/library.go:128 | part(compiler) -> drop statement/whole loop | TestRuntimeKeyIncludesEveryInput |
| M03 | internal/native/library.go:131 | for _, flag := range flags { 		part(flag) 	} -> drop statement/whole loop | TestRuntimeKeyIncludesEveryInput |
| M04 | internal/native/library.go:124 | fmt.Fprintf(hash, "%d:", len(value)); -> drop statement/whole loop |  |
| M05 | internal/native/library.go:138 | hash.Write(file.contents) -> drop statement/whole loop | TestRuntimeKeyIncludesEveryInput, TestRuntimeCacheRebuildsChangedSources |
| M06 | internal/native/library.go:109 | !strings.HasSuffix(entry.Name(), ".h") -> !strings.HasSuffix(entry.Name(), ".hh") | TestRuntimeCacheRebuildsChangedSources, TestRuntimeCacheKeepsCountFlags |
| M07 | internal/native/library.go:150 | info.Mode().IsRegular() -> !info.Mode().IsRegular() | TestRuntimeCacheConcurrentBuilders |
| M08 | internal/native/library.go:202 | os.Rename(temporary, directory) -> os.Rename(directory, temporary) | TestRuntimeCacheRebuildsChangedSources, TestRuntimeCacheConcurrentProcesses, TestRuntimeCacheConcurrentBuilders, TestRuntimeCacheKeepsCountFlags |
| M09 | internal/native/library.go:216 | "--whole-archive" -> "--no-whole-archive" | TestRuntimeCacheKeepsCountFlags |
| M10 | internal/native/element_borrow.go:47 | program.Locals[loop.Local].Borrowed = true -> program.Locals[loop.Local].Borrowed = false | TestLoopBorrowPlan, TestNbodyBorrowedLoopC, TestLoopCallCoverage |
| M11 | internal/native/element_borrow.go:48 | lending[loop.Iterable.(ir.Read).Local] = true -> lending[loop.Iterable.(ir.Read).Local] = false | TestLoopBorrowPlan, TestLoopCallCoverage |
| M12 | internal/native/element_borrow.go:45 | !changes(program, changing, loop.Body) -> changes(program, changing, loop.Body) | TestLoopBorrowPlan, TestNbodyBorrowedLoopC, TestLoopCallCoverage, TestLoopArrayHoldC |
| M13 | internal/native/element_borrow.go:170 | return false -> return true | TestLoopBorrowPlan, TestLoopCallCoverage, TestLoopArrayHoldC |
| M14 | internal/native/element_borrow.go:96 | !held.Captured && -> drop statement/whole loop | TestLoopBorrowPlan, TestLoopArrayHoldC |
| M15 | internal/native/borrow.go:144 | !pure(argument) -> pure(argument) | TestGlobalArgumentLending |
| M16 | internal/native/emit_statements.go:448 | !e.elementBorrows[e.at] -> e.elementBorrows[e.at] | TestLoopCallCoverage, TestLoopArrayHoldC |
| M17 | internal/native/emit_statements.go:442 | else if e.elementBorrows[e.at] -> else if !e.elementBorrows[e.at] | TestLoopCallCoverage, TestLoopArrayHoldC |
| M18 | internal/native/runtime/string_build_impl.h:31 | units > ADAMIC_STRING_MAX_UNITS -> units >= ADAMIC_STRING_MAX_UNITS | TestStringLengthLimit |
| M19 | internal/native/runtime/string_build_impl.h:31 | units > ADAMIC_STRING_MAX_UNITS -> units > ADAMIC_STRING_MAX_UNITS + 1 | TestStringLengthLimit |
| M20 | internal/native/runtime/map_set.c:29 | hash ^= hash >> 12 -> hash ^= hash >> 11 |  |

Survivors:

M04: flags [a,bc] and [ab,c] yield different production runtimeKey values before the length-prefix statement is dropped, and equal values afterward. Command: ADAMIC_MUTANT=M04 go run review/test-audit/internal-native-library/witness-key-linked.go. Both logs are retained. The linked executable calls the actual unexported production function; an earlier faithful-copy witness is also retained.
M20: the production runtime hash of 1 changes from 316017654 to 316119145, while MapHashProbeBound still passes. One switched C runtime witness was built by witness-hash.go, then executed before and under ADAMIC_MUTANT=M20. Logs and the build timing are retained. This demonstrates changed internal hash output, not a failed map semantic contract.

Witness: W01 is the one permitted checker edit. internal/native/testdata/map_hash.c:140 changes return failures > 0 ? 1 : 0 to return 0. Command: timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^TestMapHashProbeCatchesMutants$ > W01.log 2>&1. All three subcases failed at map_hash_test.go:131 with mutant escaped or failed for another reason: <nil>. The runtime mutants were unchanged. This proves the witness detects a disabled aggregate failure result. W01.diff is separate from production mutants.

Probes: probes.json lists eight actual entry returns; P01 key string empty; P02 cached archive empty; P03 Build returns nil without a product; P04 borrowing maps nil; P05 C source empty; P06 lending answer false; P07 length checker returns; P08 numeric hash zero. They do not contribute to production kills, uniqueness or subsumption. All applicable production rows failed. StringLengthLimit's zero/max-length positive inputs still passed P07, while its excessive/infinite inputs failed; the whole row is not vacuous. The witness was not empty-answer probed and has vacuous null. Build/cache probes can fail at missing-product preconditions; that alone does not establish the strength of later semantic checks.

Brief problems, uncertainties and costs:

- The full internal/native package exceeded the 90-second binary budget, as anticipated for a big package. Scope was narrowed to the assigned slice and then each main code-under-test group. Central replay is needed for omitted rows and package-wide uniqueness.
- I attempted dynamic C coverage by wrapping clang and isolating XDG_CACHE_HOME. That also relocated Go's default build cache, triggered cold Go/CGo compilation, and hit the outer 120-second backstop before a test binary started. No C dynamic inventory was produced. The failed attempt is logged.
- The cold profiling build consumed scratch space. /tmp filled during the isolated map timings, producing ar failures and No space left on device. These were infrastructure failures, not kills. The task-created profiling cache was discarded, Go's warm GOCACHE was pinned, scratch products moved to /workspace, a clean recovered slice passed at 21.812 seconds, and all affected map timings were repeated. disk-full-runs preserves excluded logs.
- My initial C selector omitted stdlib.h for getenv in map_set.c. The standalone production diffs compiled, but the combined selector did not. Those construction failures were excluded in selector-build-failures. I added the scratch include, verified the inactive selector passed at 13.672 seconds, and repeated all affected cache/C/probe modes. Borrow-plan and emitted-C modes did not build that runtime and were unaffected.
- One optional captured-holder witness used numeric console.log arguments, which Adamic's loader refused because console.log requires string. It is retained as a failed exploratory probe and is used for no verdict. M14 itself was caught by two assigned rows.
- The twenty-mutant cap is less than three mutants per thirteen rows. Mutants were fixed from source before catches, not replaced to manufacture uniqueness. The map bound received one production candidate and its empty-answer probe; the witness used its existing three independent built-in mutants.
- MapHashProbeBound's untrue verdict is limited to that single production candidate. P08 makes it fail, and W01 proves the independent built-in-mutant witness can detect checker weakening. No claim that the bound test can never fail is supported.
- Cache construction mutations can be caught by clang/header/publication failures instead of an output comparison. The report names those actual catches, without treating them as proof of unrelated runtime semantics.
- The processes test is a parent with real assertions as well as its own child entry. Calling it helper merely because of its environment variable would discard the parent check.
- No requested row skipped. Timings select one row in each test invocation, but the failed profiling attempt overlapped some first timings and caused contention. Affected map timings were replaced after recovery.
- Full native-runtime rebuild durations inside tests were not separately isolated from test execution. Matrix command/binary durations, standalone C compiler times and the hash-witness build duration are recorded instead.
- No tests in other packages were run. No main push or pull request was made. No general semantic defect is claimed for the surviving altered hash: only the observed internal output change and the corpus's failure to detect it.

Timing: {"finished_utc": "2026-10-09T11:07:09.322289+00:00", "full_baseline_binary_seconds": 90.02, "inactive_selector_binary_seconds": 13.672, "npm_reported": "added 3 packages in 423ms", "probe_compile_seconds": 2.1240399699963746, "recovered_baseline_binary_seconds": 21.812, "setup_seconds": 0, "standalone_compile_seconds": 3.0850414100059425, "timing_command_wall_seconds": 215.94818586199472, "unit_baseline_binary_seconds": 21.208, "valid_matrix_binary_seconds": 112.526, "valid_matrix_command_wall_seconds": 174.17941350498222, "witness_seconds": {"exit": 1, "wall": 23.69543614199938}}

Each standalone production diff was independently applied and compiled: Go via go vet ./internal/native/, C via the runtime's clang flags. compile-status.json records commands and durations for all twenty. Every entry probe was independently compiled in probe-compile-status.json. Source and checker mutations were restored. The valid matrix and probes are separated from excluded infrastructure/construction logs.
