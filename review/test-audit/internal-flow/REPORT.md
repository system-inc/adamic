u018: internal/flow at 7b18d0576930caca4e22ce2eef92fcf563af52d0.
833 top-level tests grouped into 7 rows; clean baseline passed in 59.558 s; no skips.
Nine production mutants caught; all full runs exceeded 90 s, followed by completed bounded reruns.
Verdicts: 1 slow-worthy, 5 subsumed, 1 setup-check; empty ranges pass both range-bearing rows in the bounded probe.
Evidence branch: test-audit/internal-flow; directory: review/test-audit/internal-flow/.

```json
[
  {
    "test": "TestFlowProgram family",
    "package": "internal/flow",
    "file": "internal/flow/corpus_units_test.go; internal/flow/corpus_coverage_test.go",
    "seconds": 62.734,
    "oracle": "Node runs marked Adamic JavaScript and checks graph paths, observed mutations and future reads. SSA uses self-written reaching definitions, VerifySSA and fmt IR-read counts. Mutation ranges left unset are skipped, so absence is accepted. Coverage and remainder included per corpus-family rule.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4",
      "M5",
      "M6",
      "M7",
      "M8",
      "M9"
    ],
    "unique_kills": [
      "M4",
      "M9"
    ],
    "last_proven_fail": "M9: corpus_units_test.go:9103: function 6 (bump): counter$1 was mutated at instruction 0 (order 1), outside its range [1, 2)",
    "verdict": "slow-worthy",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "PBuild",
      "PConstruct",
      "PLiveOut"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestFlowProgram family",
      "TestFlowCorpusSetupIsShared",
      "TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95",
      "TestDebuggerHasNoFlowInstruction"
    ],
    "evidence": "ADAMIC_MUTANT=M9 ADAMIC_BUILD_CACHE_DIR=/tmp/u018/cache/M9 timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run '^(TestFlowCorpusUnitsCoverEveryProgram|TestFlowCorpusSetupIsShared|TestFlowCorpusRemainder|TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95|TestFlowProgram_testdata_joins_a_091c4a59e83f|TestFlowProgram_testdata_mutations_a_fe30ed94ac7e|TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95|TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95|TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95|TestDebuggerHasNoFlowInstruction)$' > M9-bounded.log 2>&1; corpus_units_test.go:9103: function 6 (bump): counter$1 was mutated at instruction 0 (order 1), outside its range [1, 2)",
    "probe_results": {
      "PBuild": "fail",
      "PConstruct": "fail",
      "PLiveOut": "fail",
      "PRanges": "pass"
    },
    "subsumption_mutants": null,
    "vacuous_subcases": [
      "PRanges: every selected corpus member passes; the mutation-range checker skips unset ranges.",
      "PConstruct and PLiveOut: passed members are listed in vacuous-subcases.json; other members failed these entry probes."
    ],
    "vacuous_entries": [
      "InferMutableRanges"
    ],
    "nonvacuous_entries": [
      "Build",
      "Construct",
      "LiveOut"
    ],
    "probe_scope": "bounded corpus inputs; full PRanges run timed out without a completed package result",
    "setup_kills": [
      "SSetupCorpus"
    ],
    "members_file": "rows.json"
  },
  {
    "test": "TestFlowCorpusSetupIsShared",
    "package": "internal/flow",
    "file": "internal/flow/corpus_coverage_test.go",
    "seconds": 0.046,
    "oracle": "Self: repeated preparation must return pointer-identical lowered IR and trace setup. SSetupLower and SSetupTrace independently break these caches.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "SSetupTrace: corpus_coverage_test.go:111: trace setup was rebuilt",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFlowProgram family",
      "TestFlowCorpusSetupIsShared",
      "TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95",
      "TestDebuggerHasNoFlowInstruction"
    ],
    "evidence": "go test -json -count=1 -timeout 90s ./internal/flow/ -run ^TestFlowCorpusSetupIsShared$; SSetupTrace: corpus_coverage_test.go:111: trace setup was rebuilt",
    "probe_results": {},
    "subsumption_mutants": null,
    "setup_kills": [
      "SSetupLower",
      "SSetupTrace"
    ],
    "members_file": "rows.json"
  },
  {
    "test": "TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95",
    "package": "internal/flow",
    "file": "internal/flow/corpus_units_test.go",
    "seconds": 0.059,
    "oracle": "Self: VerifySSA, independent textbook reaching definitions over Build, and fmt IR-read count. No outside expected value checked.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2",
      "M5",
      "M6"
    ],
    "unique_kills": [],
    "last_proven_fail": "M6: corpus_units_test.go:8533: ../oracle/testdata/timsort.a, function -1 (main): shape$1 in instruction 17 names [] (a phi: false), want a phi over what reaches it, [{6 0} {16 0}]",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFlowProgram family"
    ],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "PConstruct"
    ],
    "subsumer_seconds": 62.734,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFlowProgram family",
      "TestFlowCorpusSetupIsShared",
      "TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95",
      "TestDebuggerHasNoFlowInstruction"
    ],
    "evidence": "ADAMIC_MUTANT=M6 ADAMIC_BUILD_CACHE_DIR=/tmp/u018/cache/M6 timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run '^(TestFlowCorpusUnitsCoverEveryProgram|TestFlowCorpusSetupIsShared|TestFlowCorpusRemainder|TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95|TestFlowProgram_testdata_joins_a_091c4a59e83f|TestFlowProgram_testdata_mutations_a_fe30ed94ac7e|TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95|TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95|TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95|TestDebuggerHasNoFlowInstruction)$' > M6-bounded.log 2>&1; corpus_units_test.go:8533: ../oracle/testdata/timsort.a, function -1 (main): shape$1 in instruction 17 names [] (a phi: false), want a phi over what reaches it, [{6 0} {16 0}]",
    "probe_results": {
      "PConstruct": "fail"
    },
    "subsumption_mutants": 4,
    "members_file": "rows.json"
  },
  {
    "test": "TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95",
    "package": "internal/flow",
    "file": "internal/flow/corpus_units_test.go",
    "seconds": 13.205,
    "oracle": "Node observes object changes in Adamic JavaScript; set ranges must contain observed mutations. Unset ranges explicitly skip checks, so an empty table is accepted.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M8"
    ],
    "unique_kills": [],
    "last_proven_fail": "M8: corpus_units_test.go:9109: function 2 (numbers): made$7 was mutated at instruction 3 (order 8), outside its range [1, 2)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFlowProgram family"
    ],
    "mutants_in_matrix": 9,
    "probe_kills": [],
    "subsumer_seconds": 62.734,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestFlowProgram family",
      "TestFlowCorpusSetupIsShared",
      "TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95",
      "TestDebuggerHasNoFlowInstruction"
    ],
    "evidence": "ADAMIC_MUTANT=M8 ADAMIC_BUILD_CACHE_DIR=/tmp/u018/cache/M8 timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run '^(TestFlowCorpusUnitsCoverEveryProgram|TestFlowCorpusSetupIsShared|TestFlowCorpusRemainder|TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95|TestFlowProgram_testdata_joins_a_091c4a59e83f|TestFlowProgram_testdata_mutations_a_fe30ed94ac7e|TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95|TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95|TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95|TestDebuggerHasNoFlowInstruction)$' > M8-bounded.log 2>&1; corpus_units_test.go:9109: function 2 (numbers): made$7 was mutated at instruction 3 (order 8), outside its range [1, 2)",
    "probe_results": {
      "PRanges": "pass"
    },
    "subsumption_mutants": 3,
    "members_file": "rows.json"
  },
  {
    "test": "TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95",
    "package": "internal/flow",
    "file": "internal/flow/corpus_units_test.go",
    "seconds": 13.339,
    "oracle": "Node runs marked Adamic JavaScript; trace must follow Build edges and finish at a return. It requires at least one numeric trace point.",
    "oracle_kind": "external-run",
    "kills": [
      "M2",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: corpus_units_test.go:9115: function -1 (main): instruction 17 ran after 6, and no edge from bb3 leads there",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFlowProgram family"
    ],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "PBuild"
    ],
    "subsumer_seconds": 62.734,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFlowProgram family",
      "TestFlowCorpusSetupIsShared",
      "TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95",
      "TestDebuggerHasNoFlowInstruction"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u018/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run . > M3.log 2>&1; corpus_units_test.go:9115: function -1 (main): instruction 17 ran after 6, and no edge from bb3 leads there",
    "probe_results": {
      "PBuild": "fail"
    },
    "subsumption_mutants": 2,
    "members_file": "rows.json"
  },
  {
    "test": "TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95",
    "package": "internal/flow",
    "file": "internal/flow/corpus_units_test.go",
    "seconds": 14.11,
    "oracle": "Node trace determines later reads; every such variable must be live. Only under-approximation is rejected; extra live variables are accepted.",
    "oracle_kind": "external-run",
    "kills": [
      "M7"
    ],
    "unique_kills": [],
    "last_proven_fail": "M7: corpus_units_test.go:9121: function 2 (numbers): a variable (declaration 12) is read after instruction 9, where liveness says it's dead",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFlowProgram family"
    ],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "PLiveOut"
    ],
    "subsumer_seconds": 62.734,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFlowProgram family",
      "TestFlowCorpusSetupIsShared",
      "TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95",
      "TestDebuggerHasNoFlowInstruction"
    ],
    "evidence": "ADAMIC_MUTANT=M7 ADAMIC_BUILD_CACHE_DIR=/tmp/u018/cache/M7 timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run . > M7.log 2>&1; corpus_units_test.go:9121: function 2 (numbers): a variable (declaration 12) is read after instruction 9, where liveness says it's dead",
    "probe_results": {
      "PLiveOut": "fail"
    },
    "subsumption_mutants": 1,
    "members_file": "rows.json"
  },
  {
    "test": "TestDebuggerHasNoFlowInstruction",
    "package": "internal/flow",
    "file": "internal/flow/debugger_test.go",
    "seconds": 0.009,
    "oracle": "Self: hand-written count of one Evaluate instruction; only the count is checked.",
    "oracle_kind": "self",
    "kills": [
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: debugger_test.go:18: debugger changed the flow: 0 instructions, want 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95"
    ],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "PBuild"
    ],
    "subsumer_seconds": 0.059,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFlowProgram family",
      "TestFlowCorpusSetupIsShared",
      "TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95",
      "TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95",
      "TestDebuggerHasNoFlowInstruction"
    ],
    "evidence": "ADAMIC_MUTANT=M2 ADAMIC_BUILD_CACHE_DIR=/tmp/u018/cache/M2 timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run . > M2.log 2>&1; debugger_test.go:18: debugger changed the flow: 0 instructions, want 1",
    "probe_results": {
      "PBuild": "fail"
    },
    "subsumption_mutants": 1,
    "members_file": "rows.json"
  }
]
```

