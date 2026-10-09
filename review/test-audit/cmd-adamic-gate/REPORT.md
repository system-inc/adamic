Starting origin/main: 470cc0cde025aee921b7c7856bebe2465c1461b1; eight rows, no families or helpers.
Clean whole-package baseline passed in 9.046 binary seconds; no skips.
Seven rows sacred on production kills; one construction row setup-check.
Nineteen production mutants, one construction break and twelve separate probes.
Evidence branch test-audit/cmd-adamic-gate; production source restored.

```json
[
  {
    "test": "TestAnchoredSelectors",
    "package": "cmd/adamic-gate",
    "file": "cmd/adamic-gate/main_test.go",
    "seconds": 9.828,
    "oracle": "Go test actually executes the generated selectors; sorted leaf names must equal a self-written exact list. M01 produced three leaves, the correct count, with wrong membership.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M04"
    ],
    "unique_kills": [
      "M01",
      "M04"
    ],
    "last_proven_fail": "M04: main_test.go:69: Go executed [TestParent/internal/load/testdata/0.1/compile/main.a TestParent/internal/oracle/testdata/a.a TestParent/internal/oracle/testdata/b+.a TestParentExtra TestParentExtra], want exactly [TestParent/internal/load/testdata/0.1/compile/main.a TestParent/internal/oracle/testdata/a.a TestParent/internal/oracle/testdata/b+.a]",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestAnchoredSelectors",
      "TestCoverageRejectsOverlapAndUnplannedTests",
      "TestRawEvidenceKeepsSkipReasonsAndFailures",
      "TestEveryTestIsParallelOrSaysWhy",
      "TestCheckpointKeysIncludeAllExecutionInputs",
      "TestPackageCheckpointRequiresIntactCompleteEvidence",
      "TestTypeAwareChildrenAndParentCost",
      "TestResumeCleanupPreservesForeignPaths"
    ],
    "evidence": "ADAMIC_MUTANT=M04 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-gate/ -run . > M04.log 2>&1; main_test.go:69: Go executed [TestParent/internal/load/testdata/0.1/compile/main.a TestParent/internal/oracle/testdata/a.a TestParent/internal/oracle/testdata/b+.a TestParentExtra TestParentExtra], want exactly [TestParent/internal/load/testdata/0.1/compile/main.a TestParent/internal/oracle/testdata/a.a TestParent/internal/oracle/testdata/b+.a]",
    "incidental_probe_failures": [
      "P03"
    ]
  },
  {
    "test": "TestCoverageRejectsOverlapAndUnplannedTests",
    "package": "cmd/adamic-gate",
    "file": "cmd/adamic-gate/main_test.go",
    "seconds": 0.003,
    "oracle": "Self-written ownership maps and refusals: valid evidence, exact overlap and unplanned messages, and no ancestor satisfying a missing child.",
    "oracle_kind": "self",
    "kills": [
      "M05",
      "M06",
      "M07",
      "M08"
    ],
    "unique_kills": [
      "M05",
      "M07",
      "M08"
    ],
    "last_proven_fail": "M08: main_test.go:81: [unplanned test: p::TestParent]",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestAnchoredSelectors",
      "TestCoverageRejectsOverlapAndUnplannedTests",
      "TestRawEvidenceKeepsSkipReasonsAndFailures",
      "TestEveryTestIsParallelOrSaysWhy",
      "TestCheckpointKeysIncludeAllExecutionInputs",
      "TestPackageCheckpointRequiresIntactCompleteEvidence",
      "TestTypeAwareChildrenAndParentCost",
      "TestResumeCleanupPreservesForeignPaths"
    ],
    "evidence": "ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-gate/ -run . > M08-rerun.log 2>&1; main_test.go:81: [unplanned test: p::TestParent]",
    "vacuous_subcases": [
      "P02: initial valid first/second len(problems)==0 checks pass; missing produced evidence then fails at main_test.go:88."
    ]
  },
  {
    "test": "TestRawEvidenceKeepsSkipReasonsAndFailures",
    "package": "cmd/adamic-gate",
    "file": "cmd/adamic-gate/main_test.go",
    "seconds": 0.003,
    "oracle": "Self-written raw JSON, preserved skip text and failure bytes, terminal count 2, cache-hit count 1, and rejection of corrupt JSON.",
    "oracle_kind": "self",
    "kills": [
      "M09",
      "M10",
      "M11"
    ],
    "unique_kills": [
      "M09",
      "M10",
      "M11"
    ],
    "last_proven_fail": "M11: main_test.go:123: lost raw evidence: [{p TestSkip skip     probe_test.go:7: needs external corpus",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P03"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestAnchoredSelectors",
      "TestCoverageRejectsOverlapAndUnplannedTests",
      "TestRawEvidenceKeepsSkipReasonsAndFailures",
      "TestEveryTestIsParallelOrSaysWhy",
      "TestCheckpointKeysIncludeAllExecutionInputs",
      "TestPackageCheckpointRequiresIntactCompleteEvidence",
      "TestTypeAwareChildrenAndParentCost",
      "TestResumeCleanupPreservesForeignPaths"
    ],
    "evidence": "ADAMIC_MUTANT=M11 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-gate/ -run . > M11-rerun.log 2>&1; main_test.go:123: lost raw evidence: [{p TestSkip skip     probe_test.go:7: needs external corpus"
  },
  {
    "test": "TestEveryTestIsParallelOrSaysWhy",
    "package": "cmd/adamic-gate",
    "file": "cmd/adamic-gate/parallel_test.go",
    "seconds": 0.288,
    "oracle": "Self-written repository parallel policy and serial baseline. Checks suite construction; its production scan helpers live in parallel_test.go.",
    "oracle_kind": "self",
    "kills": [
      "S01"
    ],
    "unique_kills": [
      "S01"
    ],
    "last_proven_fail": "S01: parallel_test.go:137: cmd/adamic-gate/main_test.go\tTestAnchoredSelectors: call the test parameter's Parallel() as the first statement, or add // Not parallel: <shared state it touches> directly above the test with a non-empty reason",
    "verdict": "setup-check",
    "subsumed_by": null,
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestAnchoredSelectors",
      "TestCoverageRejectsOverlapAndUnplannedTests",
      "TestRawEvidenceKeepsSkipReasonsAndFailures",
      "TestEveryTestIsParallelOrSaysWhy",
      "TestCheckpointKeysIncludeAllExecutionInputs",
      "TestPackageCheckpointRequiresIntactCompleteEvidence",
      "TestTypeAwareChildrenAndParentCost",
      "TestResumeCleanupPreservesForeignPaths"
    ],
    "evidence": "ADAMIC_MUTANT=S01 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-gate/ -run . > S01.log 2>&1; parallel_test.go:137: cmd/adamic-gate/main_test.go\tTestAnchoredSelectors: call the test parameter's Parallel() as the first statement, or add // Not parallel: <shared state it touches> directly above the test with a non-empty reason"
  },
  {
    "test": "TestCheckpointKeysIncludeAllExecutionInputs",
    "package": "cmd/adamic-gate",
    "file": "cmd/adamic-gate/resume_test.go",
    "seconds": 0.003,
    "oracle": "Self-written key inequalities for plan, context, shard, package, patterns, file bytes/mode, environment, tools and external input; equality for normalized random Go scratch paths.",
    "oracle_kind": "self",
    "kills": [
      "M12",
      "M13",
      "M14"
    ],
    "unique_kills": [
      "M12",
      "M13",
      "M14"
    ],
    "last_proven_fail": "M14: resume_test.go:52: external input permissions absent from key",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P04",
      "P05",
      "P06",
      "P07"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestAnchoredSelectors",
      "TestCoverageRejectsOverlapAndUnplannedTests",
      "TestRawEvidenceKeepsSkipReasonsAndFailures",
      "TestEveryTestIsParallelOrSaysWhy",
      "TestCheckpointKeysIncludeAllExecutionInputs",
      "TestPackageCheckpointRequiresIntactCompleteEvidence",
      "TestTypeAwareChildrenAndParentCost",
      "TestResumeCleanupPreservesForeignPaths"
    ],
    "evidence": "ADAMIC_MUTANT=M14 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-gate/ -run . > M14-rerun.log 2>&1; resume_test.go:52: external input permissions absent from key",
    "vacuous_subcases": [
      "P07: random work-directory equivalence passes with empty keys; changed compiler flags then fail at resume_test.go:75."
    ]
  },
  {
    "test": "TestPackageCheckpointRequiresIntactCompleteEvidence",
    "package": "cmd/adamic-gate",
    "file": "cmd/adamic-gate/resume_test.go",
    "seconds": 0.004,
    "oracle": "Self-written positive found control, key mismatch refusal, checksum and terminal error text, complete failed-log reuse, and absent checkpoint handling. Some negative cases check only non-nil error; invocation controls use production patterns/testArgs helpers.",
    "oracle_kind": "self",
    "kills": [
      "M06",
      "M15",
      "M16"
    ],
    "unique_kills": [
      "M15",
      "M16"
    ],
    "last_proven_fail": "M16: resume_test.go:114: changed log accepted <nil>",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P08"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestAnchoredSelectors",
      "TestCoverageRejectsOverlapAndUnplannedTests",
      "TestRawEvidenceKeepsSkipReasonsAndFailures",
      "TestEveryTestIsParallelOrSaysWhy",
      "TestCheckpointKeysIncludeAllExecutionInputs",
      "TestPackageCheckpointRequiresIntactCompleteEvidence",
      "TestTypeAwareChildrenAndParentCost",
      "TestResumeCleanupPreservesForeignPaths"
    ],
    "evidence": "ADAMIC_MUTANT=M16 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-gate/ -run . > M16-rerun.log 2>&1; resume_test.go:114: changed log accepted <nil>",
    "incidental_probe_failures": [
      "P01",
      "P02",
      "P03"
    ]
  },
  {
    "test": "TestTypeAwareChildrenAndParentCost",
    "package": "cmd/adamic-gate",
    "file": "cmd/adamic-gate/resume_test.go",
    "seconds": 0.004,
    "oracle": "Self-written source-table labels, predicted costs 5 and 7, largest child, and differing plan digests. Table contents belong to Adamic, not an external oracle.",
    "oracle_kind": "self",
    "kills": [
      "M17",
      "M18"
    ],
    "unique_kills": [
      "M17",
      "M18"
    ],
    "last_proven_fail": "M18: resume_test.go:143: cannot enumerate TestVolumeAgreementAndMutants",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P09",
      "P10",
      "P11"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestAnchoredSelectors",
      "TestCoverageRejectsOverlapAndUnplannedTests",
      "TestRawEvidenceKeepsSkipReasonsAndFailures",
      "TestEveryTestIsParallelOrSaysWhy",
      "TestCheckpointKeysIncludeAllExecutionInputs",
      "TestPackageCheckpointRequiresIntactCompleteEvidence",
      "TestTypeAwareChildrenAndParentCost",
      "TestResumeCleanupPreservesForeignPaths"
    ],
    "evidence": "ADAMIC_MUTANT=M18 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-gate/ -run . > M18-rerun.log 2>&1; resume_test.go:143: cannot enumerate TestVolumeAgreementAndMutants"
  },
  {
    "test": "TestResumeCleanupPreservesForeignPaths",
    "package": "cmd/adamic-gate",
    "file": "cmd/adamic-gate/resume_test.go",
    "seconds": 0.004,
    "oracle": "Self-written refusal and filesystem state: foreign path must remain, owned package scratch must disappear.",
    "oracle_kind": "self",
    "kills": [
      "M19"
    ],
    "unique_kills": [
      "M19"
    ],
    "last_proven_fail": "M19: resume_test.go:172: foreign scratch path accepted",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P12"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestAnchoredSelectors",
      "TestCoverageRejectsOverlapAndUnplannedTests",
      "TestRawEvidenceKeepsSkipReasonsAndFailures",
      "TestEveryTestIsParallelOrSaysWhy",
      "TestCheckpointKeysIncludeAllExecutionInputs",
      "TestPackageCheckpointRequiresIntactCompleteEvidence",
      "TestTypeAwareChildrenAndParentCost",
      "TestResumeCleanupPreservesForeignPaths"
    ],
    "evidence": "ADAMIC_MUTANT=M19 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-gate/ -run . > M19-rerun.log 2>&1; resume_test.go:172: foreign scratch path accepted"
  }
]
```

