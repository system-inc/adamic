Unit u106: 6503 requested functions verified against fresh origin/main.
Grouped into 11 rows, with mixed agreement and witness roles preserved.
Baseline whole package cooked; enabled bounded baseline passed.
Uniqueness and subsumption are bounded, pending central replay.
Evidence includes standalone diffs, raw logs, inventories and timing samples.

```json
[
  {
    "test": "TestDecodedOptionsAndMutant family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/bv61ffb_decoded_options_test.go:49",
    "seconds": 5.084,
    "oracle": "Go cohere executed live; canonical diagnostic, suggestion, rejection, convergence and fixed-source comparison; member 003 checks only Go count, members 004/005 witness built-in mutants, Union checks registration",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M4"
    ],
    "unique_kills": [
      "M1",
      "M2"
    ],
    "last_proven_fail": "M4 bv61ffb_decoded_options_test.go:54: ignored decoded-option mutant caught on mutant emitted JavaScript: case 0 line 2: port \"/tmp/adamic-gate/lint-shared-588547747/catch.ts:1:26\", Go \"fixed\\ttry { work(); } catch(e) {}\\\\u000a\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDecodedOptionsAndMutant_000",
      "TestDecodedOptionsAndMutant_001",
      "TestDecodedOptionsAndMutant_002",
      "TestDecodedOptionsAndMutant_003",
      "TestDecodedOptionsAndMutant_004",
      "TestDecodedOptionsAndMutant_005",
      "TestDecodedOptionsAndMutantUnion",
      "TestClosedComparatorGaps",
      "TestCompilerAndStage1Agree_002",
      "TestCompilerAndStage1Agree_007",
      "TestCompilerAndStage1Agree_012"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestDecodedOptionsAndMutant_[0-9]+|TestDecodedOptionsAndMutantUnion|TestClosedComparatorGaps|TestCompilerAndStage1Agree_(002|007|012))$ > M4.log; bv61ffb_decoded_options_test.go:54: ignored decoded-option mutant caught on mutant emitted JavaScript: case 0 line 2: port \"/tmp/adamic-gate/lint-shared-588547747/catch.ts:1:26\", Go \"fixed\\ttry { work(); } catch(e) {}\\\\u000a\"",
    "members": [
      "TestDecodedOptionsAndMutant_000",
      "TestDecodedOptionsAndMutant_001",
      "TestDecodedOptionsAndMutant_002",
      "TestDecodedOptionsAndMutant_003",
      "TestDecodedOptionsAndMutant_004",
      "TestDecodedOptionsAndMutant_005",
      "TestDecodedOptionsAndMutantUnion"
    ],
    "timing_samples": [
      4.918,
      5.084,
      5.619
    ],
    "timing_censored_runs": 0,
    "seconds_lower_bound": null,
    "vacuous_subcases": [
      "TestDecodedOptionsAndMutantUnion",
      "TestDecodedOptionsAndMutant_003",
      "TestDecodedOptionsAndMutant_004",
      "TestDecodedOptionsAndMutant_005"
    ]
  },
  {
    "test": "TestDecodedOptionsAndMutantPlantedFailure",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/bv61ffb_decoded_options_test.go:141",
    "seconds": 0.027,
    "oracle": "Suite-authored construction or planted-failure expectations",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2 bv61ffb_decoded_options_test.go:172: planted disagreement caught by shards []",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestDecodedOptionsAndMutantPlantedFailure"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestDecodedOptionsAndMutantPlantedFailure$ > W2.log; bv61ffb_decoded_options_test.go:172: planted disagreement caught by shards []",
    "members": [
      "TestDecodedOptionsAndMutantPlantedFailure"
    ],
    "timing_samples": [
      0.027,
      0.025,
      0.062
    ],
    "timing_censored_runs": 0,
    "seconds_lower_bound": null
  },
  {
    "test": "TestDecodedOptionsAndMutant_Setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/bv61ffb_decoded_options_test.go:449",
    "seconds": 4.895,
    "oracle": "Suite-authored construction or planted-failure expectations",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S4 bv61ffb_decoded_options_test.go:236: build lint-decoded-options-lowered-setup-v2: open /tmp/u106/cache/S4/.building-3bf5b24adfe0-2387858918/missing/lint.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestDecodedOptionsAndMutant_Setup"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestDecodedOptionsAndMutant_Setup$ > S4.log; bv61ffb_decoded_options_test.go:236: build lint-decoded-options-lowered-setup-v2: open /tmp/u106/cache/S4/.building-3bf5b24adfe0-2387858918/missing/lint.c: no such file or directory",
    "members": [
      "TestDecodedOptionsAndMutant_Setup"
    ],
    "timing_samples": [
      5.114,
      4.895,
      4.631
    ],
    "timing_censored_runs": 0,
    "seconds_lower_bound": null
  },
  {
    "test": "TestClosedComparatorGaps",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/closed_gaps_test.go:17",
    "seconds": 0.805,
    "oracle": "Node executed live plus handwritten expected outputs 2, 1, 1; checked by clean baseline",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4 None",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDecodedOptionsAndMutant family"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 5.084,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDecodedOptionsAndMutant_000",
      "TestDecodedOptionsAndMutant_001",
      "TestDecodedOptionsAndMutant_002",
      "TestDecodedOptionsAndMutant_003",
      "TestDecodedOptionsAndMutant_004",
      "TestDecodedOptionsAndMutant_005",
      "TestDecodedOptionsAndMutantUnion",
      "TestClosedComparatorGaps",
      "TestCompilerAndStage1Agree_002",
      "TestCompilerAndStage1Agree_007",
      "TestCompilerAndStage1Agree_012"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestDecodedOptionsAndMutant_[0-9]+|TestDecodedOptionsAndMutantUnion|TestClosedComparatorGaps|TestCompilerAndStage1Agree_(002|007|012))$ > M4.log; None",
    "members": [
      "TestClosedComparatorGaps"
    ],
    "timing_samples": [
      1.542,
      0.657,
      0.805
    ],
    "timing_censored_runs": 0,
    "seconds_lower_bound": null
  },
  {
    "test": "TestCommandDiagnosticsDropOnlyModuleDownloads",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/command_test.go:12",
    "seconds": 0.026,
    "oracle": "Suite-authored construction or planted-failure expectations",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1 command_test.go:27: one download: got \"go: downloading github.com/dlclark/regexp2 v1.11.5\\n\", want \"\"",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCommandDiagnosticsDropOnlyModuleDownloads"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestCommandDiagnosticsDropOnlyModuleDownloads$ > S1.log; command_test.go:27: one download: got \"go: downloading github.com/dlclark/regexp2 v1.11.5\\n\", want \"\"",
    "members": [
      "TestCommandDiagnosticsDropOnlyModuleDownloads"
    ],
    "timing_samples": [
      0.026,
      0.024,
      0.051
    ],
    "timing_censored_runs": 0,
    "seconds_lower_bound": null
  },
  {
    "test": "TestExecuteFailsOnStderrOtherThanModuleDownloads",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/command_test.go:35",
    "seconds": 0.171,
    "oracle": "Suite-authored construction or planted-failure expectations",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1 command_test.go:61: a download and a warning: execute passed, or failed without naming the stderr: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestExecuteFailsOnStderrOtherThanModuleDownloads"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestExecuteFailsOnStderrOtherThanModuleDownloads$ > W1.log; command_test.go:61: a download and a warning: execute passed, or failed without naming the stderr: <nil>",
    "members": [
      "TestExecuteFailsOnStderrOtherThanModuleDownloads"
    ],
    "timing_samples": [
      0.193,
      0.171,
      0.128
    ],
    "timing_censored_runs": 0,
    "seconds_lower_bound": null
  },
  {
    "test": "TestCompilerAndStage1Agree_Setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/compiler_stage1_split_test.go:231",
    "seconds": 0.195,
    "oracle": "Suite-authored construction or planted-failure expectations",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S2 compiler_stage1_split_test.go:259: enumerated 6485, want 7782",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCompilerAndStage1Agree_Setup"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestCompilerAndStage1Agree_Setup$ > S2.log; compiler_stage1_split_test.go:259: enumerated 6485, want 7782",
    "members": [
      "TestCompilerAndStage1Agree_Setup"
    ],
    "timing_samples": [
      0.201,
      0.195,
      0.181
    ],
    "timing_censored_runs": 0,
    "seconds_lower_bound": null
  },
  {
    "test": "TestProduct_CompilerAgreement family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/compiler_stage1_split_test.go:285",
    "seconds": 22.409,
    "oracle": "Suite-authored construction or planted-failure expectations",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S5 compiler_stage1_split_test.go:287: product compiler-agreement-go-oracle failed preparation",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_CompilerAgreementGoOracle",
      "TestProduct_CompilerAgreementSanitizedNative",
      "TestProduct_CompilerAgreementReleaseNative"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestProduct_CompilerAgreementGoOracle$ > S5.log; compiler_stage1_split_test.go:287: product compiler-agreement-go-oracle failed preparation",
    "members": [
      "TestProduct_CompilerAgreementGoOracle",
      "TestProduct_CompilerAgreementSanitizedNative",
      "TestProduct_CompilerAgreementReleaseNative"
    ],
    "timing_samples": [
      22.409,
      0.857
    ],
    "timing_censored_runs": 1,
    "seconds_lower_bound": null
  },
  {
    "test": "TestCompilerAndStage1Agree family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/compiler_stage1_split_test.go:397",
    "seconds": null,
    "oracle": "Go cohere executed live; canonical diagnostic, suggestion, rejection, convergence and fixed-source comparison",
    "oracle_kind": "external-run",
    "kills": [
      "M3"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M3 compiler_stage1_split_test.go:401: case 3 line 8: port \"fixed\\timport {\\\\u000d\\\\u000a    __String,\\\\u000d\\\\u000a    AccessorDeclaration,\\\\u000d\\\\u000a    addEmitHelpers,\\\\u000d\\\\u000a    addRange,\\\\u000d\\\\u000a    addSyntheticLeadingComment,\\\\u000d\\\\u000a    AllAccessorDeclarations,\\\\u000d\\\\u000a    append,\\\\u000d\\\\u000a    arrayIsEqualTo,\\\\u000d\\\\ [line continues in log]",
    "verdict": "slow-worthy",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDecodedOptionsAndMutant_000",
      "TestDecodedOptionsAndMutant_001",
      "TestDecodedOptionsAndMutant_002",
      "TestDecodedOptionsAndMutant_003",
      "TestDecodedOptionsAndMutant_004",
      "TestDecodedOptionsAndMutant_005",
      "TestDecodedOptionsAndMutantUnion",
      "TestClosedComparatorGaps",
      "TestCompilerAndStage1Agree_002",
      "TestCompilerAndStage1Agree_007",
      "TestCompilerAndStage1Agree_012"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestCompilerAndStage1Agree_(002|007|012)$ > M3.log; compiler_stage1_split_test.go:401: case 3 line 8: port \"fixed\\timport {\\\\u000d\\\\u000a    __String,\\\\u000d\\\\u000a    AccessorDeclaration,\\\\u000d\\\\u000a    addEmitHelpers,\\\\u000d\\\\u000a    addRange,\\\\u000d\\\\u000a    addSyntheticLeadingComment,\\\\u000d\\\\u000a    AllAccessorDeclarations,\\\\u000d\\\\u000a    append,\\\\u000d\\\\u000a    arrayIsEqualTo,\\\\u000d\\\\ [line continues in log]",
    "members": {
      "file": "groups.json",
      "group": 8,
      "count": 6485
    },
    "timing_samples": [],
    "timing_censored_runs": 3,
    "seconds_lower_bound": 90
  },
  {
    "test": "TestCompilerAndStage1AgreePlantedDisagreement",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/compiler_stage1_split_test.go:13369",
    "seconds": 0.112,
    "oracle": "Suite-authored construction or planted-failure expectations",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W3 compiler_stage1_split_test.go:13382: planted disagreement not caught exactly once: === RUN   TestCompilerAndStage1Agree_6471",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCompilerAndStage1AgreePlantedDisagreement"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestCompilerAndStage1AgreePlantedDisagreement$ > W3.log; compiler_stage1_split_test.go:13382: planted disagreement not caught exactly once: === RUN   TestCompilerAndStage1Agree_6471",
    "members": [
      "TestCompilerAndStage1AgreePlantedDisagreement"
    ],
    "timing_samples": [
      0.112,
      0.105,
      0.112
    ],
    "timing_censored_runs": 0,
    "seconds_lower_bound": null
  },
  {
    "test": "TestCompleteSuggestionSerialization_IsolatedShard",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/complete_suggestion_isolated_shard_test.go:12",
    "seconds": 32.553,
    "oracle": "Suite-authored construction or planted-failure expectations",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S3 complete_suggestion_isolated_shard_test.go:26: isolated shard: exit status 1",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCompleteSuggestionSerialization_IsolatedShard"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestCompleteSuggestionSerialization_IsolatedShard$ > S3.log; complete_suggestion_isolated_shard_test.go:26: isolated shard: exit status 1",
    "members": [
      "TestCompleteSuggestionSerialization_IsolatedShard"
    ],
    "timing_samples": [
      45.675,
      32.553,
      31.663
    ],
    "timing_censored_runs": 0,
    "seconds_lower_bound": null
  }
]
```

