Audited all four package rows at origin/main b955d3030e79e9f01d12a96ab1fb4b167fc91c0c.
All four are sacred within the complete package matrix; repo-wide uniqueness is unknown.
Twelve production mutants and six separate empty-answer probes; nproc=5.
One survivor: M10 changes multiline paragraph output and every package row passes it.
All eighteen standalone diffs apply and pass go vet; source restored and package green.

```json
[
  {
    "test": "TestPhysicalInventoryBoundaries",
    "package": "cmd/adamic-stage1-progress",
    "file": "cmd/adamic-stage1-progress/main_test.go:11",
    "seconds": 0.003,
    "oracle": "Hand-written physical line counts and production-path inclusion booleans.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M3 main_test.go:18: lines=0 want 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M3 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-stage1-progress/ -run . > review/test-audit/cmd-adamic-stage1-progress/M3.log 2>&1; main_test.go:18: lines=0 want 1",
    "timing_samples": [
      0.003,
      0.004,
      0.003
    ],
    "own_entry_probes": [
      "P1",
      "P2"
    ]
  },
  {
    "test": "TestDeclarationCreditIsBoundedAndFailsClosed",
    "package": "cmd/adamic-stage1-progress",
    "file": "cmd/adamic-stage1-progress/main_test.go:27",
    "seconds": 0.003,
    "oracle": "Hand-written credited declaration line set, including attached documentation and closing line; missing name must return some error, without checking error identity.",
    "oracle_kind": "self",
    "kills": [
      "M4",
      "M5",
      "M6"
    ],
    "unique_kills": [
      "M4",
      "M5",
      "M6"
    ],
    "last_proven_fail": "M6 main_test.go:38: missing declaration silently credited",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M6 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-stage1-progress/ -run . > review/test-audit/cmd-adamic-stage1-progress/M6.log 2>&1; main_test.go:38: missing declaration silently credited",
    "timing_samples": [
      0.003,
      0.003,
      0.003
    ],
    "own_entry_probes": [
      "P3"
    ]
  },
  {
    "test": "TestUnionDoesNotCountSharedLinesTwice",
    "package": "cmd/adamic-stage1-progress",
    "file": "cmd/adamic-stage1-progress/main_test.go:41",
    "seconds": 0.005,
    "oracle": "Hand-written deduplicated line count, denominator, 75 percent and complete count; incompatible pins must return some error, without checking error identity.",
    "oracle_kind": "self",
    "kills": [
      "M7",
      "M8"
    ],
    "unique_kills": [
      "M7",
      "M8"
    ],
    "last_proven_fail": "M8 main_test.go:49: cannot union different cohere pins",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P4"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M8 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-stage1-progress/ -run . > review/test-audit/cmd-adamic-stage1-progress/M8.log 2>&1; main_test.go:49: cannot union different cohere pins",
    "timing_samples": [
      0.005,
      0.004,
      0.008
    ],
    "own_entry_probes": [
      "P4"
    ]
  },
  {
    "test": "TestGitSnapshotAndEvidenceBoundary",
    "package": "cmd/adamic-stage1-progress",
    "file": "cmd/adamic-stage1-progress/main_test.go:59",
    "seconds": 0.19,
    "oracle": "Hand-written progress totals/status counts and exact pending/integrated ref lists for committed Git fixtures. Git constructs and reads snapshots, not a second expected progress report. Unknown slices need some error, without checking error identity; gap-summary text is not asserted.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2",
      "M9",
      "M11",
      "M12"
    ],
    "unique_kills": [
      "M9",
      "M11",
      "M12"
    ],
    "last_proven_fail": "M12 main_test.go:109: unmapped stage1/cohere/gitignore: add reviewed source coverage before reporting",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P5",
      "P6"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M12 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-stage1-progress/ -run . > review/test-audit/cmd-adamic-stage1-progress/M12.log 2>&1; main_test.go:109: unmapped stage1/cohere/gitignore: add reviewed source coverage before reporting",
    "timing_samples": [
      0.155,
      0.19,
      0.222
    ],
    "own_entry_probes": [
      "P5",
      "P6"
    ]
  }
]
```

Table labels: physical = TestPhysicalInventoryBoundaries; declaration = TestDeclarationCreditIsBoundedAndFailsClosed; union = TestUnionDoesNotCountSharedLinesTwice; snapshot = TestGitSnapshotAndEvidenceBoundary.