F = corpus family, S = shared setup, A = timsort SSA, R = timsort mutation ranges, G = timsort graph paths, L = timsort liveness, D = debugger.

| ID | origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M1 | internal/flow/build.go:88 | Flip !declared.Global to declared.Global | F, A, R |
| M2 | internal/flow/build.go:178 | Drop Evaluate emission, return instead | D, F, A, G, R |
| M3 | internal/flow/build.go:258 | Flip statement.CheckAfter | F, G |
| M4 | internal/flow/build.go:450 | Return false at CanThrow entry, remove its body | F |
| M5 | internal/flow/ssa.go:105 | Drop EliminateRedundantPhis(function) | F, A |
| M6 | internal/flow/ssa.go:184 | Flip role != PlaceRoleDefine to == | F, A |
| M7 | internal/flow/liveness.go:44 | Drop the whole instruction.Uses propagation loop | F, L |
| M8 | internal/flow/ranges.go:354 | Drop widenRanges(function, effects, result) | F, R |
| M9 | internal/flow/ranges.go:197 | Change order >= r.Start to order > r.Start | F |

M4 had no failing selected member. Its F kill is observed in the full log, which contains 47 completed corpus failures; every other grouped row passed the bounded run. All uniqueness and subsumption claims are bounded, and subsumption rests only on the listed caught mutants. These are retention questions, not deletion recommendations.