| ID | Origin file:line | Change | Observed failing top-level members |
|---|---|---|---|
| M1 (production) | stage1/cohere/lint/settings.ts:29 | `toLowerCase()` to `toUpperCase()` | TestDecodedOptionsAndMutant_000, TestDecodedOptionsAndMutant_001, TestDecodedOptionsAndMutant_002 |
| M2 (production) | stage1/cohere/lint/settings.ts:70 | `return values[0] ?? fallback;` to `return values[1] ?? fallback;` | TestDecodedOptionsAndMutant_000, TestDecodedOptionsAndMutant_001, TestDecodedOptionsAndMutant_002 |
| M3 (production) | stage1/cohere/lint/main.ts:33 | `    linter.run();` to `` | TestCompilerAndStage1Agree_002, TestCompilerAndStage1Agree_007, TestDecodedOptionsAndMutant_004, TestDecodedOptionsAndMutant_005 |
| M4 (production) | internal/native/emit_branches.go:86 | `condition, whenTrue, whenNot)` to `condition, whenNot, whenTrue)` | TestClosedComparatorGaps, TestDecodedOptionsAndMutant_002 |
| P1 (probe) | stage1/cohere/lint/main.ts:11 | `function run(row: string, countOnly: boolean): number {` to `function run(row: string, countOnly: boolean): number {
    if (row.length >= 0) return 0;` | TestCompilerAndStage1Agree_002, TestCompilerAndStage1Agree_007, TestCompilerAndStage1Agree_012, TestDecodedOptionsAndMutant_000, TestDecodedOptionsAndMutant_001, TestDecodedOptionsAndMutant_002 |
