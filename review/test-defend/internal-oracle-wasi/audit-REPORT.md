Unit u072: all five requested names exist at base 09fe4b54913753188a9357982bfd47cdf36ef97c.
Bounded matrix: nine production mutants, one empty-entry probe, one witness weakening.
Agreement is slow-worthy within the bounded matrix; emission and weak reads are subsumed hints.
Both named witness rows fail when their comparison is disabled.
Evidence is under review/test-audit/internal-oracle-wasi; production sources are restored.

[
  {
    "test": "TestWASIAgreesWithNode",
    "package": "internal/oracle",
    "file": "internal/oracle/wasi_test.go",
    "seconds": null,
    "oracle": "Node stdout, stderr and exit code; checked fixtures use Adamic JavaScript backend instead. The bounded four fixtures use Node. Full isolated row exceeded 90s.",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M05",
      "M06",
      "M09"
    ],
    "unique_kills": [
      "M09"
    ],
    "last_proven_fail": "M09: wasi_test.go:75: TestWASIAgreesWithNode/shard-001/dedication/dedication.a: exit codes differ",
    "verdict": "slow-worthy",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASIAgreesWithNode",
      "TestWASIEmission",
      "TestWeakReadsUndefinedOnceFreed"
    ],
    "evidence": "ADAMIC_MUTANT=M09 ADAMIC_BUILD_CACHE_DIR=/workspace/u072-products/M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run \"$BOUNDED\" > M09-bounded.log 2>&1; wasi_test.go:75: TestWASIAgreesWithNode/shard-001/dedication/dedication.a: exit codes differ",
    "seconds_lower_bound": 90,
    "timing_runs": 1,
    "over_budget": true
  },
  {
    "test": "TestWASIOracleCatchesMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/wasi_test.go",
    "seconds": 0.363,
    "oracle": "Executed Node observation plus self-written disagreement labels stdout differs and exit codes differ. W01 disables comparison and both witnesses fail.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: wasi_test.go:105: byte mutant: got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestWASIOracleCatchesMutants",
      "TestWASIRunnerCatchesMutants"
    ],
    "evidence": "go test -overlay /workspace/u072-tmp/witness.json -json -count=1 -timeout 90s ./internal/oracle/ -run \"^(TestWASIOracleCatchesMutants|TestWASIRunnerCatchesMutants)$\" > W01.log 2>&1; wasi_test.go:105: byte mutant: got \"\""
  },
  {
    "test": "TestWASIRunnerCatchesMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/wasi_test.go",
    "seconds": 0.41,
    "oracle": "Executed Node observation plus self-written disagreement labels stdout differs and exit codes differ. W01 disables comparison and both witnesses fail.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: wasi_test.go:136: byte mutant: \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestWASIOracleCatchesMutants",
      "TestWASIRunnerCatchesMutants"
    ],
    "evidence": "go test -overlay /workspace/u072-tmp/witness.json -json -count=1 -timeout 90s ./internal/oracle/ -run \"^(TestWASIOracleCatchesMutants|TestWASIRunnerCatchesMutants)$\" > W01.log 2>&1; wasi_test.go:136: byte mutant: \"\""
  },
  {
    "test": "TestWASIEmission",
    "package": "internal/oracle",
    "file": "internal/oracle/wasi_test.go",
    "seconds": 33.286,
    "oracle": "WASI SDK clang 20.1.8 compilation succeeds under native.Flags, including -Werror and -pedantic. Checks compilation, not output; M06 triggered -Wliteral-range. Expected success is a self-written criterion.",
    "oracle_kind": "self",
    "kills": [
      "M06"
    ],
    "unique_kills": [],
    "last_proven_fail": "M06: wasi_test.go:174: emitted C: exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestWeakReadsUndefinedOnceFreed"
    ],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.133,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASIAgreesWithNode",
      "TestWASIEmission",
      "TestWeakReadsUndefinedOnceFreed"
    ],
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/workspace/u072-products/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run \"^TestWASIEmission$\" > M06-emission.log 2>&1; wasi_test.go:174: emitted C: exit status 1",
    "subsumption_mutants": 1
  },
  {
    "test": "TestWeakReadsUndefinedOnceFreed",
    "package": "internal/oracle",
    "file": "internal/oracle/weak_test.go",
    "seconds": 0.133,
    "oracle": "Handwritten native and Node/JavaScript outcomes in weak_test.go. Native lifetime semantics follow repository docs/memory.md, which is not an outside authority. Node runs, but does not supply the native expected answer.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M05",
      "M06"
    ],
    "unique_kills": [],
    "last_proven_fail": "M06: weak_test.go:52: native: stdout differs: exit 0, stdout \"while held: bbbb\\nafter: gone true\\n\", stderr \"\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestWASIAgreesWithNode"
    ],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASIAgreesWithNode",
      "TestWASIEmission",
      "TestWeakReadsUndefinedOnceFreed"
    ],
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/workspace/u072-products/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run \"$BOUNDED\" > M06-bounded.log 2>&1; weak_test.go:52: native: stdout differs: exit 0, stdout \"while held: bbbb\\nafter: gone true\\n\", stderr \"\"",
    "subsumer_seconds_lower_bound": 90,
    "subsumption_mutants": 3
  }
]