| ID | File:line on origin/main | Change | Failed rows |
|---|---|---|---|
| M1 | cmd/adamic-stage1-progress/main.go:80 | Count CR instead of LF. | physical, snapshot |
| M2 | cmd/adamic-stage1-progress/main.go:87 | Excluded suffix _test.go becomes _audit_test.go. | physical, snapshot |
| M3 | cmd/adamic-stage1-progress/main.go:82 | Drop unterminated-last-line increment. | physical |
| M4 | cmd/adamic-stage1-progress/main.go:181 | Disable inclusion of attached documentation. | declaration |
| M5 | cmd/adamic-stage1-progress/main.go:184 | Declaration end bound <= becomes <. | declaration |
| M6 | cmd/adamic-stage1-progress/main.go:188 | Drop the complete missing-name validation loop. | declaration |
| M7 | cmd/adamic-stage1-progress/main.go:348 | Percentage multiplier 100 becomes 1. | union |
| M8 | cmd/adamic-stage1-progress/main.go:358 | Pin mismatch != becomes ==. | union |
| M9 | cmd/adamic-stage1-progress/main.go:436 | Pending prefix stage1- becomes stage2-. | snapshot |
| M10 | cmd/adamic-stage1-progress/main.go:419 | Paragraph separator space becomes |. |  |
| M11 | cmd/adamic-stage1-progress/main.go:139 | Batch blob buffer size+1 becomes size. | snapshot |
| M12 | cmd/adamic-stage1-progress/main.go:301 | Unknown-directory guard !known becomes known. | snapshot |

Survivor M10: firstParagraph("# Heading\n\none\ntwo\n\nnext\n") returns "one two" before and "one|two" after. The observer copies the exact switched production source and renames only the unused CLI main to permit a dedicated observer entry. No original test, assertion or oracle changes. Command: source /workspace/adamic-tools/env.sh && python3 review/test-audit/cmd-adamic-stage1-progress/observe-survivor.py; exact Go commands and outputs are in survivor-observation.json and survivor-{before,after}.log. This is changed behavior unguarded by every package row, not an equivalent candidate.

Brief feedback:

- The package name includes stage1, but this unit is a Go inventory command, not a .a/.ts port. Reading its actual implementation established that main.go is the production code to mutate. No port, compiler or external checker was mutated, and native product cache isolation was unnecessary.
- The external-run definition can be read too broadly when Git is used for fixture construction. This test runs real Git, but its expected progress counts and classifications are hand-written. I classified the oracle as self rather than treating every Git invocation as independent truth.
- Empty-answer probes for pointer-valued reports are underspecified. P4 and P5 return an empty &report{} with nil error, preserving a readable result while removing every answer field; they are not nil-pointer crash probes. Both rows reject the empty report. P3 returns an empty map.
- Multiple genuine entries need explicit attribution. The physical row directly calls physicalLines and production; the snapshot row directly calls measure and pending. I probed both entries for each. P1..P3 also break snapshot preparation indirectly, but its probe_kills and vacuity use only P5/P6. The full matrix retains all incidental failures.
- Positive fatal assertions stop later negative cases. credit, union and measure probes therefore do not establish what their later negative cases would do. No unexecuted subcase is labelled vacuous or proven.
- The unconditional npm ci requirement cost 770 ms even though this unit uses no Node dependencies. Setup was skipped because the warm environment worked.
- The four tests contain compound assertions and do not provide names for each subcase. A kill can stop at a positive precondition rather than exercising a later fail-closed assertion. The oracle descriptions explicitly retain the weak any-error checks.
- The brief gave no starting row list for this whole package, as intended. go test -list found four standalone rows, no generated families, helpers, witness rows or setup checks. No row skipped, no timeout occurred, and no narrowing was needed.

No audit-driver failure or invalid standalone diff occurred in this unit. M6 drops the whole validation loop so its range variables disappear too. All other deletion and bound changes compiled independently without cleanup repairs.

Timing and limits:

Warm setup skipped: 0 setup seconds. npm ci reported 770 ms. Go 1.27.1, Node 24.19.0, nproc 5. Clean coverage baseline test binary: 0.166 s. Restored package test binary: 0.184 s.
Restored CLI build: 0.713 wall s. Standalone go vet validations: 2.894 wall s. All 18 apply and compile. The first switched matrix control includes its compilation; the selector source was formatted before running it and reused through Go's build cache. No separate native builds occurred.
Twelve individual timing commands: 3.694 wall s including Go invocation, with test binary medians 0.003, 0.003, 0.005 and 0.190 s. Full matrix control plus 12 mutants plus 6 probes: 8.035 wall s; its summed test-binary printed times are 2.608 s. Survivor observer commands total 0.793 wall s.
Coverage baseline reaches 72.2 percent of statements and 13 functions. main and fail are not called by these tests. CLI output/flags, broader real-repository mappings, multiple-line gap summaries, malformed blob variants beyond the planted bound, and repo-wide uniqueness were not established. No other package tests ran. The separately built CLI was compiled, not executed. All production source is restored, no PR was opened and main was not pushed.