| P2 (probe) | internal/native/emit.go:23 | `return cProgram(program, -1)` to `return ""` | TestClosedComparatorGaps |
| W1 (witness) | stage1/cohere/lint/lint_test.go:87 | `len(commandDiagnostics(name, stderr.Bytes())) != 0` to `len(commandDiagnostics(name, stderr.Bytes())) < 0` | TestExecuteFailsOnStderrOtherThanModuleDownloads |
| W2 (witness) | stage1/cohere/lint/bv61ffb_decoded_options_test.go:125 | `func decodedOptionsCheck(index int, got, want []byte) error {` to `func decodedOptionsCheck(index int, got, want []byte) error {
	if len(want) >= 0 { return nil }` | TestDecodedOptionsAndMutantPlantedFailure |
| W3 (witness) | stage1/cohere/lint/compiler_stage1_split_test.go:13364 | `if diff := difference(got, want); diff != "" {` to `if diff := difference(got, want); diff == "u106-never" {` | TestCompilerAndStage1AgreePlantedDisagreement |
| S1 (setup) | stage1/cohere/lint/lint_test.go:65 | `if name == "go" {` to `if name != "go" {` | TestCommandDiagnosticsDropOnlyModuleDownloads |
| S2 (setup) | stage1/cohere/lint/compiler_stage1_split_test.go:34 | `const compilerAgreementSides = 5` to `const compilerAgreementSides = 6` | TestCompilerAndStage1Agree_Setup |
| S5 (setup) | stage1/cohere/lint/compiler_stage1_split_test.go:183 | `if prepared.path == "" {` to `if prepared.path != "" {` | TestProduct_CompilerAgreementGoOracle |
| S4 (setup) | stage1/cohere/lint/bv61ffb_decoded_options_test.go:373 | `filepath.Join(output, "lint.c")` to `filepath.Join(output, "missing/lint.c")` | TestDecodedOptionsAndMutant_Setup |
| S3 (setup) | stage1/cohere/lint/complete_suggestion_isolated_shard_test.go:20 | `ADAMIC_COMPLETE_SUGGESTION_PLANT=0` to `ADAMIC_COMPLETE_SUGGESTION_PLANT=1` | TestCompleteSuggestionSerialization_IsolatedShard |