| ID | File:line on starting origin/main | Change | Failed rows |
|---|---|---|---|
| M01 | cmd/adamic-gate/main.go:527 | drop entire leaf escaping loop | TestAnchoredSelectors |
| M02 | cmd/adamic-gate/main.go:534 | drop prefix escaping | [] |
| M03 | cmd/adamic-gate/main.go:530 | change leaf end-anchor constant | [] |
| M04 | cmd/adamic-gate/main.go:534 | change parent end-anchor constant | TestAnchoredSelectors |
| M05 | cmd/adamic-gate/main.go:1016 | drop seen-test update | TestCoverageRejectsOverlapAndUnplannedTests |
| M06 | cmd/adamic-gate/main.go:1010 | flip shard mismatch condition | TestCoverageRejectsOverlapAndUnplannedTests, TestPackageCheckpointRequiresIntactCompleteEvidence |
| M07 | cmd/adamic-gate/main.go:976 | drop descendant ownership condition | TestCoverageRejectsOverlapAndUnplannedTests |
| M08 | cmd/adamic-gate/main.go:979 | change ancestor acceptance to false | TestCoverageRejectsOverlapAndUnplannedTests |
| M09 | cmd/adamic-gate/main.go:593 | drop output accumulation | TestRawEvidenceKeepsSkipReasonsAndFailures |
| M10 | cmd/adamic-gate/main.go:543 | flip skip reason condition | TestRawEvidenceKeepsSkipReasonsAndFailures |
| M11 | cmd/adamic-gate/main.go:603 | off-by-one terminal counter | TestRawEvidenceKeepsSkipReasonsAndFailures |
| M12 | cmd/adamic-gate/resume.go:271 | change checkpoint package input to empty | TestCheckpointKeysIncludeAllExecutionInputs |
| M13 | cmd/adamic-gate/resume.go:253 | drop environment input | TestCheckpointKeysIncludeAllExecutionInputs |
| M14 | cmd/adamic-gate/resume.go:135 | change file mode input to zero | TestCheckpointKeysIncludeAllExecutionInputs |
| M15 | cmd/adamic-gate/resume.go:357 | flip package terminal count condition | TestPackageCheckpointRequiresIntactCompleteEvidence |
| M16 | cmd/adamic-gate/resume.go:403 | flip checksum OR to AND | TestPackageCheckpointRequiresIntactCompleteEvidence |
| M17 | cmd/adamic-gate/resume.go:64 | flip positive parent residual condition | TestTypeAwareChildrenAndParentCost |
| M18 | cmd/adamic-gate/main.go:252 | flip AST parent-name condition | TestTypeAwareChildrenAndParentCost |
| M19 | cmd/adamic-gate/resume.go:476 | flip owned scratch-name condition | TestResumeCleanupPreservesForeignPaths |
| S01 | cmd/adamic-gate/main_test.go:14 | drop Parallel statement from suite construction | TestEveryTestIsParallelOrSaysWhy |