Survivors: none among M1 through M9. M9 also has an independent behavior witness: MutableRange{Start:1, End:3}.Contains(1) prints true clean and false under M9.

PBuild, PConstruct, PLiveOut and PRanges are empty-answer probes, excluded from production kills and uniqueness. PBuild aborted the binary and was rerun independently for all seven rows. PRanges passed the selected corpus members and the full timsort mutation-range row. The combined family reaches four production entries: it rejects empty Build, Construct and LiveOut, but accepts empty InferMutableRanges. Its vacuous=true records that entry-specific finding; it does not say its other checks are empty. Completed positive and unfinished corpus members for each probe are recorded in vacuous-subcases.json.

SSetupLower and SSetupTrace independently make cached preparation return a fresh object, and the shared-setup row fails. SSetupCorpus changes the first registered corpus input from dedication to the already registered hello input, and the corpus coverage member fails. These construction changes are separate from the nine production mutants. The initial expected-name helper edit was rejected as an oracle edit and contributes to no verdict.

The bounded selection contains ten top-level members: two corpus inputs (joins and mutations), coverage, remainder, shared setup, debugger and the four timsort analysis rows. Exact names are in bounded-members.json and exact commands are in commands.json. It is a representative selection, not an exhaustive function-caller set. Full-run completed failures are retained; unfinished rows are not counted as passes. matrix.json includes the full observations and the bounded results.