Survivors:

Brief issues and limits:
The brief says 13 rows but enumerates 6503 functions. Family grouping produces 11 rows. All requested names exist. The stated historical commit differs from current origin/main; base.txt pins the actual start.
The decoded family combines positive port comparisons, a Go-only count check, built-in-mutant witnesses and a registration union. Production failures in the built-in-mutant members are preserved as raw evidence but excluded from semantic kills.
The compiler family includes Go-only shards and empty buckets. A completed Go-only or empty shard cannot establish port correctness. Full family timing cooks at 90 seconds; only selected source-Node shards enter the production matrix. Other backend and package kills are unknown.
Product declaration tests use one shared recipe checker and are construction checks, rather than semantic agreement rows. The stderr tests also test the suite harness itself; their edits are separately marked construction or witness experiments.
The function inventories are static conservative supersets; exact dynamic reachability was not measured. Four native rebuild mutations were chosen instead of 20 variants. No claim covers unexecuted functions.
Warm tools did not include the pinned TypeScript corpus. It was installed at /tmp/u106/typescript from v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8, to enable compiler agreement rows. npm ci ran before baseline.
Cold product construction can exceed the 90-second budget even when warm checks pass. Censored timing samples are not successful medians. Native rebuild durations are preserved in each build log; process wall times are in results.jsonl.
Initial compiler failures for M1/M2/M3/P1 were corpus cleanliness refusals, not semantic kills. The corpusfiles loader compares selected tracked TypeScript inputs to HEAD. Clean temporary variant commits allowed real comparisons without weakening that guard; replays replaced the refused outcomes. Central replay must likewise commit applied port diffs or use a clean variant checkout. Scratch commit ids and initial logs are retained in results.jsonl.
The initial P1 unconditional early return made later TypeScript discriminated-union refinements unreachable and failed compilation. That attempt is excluded. The corrected probe uses if (row.length >= 0) return 0 at entry. String length is always nonnegative, but the guard preserves type checking of the remaining source and its built-in-mutant anchor. Its actual native rebuild and agreement failures are logged.
P2 returns empty C source. The compiler-gap row rejects it during clang compilation, before semantic output comparison. This proves an empty artifact is rejected by the row pipeline, not that its semantic assertions run on an empty native executable.
Source Node compiler shard 012 passed M3, while 002 and 007 caught it. Those are observations, not evidence that 012 is permanently vacuous. P1 made all three selected source-Node shards fail. Shards 003, 004, 005 and the decoded Union passed P1: 003 is Go-only, 004/005 accept any inequality rather than identifying why the built-in mutant differs, and Union checks enumeration.
One native builder family construction fault was run on the Go-oracle member, using the common preparation-success checker. The other two members were timed but were not separately faulted; no semantic verdict is assigned to builders.
The 13-row count appears to omit family grouping of the three product declarations. Numbered wrappers are verified programmatically rather than pretending that a filename alone establishes scope. The complete member lists and source locations are in groups.json and scope.json.
No repo-wide uniqueness, full-package per-mutant run, exhaustive native compiler corpus matrix or exact coverage profile was completed. Empty-answer probes are excluded from all production kills.
Setup: warm Go 1.27.1, nproc 5, no cloud setup. npm and corpus installation logs are saved. Timing driver wall costs are in commands.jsonl. Audit elapsed time is recorded in completion.txt.