Survivors:
M02: removing parent escaping executes the additional TestParent/internal/load/testdata/0X1/compile/main.a in the diagnostic. Baseline executes only the requested 0.1 path and a.a.
M03: removing the leaf end anchor executes the additional TestParent/internal/oracle/testdata/a.a_suffix. Baseline executes only the two requested leaves. Both are unguarded by this whole-package matrix, not equivalent candidates.

Brief ambiguities, weaknesses and costs:
1. Go test itself is absent from the external-run examples, although this package explicitly asks Go which tests execute. It is named as the external-run oracle, with the handwritten requested list separately marked self.
2. Raw JSON and failed checkpoint fixtures are ordinary inputs to the production gate checker. They are not port mutants demonstrating an external harness witness, so they are judged with production mutants. The parallel-policy row is suite construction and receives setup-check.
3. The parallel scanner walks review/. Saving validation copies as *_test.go accidentally made them live construction inputs, introducing an unrelated S01 violation in initial M06..M19 runs. Files were renamed *.go.fixture and every affected whole-package column rerun. No contaminated failure contributes to a verdict. This cost fourteen repeated package runs.
4. A runtime if around the first Parallel statement changes its AST even when the selector is disabled. S01 therefore uses an actual source edit and a separate build after all other runs finish. The unmutated switched control was rechecked first.
5. Selector scaffolding initially had one-line function separators, AST offsets taken from modified rather than original files, and an incorrectly typed helper parameter. These were repaired before mutant observations. The recorded passing switched control excludes those failed builds from the matrix.
6. Empty patterns and predictions cause caller index panics. Every row was rerun alone for P01 and P10; unexecuted rows were never counted as survivors. Empty probes do not support uniqueness or sacred verdicts.
7. The parallel scan has no production answer entry, so vacuous is null. For other rows, probe_kills contains only their own answer entries. Incidental failures from preparation helpers appear separately. The coverage and key rows have observed positive comparisons that accept empty answers before later checks fail; they are listed as vacuous_subcases, not whole-row vacuity.
8. M01 produces three executed leaves, exactly the expected count, but the wrong names. The exact-name oracle catches this; counting alone would not.
9. Checkpoint controls construct invocation arguments with the same production patterns/testArgs helpers being validated. Some negative cases require only a non-nil error. Positive found checks and explicit checksum/terminal messages strengthen other branches, but the oracle remains self.
10. Two selector survivors needed separate diagnostic fixtures: the original 0X1 leaf is extra.a, so it does not expose an unescaped prefix with main.a; it also has no a.a_suffix leaf. Diagnostics demonstrate missing behaviors without changing the package's oracle.
11. The requested scope comes from go test -list: embedded TestParent/TestOther text creates subprocess fixtures, not extra package rows. No top-level row skipped or required an installable opt-in.
12. The nineteen production mutants cover reachable helper implementations. The tests do not directly run gate planning, shard execution, merge/compare, live execution identity or CLI dispatch. Seven sacred verdicts are relative to this finite mutant set and package, not proof of the entire gate or repo-wide uniqueness. No other repository package tests ran; scratch diagnostics copied the same production selector code solely to witness survivors.
13. P07's standalone empty body made regexp unused. Its import was dropped and the repaired diff vetted. M01 drops the entire escaping loop, avoiding vet's self-assignment warning. Every standalone diff applies to the immutable base and compiles independently of the switch.

Evidence sources use .go.fixture so they do not create repository packages or policy-scan inputs. main.go, resume.go and main_test.go are restored, and audit_switch.go is removed. No pull request or main push occurred.

Timing:
```json
{
  "setup": {
    "npm_seconds": 0.7161337270008516,
    "npm_exit": 0,
    "nproc": "5",
    "tools": "go version go1.27.1 linux/amd64",
    "setup": "warm env works; skipped"
  },
  "switched_build": {
    "seconds": 0.5128781999992498,
    "exit": 0
  },
  "construction_build": {
    "wall": 0.5675152030016761,
    "exit": 0
  },
  "baseline_wall": 9.24176548000105,
  "baseline_binary": 9.046,
  "restored_binary": 8.829,
  "final_matrix_cumulative_wall": 366.998766632998,
  "final_matrix_cumulative_binary": 358.276,
  "all_logged_run_cumulative_wall": 576.239312967009,
  "validation_cumulative_wall": 12.579982995008322,
  "recorded_at": "2026-10-09T09:42:09.530904+00:00",
  "note": "Cumulative run times are work totals, not elapsed audit time; reruns and diagnostics overlapped. Scaffolding debugging time is not instrumented."
}
```