Brief issues and costs:

- No Your rows section was supplied. I used the entire discovered package. Historical movement or disappearance of separately intended names cannot be judged without those names.
- The corpus-family rule includes coverage and remainder even though they assert different properties. I followed that explicit rule, keeping the separate shared-setup test and the four different timsort checkers as their own rows. Every member is listed in rows.json; test-body-groups.json records the body comparison.
- A combined family calls four entries, so one scalar vacuous field cannot describe all four. I marked it true for the passing range entry and named the rejecting entries separately.
- A 59.558 s green baseline did not predict mutant runtime: every production full run reached 90 s. Nine full runs consumed 850.011 wall seconds, including compilation and shutdown. The bounded reruns and required timing runs substantially exceeded the approximate 20-minute target. Total session time was roughly 50 minutes.
- The liveness mutants emitted multi-gigabyte repeated assertions. Raw logs are preserved as gzip archives; failing-lines.txt contains compact assertion evidence. A test name containing dead was initially mistaken for an error marker during report assembly; assertions were reselected using source-line prefixes and completed failure actions. Verdicts did not change.
- Warm tools lacked the required cohere submodule. Initial discovery compilation exceeded 90 s and was stopped; a warmed retry succeeded. The first scratch switch needed a syntax correction. Standalone diffs were regenerated directly from the pinned source, then every diff was applied, vetted and reverted.
- The debugger oracle checks only instruction count. No instruction-identity claim is made. Node-based flow checks run instrumented Adamic JavaScript, not an independent native backend. SSA expectations and setup expectations are self-written, with no outside-authority value verification.

Costs and limits:

- Tool setup was skipped because /workspace/adamic-tools/env.sh worked: Go 1.27.1, nproc=5. npm ci reported 594 ms. Submodule checkout and compilation-only setup time were not separately instrumented.
- Whole-package clean baseline: 59.558 s. A final clean coverage run passed in 75.545 s with 87.6% statement coverage; 103 production functions had nonzero measured coverage. The complete origin function inventory and the measured list are included.
- Good medians come from three independent -count=1 selections and each package binary own ok line, with mutant guards disabled. Timing runs were serialized after both mutation workers finished. Exact triples are in timings.json. The switch remained installed for timing, adding guard lookup overhead; no speed comparison to untouched source is claimed.
- Recorded matrix and probe wall total: 1139.518 s. Follow-up wall total: 714.715 s, including bounded reruns, timing and construction checks. Some bounded runs overlapped the full probe worker; these totals are not elapsed session time. Compilation-only and native rebuild time were not isolated.
- All nine production diffs, four probe diffs and three valid construction diffs pass go vet ./internal/flow/. Every saved diff, including switch.diff, applies to the restored pinned source. Production and test sources are restored.
- Nine production mutants are fewer than the approximate three per row target. Untested fault models, unselected corpus behavior under unfinished full runs, and repo-wide uniqueness remain unknown. No other package tests were run. Central replay can apply the standalone diffs. No main push and no pull request.