Mutants, origin/main locations, changes and observed failed rows:

| id | origin file:line | change | failed rows |
|---|---|---|---|
| M01 | internal/native/native.go:29 | index < len(value) -> index < len(value)-1 | TestWASIAgreesWithNode, TestWeakReadsUndefinedOnceFreed |
| M02 | internal/native/emit_values.go:73 | return "NAN" -> return "0.0" |  |
| M03 | internal/native/emit_values.go:75 | return "HUGE_VAL" -> return "0.0" |  |
| M04 | internal/native/emit_values.go:77 | return "(-HUGE_VAL)" -> return "0.0" |  |
| M05 | internal/native/emit_expressions.go:254 | if expression.Present { -> if !expression.Present { | TestWASIAgreesWithNode, TestWeakReadsUndefinedOnceFreed |
| M06 | internal/native/emit_values.go:79 | strconv.FormatFloat(value, 'x', -1, 64) -> strconv.FormatFloat(value, 'x', 0, 64) | TestWASIAgreesWithNode, TestWASIEmission, TestWeakReadsUndefinedOnceFreed |
| M07 | internal/native/native.go:86 | flags = append(flags, "-ffp-contract=off") -> flags = append(flags, "-ffp-contract=fast") |  |
| M08 | internal/native/native.go:90 | flags = append(flags, "-fno-optimize-sibling-calls") -> flags = append(flags, "-foptimize-sibling-calls") |  |
| M09 | internal/native/target.go:59 | return append(flags, "-mexec-model=command") -> return append(flags, "-mexec-model=reactor") | TestWASIAgreesWithNode |
| P01 | internal/native/emit.go:22 | C returns empty string at entry | TestWASIAgreesWithNode, TestWASIEmission, TestWeakReadsUndefinedOnceFreed |
| W01 | internal/oracle/oracle_test.go:717 | disagreement returns empty string | TestWASIOracleCatchesMutants, TestWASIRunnerCatchesMutants |

Survivors, all bounded; omitted fixtures and other rows are unknown:
M02: cNumber(NaN): NAN -> 0.0; actual production-function witnesses in M02-witness.log and control-witness.log.
M03: cNumber(+Inf): HUGE_VAL -> 0.0; actual production-function witnesses in M03-witness.log and control-witness.log.
M04: cNumber(-Inf): (-HUGE_VAL) -> 0.0; actual production-function witnesses in M04-witness.log and control-witness.log.
M07: Flags: -ffp-contract=off -> -ffp-contract=fast; actual production-function witnesses in M07-witness.log and control-witness.log.
M08: Flags: -fno-optimize-sibling-calls -> -foptimize-sibling-calls; actual production-function witnesses in M08-witness.log and control-witness.log.

Code under test: native C emission (C/cProgram and all measured emitter callees), native flags and WASI link flags. The measured 164 Go functions are listed in reached-functions.txt, with origin locations and percentages; reach.out is the pre-mutant coverage profile. This is the bounded fixture inventory, not an exhaustive transitive inventory for full emission or C runtime execution. No Node runner, Go matcher, fixture or expected result was changed by production mutants.
Oracle: Node for ordinary WASI agreement, clang compilation for emission, handwritten weak expectations, and the actual comparison weakened only for the two witnesses.
The fixed nine-mutant menu is plan.json, written before mutant execution. It spans cString, cNumber, evaluate, Flags and WASILinkFlags. Nine rather than approximately fifteen mutations kept the full-emission matrix within budget. Each standalone diff is against the starting base, has no selector, applies to its index and passes go vet ./internal/native/ via an overlay of original files plus that one mutation. W01 separately passes go vet ./internal/oracle/. selector.diff plus audit_selector.go.txt preserve scratch instrumentation.
The actual matrix ran all four weak fixtures and four WASI agreement fixtures: dedication/dedication.a, internal/oracle/testdata/numbers.a, weak_narrowed.a, reuse.a. It also ran the entire emission row separately for every mutant and probe. Exact commands, timings and rows are in matrix.json. $BOUNDED in the abbreviated result commands is the exact slash-aware regex in selection.py.
P01 is not counted as a production kill or unique kill. All three production rows failed it. The two witnesses are not assigned a production verdict and were not empty-probed.
The two witness bodies differ in construction and control assertions, not only checker input, so they remain separate rows. No requested test moved or vanished.

Brief ambiguities and costs:
1. Whole package and the combined large-row baseline both cooked at 90 seconds. No assertion failure was observed before timeout. The assigned bounded baseline and complete emission row passed. Package uniqueness and kills outside the bounded rows remain unknown.
2. TestWASIAgreesWithNode alone cooked at 90.088 seconds. Its seconds is null, not an invented three-run median. The explicit instruction to stop cooked runs took precedence over repeating an already over-budget row twice. Other four rows have three successful independent timing runs.
3. The first slash-aware bounded selection accidentally ran WASI shard setup without fixture leaves. It was excluded as WASI evidence, corrected before any mutant, and verified by actual run events. The corrected subset also selected reuse.a through its shared filename component, so the saved scope names all four fixtures. Full-baseline and initial timing execution overlapped briefly; this can affect measured times. Isolated successful timings were then sequential.
4. Warm env.sh worked, but lacked WASI_SYSROOT. cloud/setup.sh --wasi-sdk was necessary to enable the opt-in rows; it also refreshed the Go build cache, taking 46.193 seconds. npm ci added three packages in 528ms. No audit row remained skipped; emission has 16 non-lowering fixture skips listed in skipped-subcases.json.
5. clang is absent from the briefs enumerated external-run tools. Emission is labeled self because success is a handwritten criterion rather than a copied diagnostic or differential answer. Its external compiler is named explicitly.
6. The weak row executes Node but its native expectation is explicitly different. Calling that a Node oracle for native lifetime behavior would be incorrect: native stdout/stderr/exit values are handwritten and based on the repositories own memory model. No outside authority was claimed or checked.
7. Flags survivors demonstrate changed actual function outputs and command options, not a demonstrated program miscompile. NaN/Inf survivors demonstrate changed emitted constants. None was silently called equivalent or repo-wide unguarded.
8. Standalone Go compilation time is separate in compile-status.json. Native/WASM rebuilding occurs inside test-binary time; individual product rebuilds were not separately instrumented. The environment selector and per-mutant ADAMIC_BUILD_CACHE_DIR distinguish observations. Existing native archives are keyed by actual compiler flags; unchanged archives may be reused.
9. Coverage listed every dynamically reached native Go function for the bounded clean run before mutations. A complete transitive inventory for the enormous full emission corpus and C runtime was not produced. No C runtime mutation was used. Other packages were not tested.
10. Subsumption rests on three shared mutations for weak reads and one for emission. It is a bounded hint, not a deletion recommendation. Agreement has no successful full-row median, so its subsumer_seconds is null with a 90-second lower bound.

Timing:
{
  "setup_seconds": 46.193,
  "sdk_ready_seconds": 3.6,
  "npm_reported_seconds": 0.528,
  "nproc": 5,
  "whole_package_baseline_seconds": 90.037,
  "large_rows_baseline_seconds": 90.077,
  "valid_matrix_wall_seconds": 420.42782397399424,
  "valid_matrix_binary_seconds": 370.71,
  "timing_wall_seconds": 226.38531761099875,
  "standalone_compile_seconds": 4.603518194991921,
  "witness_wall_seconds": 15.29965439499938
}