Rebuild measurements, from actual build-cache miss lines:

```json
{
  "M1": [
    {
      "product": "compiler-agreement-go-oracle",
      "seconds": 3.06,
      "log": "M1.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2",
      "seconds": 10.9,
      "log": "M1.initial.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2-ignored-allowemptycatch",
      "seconds": 10.8,
      "log": "M1.initial.log"
    },
    {
      "product": "lint-decoded-options-native-setup-v2",
      "seconds": 1.83,
      "log": "M1.initial.log"
    },
    {
      "product": "lint-decoded-options-go-oracle-setup-v2",
      "seconds": 3.88,
      "log": "M1.initial.log"
    }
  ],
  "M2": [
    {
      "product": "compiler-agreement-go-oracle",
      "seconds": 2.97,
      "log": "M2.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2",
      "seconds": 10.74,
      "log": "M2.initial.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2-ignored-allowemptycatch",
      "seconds": 10.65,
      "log": "M2.initial.log"
    },
    {
      "product": "lint-decoded-options-native-setup-v2",
      "seconds": 1.76,
      "log": "M2.initial.log"
    },
    {
      "product": "lint-decoded-options-go-oracle-setup-v2",
      "seconds": 3.18,
      "log": "M2.initial.log"
    }
  ],
  "M3": [
    {
      "product": "compiler-agreement-go-oracle",
      "seconds": 3.05,
      "log": "M3.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2",
      "seconds": 10.88,
      "log": "M3.initial.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2-ignored-allowemptycatch",
      "seconds": 10.98,
      "log": "M3.initial.log"
    },
    {
      "product": "lint-decoded-options-native-setup-v2",
      "seconds": 2.89,
      "log": "M3.initial.log"
    },
    {
      "product": "lint-decoded-options-go-oracle-setup-v2",
      "seconds": 4.77,
      "log": "M3.initial.log"
    }
  ],
  "M4": [
    {
      "product": "compiler-agreement-go-oracle",
      "seconds": 3.46,
      "log": "M4.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2",
      "seconds": 11.28,
      "log": "M4.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2-ignored-allowemptycatch",
      "seconds": 10.72,
      "log": "M4.log"
    },
    {
      "product": "lint-decoded-options-native-setup-v2",
      "seconds": 7.85,
      "log": "M4.log"
    },
    {
      "product": "lint-decoded-options-go-oracle-setup-v2",
      "seconds": 4.38,
      "log": "M4.log"
    }
  ],
  "P1": [
    {
      "product": "compiler-agreement-go-oracle",
      "seconds": 3.59,
      "log": "P1.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2",
      "seconds": 11.62,
      "log": "P1.log"
    },
    {
      "product": "lint-decoded-options-lowered-setup-v2-ignored-allowemptycatch",
      "seconds": 10.9,
      "log": "P1.log"
    },
    {
      "product": "lint-decoded-options-native-setup-v2",
      "seconds": 8.69,
      "log": "P1.log"
    },
    {
      "product": "lint-decoded-options-go-oracle-setup-v2",
      "seconds": 4.09,
      "log": "P1.log"
    }
  ]
}
```

Total elapsed setup and audit: approximately 27 minutes. Warm setup skipped. Exact timing costs and process wall times are saved; installation stages lack independent wall measurements.

Large raw logs are preserved losslessly as .log.gz. Use gzip -dc to read them; commands retain their original output filenames. No source mutations remain in the pushed tree.
