u125 audited all three current top-level tests, no families or skips.
Starting origin/main: ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2; nproc 5; warm Go 1.27.1.
Clean baseline passed in 0.005s own binary, 0.175s shell build/run wall.
Nine construction mutations: seven caught, two witnessed changed-behavior survivors.
All three rows are setup-check; empty Merge fails every row; source restored.

CODE UNDER TEST: Go shard-output construction Merge and isCaseLine. Run is not reached.
ORACLE: handwritten output bytes and diagnostic substrings, all self. No Node/cohere/native execution.

[
  {
    "test": "TestMergePutsCasesBackInOrder",
    "package": "stage1/cohere/lint/shards",
    "file": "stage1/cohere/lint/shards/shards_test.go:8",
    "seconds": 0.002,
    "oracle": "Handwritten complete expected byte string for interleaved case blocks in ascending order.",
    "oracle_kind": "self",
    "kills": [
      "C1",
      "C2",
      "C7",
      "C9"
    ],
    "unique_kills": [],
    "last_proven_fail": "C9 shards_test.go:15: case 2 missing from every shard",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/shards/ -run . > /workspace/adamic/review/test-audit/stage1-cohere-lint-shards/C9.log 2>&1; shards_test.go:15: case 2 missing from every shard",
    "timing_samples": [
      0.003,
      0.002,
      0.002
    ],
    "mutation_kind": "construction",
    "construction_unique_kills": []
  },
  {
    "test": "TestMergeKeepsLinesThatOnlyLookLikeCases",
    "package": "stage1/cohere/lint/shards",
    "file": "stage1/cohere/lint/shards/shards_test.go:23",
    "seconds": 0.002,
    "oracle": "Handwritten complete expected byte string retaining indented case-like text and case 1x.",
    "oracle_kind": "self",
    "kills": [
      "C1",
      "C2",
      "C4",
      "C7",
      "C9"
    ],
    "unique_kills": [],
    "last_proven_fail": "C9 shards_test.go:27: shard 0: malformed case line \"case 1x\"",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/shards/ -run . > /workspace/adamic/review/test-audit/stage1-cohere-lint-shards/C9.log 2>&1; shards_test.go:27: shard 0: malformed case line \"case 1x\"",
    "timing_samples": [
      0.002,
      0.002,
      0.001
    ],
    "mutation_kind": "construction",
    "construction_unique_kills": [
      "C4"
    ]
  },
  {
    "test": "TestMergeRefusesAMissingOrRepeatedCase",
    "package": "stage1/cohere/lint/shards",
    "file": "stage1/cohere/lint/shards/shards_test.go:34",
    "seconds": 0.002,
    "oracle": "Handwritten diagnostic substrings for missing, repeated, non-case-start, empty-shard and extra-case inputs. Error required; complete diagnostic not compared.",
    "oracle_kind": "self",
    "kills": [
      "C1",
      "C2",
      "C3",
      "C8",
      "C9"
    ],
    "unique_kills": [],
    "last_proven_fail": "C9 shards_test.go:54: extra case: error <nil>, want \"2 cases printed, want the manifest's 1\"",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/shards/ -run . > /workspace/adamic/review/test-audit/stage1-cohere-lint-shards/C9.log 2>&1; shards_test.go:54: extra case: error <nil>, want \"2 cases printed, want the manifest's 1\"",
    "timing_samples": [
      0.002,
      0.003,
      0.002
    ],
    "mutation_kind": "construction",
    "construction_unique_kills": [
      "C3",
      "C8"
    ]
  }
]

| ID | Origin shards.go line | Change | Failing rows |
|---|---:|---|---|
| C1 | 78 | `change prefix constant` | TestMergePutsCasesBackInOrder, TestMergeRefusesAMissingOrRepeatedCase, TestMergeKeepsLinesThatOnlyLookLikeCases |
| C2 | 119 | `off-by-one block key` | TestMergeRefusesAMissingOrRepeatedCase, TestMergePutsCasesBackInOrder, TestMergeKeepsLinesThatOnlyLookLikeCases |
| C3 | 116 | `flip duplicate condition` | TestMergeRefusesAMissingOrRepeatedCase |
| C4 | 107 | `drop whole-line condition` | TestMergeKeepsLinesThatOnlyLookLikeCases |
| C5 | 149 | `off-by-one digit bound` |  |
| C6 | 146 | `change digit upper bound` |  |
| C7 | 119 | `off-by-one block slice` | TestMergePutsCasesBackInOrder, TestMergeKeepsLinesThatOnlyLookLikeCases |
| C8 | 123 | `flip case-count condition` | TestMergeRefusesAMissingOrRepeatedCase |
| C9 | 149 | `flip line terminator condition` | TestMergePutsCasesBackInOrder, TestMergeKeepsLinesThatOnlyLookLikeCases, TestMergeRefusesAMissingOrRepeatedCase |
| P1 | 85 | `empty Merge body` | TestMergeRefusesAMissingOrRepeatedCase, TestMergePutsCasesBackInOrder, TestMergeKeepsLinesThatOnlyLookLikeCases |

Survivors:
C5: clean Merge preserves "case 0\ncase \nfixed\ta\n"; mutant returns malformed case line "case ". Missing minimum-digit boundary guard in current tests.
C6: clean Merge returns all ten cases 0..9; mutant reports case 9 missing. Current positive case numbers do not cover digit 9.
Witness command: go run /tmp/u125/witness.go with clean source, then each standalone survivor applied. Outputs witness-clean.log, witness-C5.log, witness-C6.log; exact driver and command timings retained.
No equivalent candidates or unexplained survivors.

Brief ambiguities, interpretation and time costs:
- This package constructs the test suite's merged shard output. Its rows check that construction, so the brief's setup-check rule applies. C IDs deliberately identify construction edits. Their observed failures appear in kills for replay, but unique_kills stays empty because construction uniqueness does not establish production sacredness. construction_unique_kills records observed singleton catches separately.
- The stage1 path does not mean these tests execute a stage1 port. All bodies call Go Merge; none calls Run. Mutating .a/.ts, cohere or native lowering would not test the asserted construction. Native rebuilding is inapplicable.
- Three functions share Merge but assert different properties: order, preservation of non-header lines, and rejection errors. They are not input-only wrappers around a shared checker, so no family grouping.
- The negative row feeds deliberately bad shard output directly to Merge. This tests construction validation; it does not witness a weakened external agreement oracle, and no test/harness assertions were edited.
- Whole-package scope came from the current list, not stale filenames. No missing or opt-in tests, no helpers, no external tools required by these bodies.
- The one-line header prefix can fail negative tests through a different error. Only observed failures are reported; a production-worthy verdict is not inferred from those precondition failures.
- Error substring checks verify narrower behavior than full diagnostic equality. A different error containing the same substring could pass. Positive rows compare complete bytes, not counts or exit codes.
- A literal empty-answer return with its old body retained triggers vet unreachable-code diagnostics. P1 replaces the complete Merge body and passes vet. This remains a probe, never a construction or production kill.
- All nine menu choices were saved before testing. They cover framing, numbering, duplicate rejection, header recognition, digit bounds, block bytes and total-case counts. No inserted-statement supplemental mutants.
- Sequential switch-free Go mutations were used because this tiny package builds and runs in fractions of a second. This also directly validates every saved diff without selector dependencies. Build/check wall is recorded per mutation. The four-mutant rebuild limit concerns native port/compiler rebuilds, which this unit performs none of.
- npm ci in stage3/api was mandatory in the brief even though no test loads node_modules. Installation was not separately timed; no other Node dependency directory is reached.
- An initially broad caller search included generated lint tests and produced excessive output. It did not run any other package. Scope and mutations were taken only from the three bodies read in full and shards.go.
- No run reached 90 seconds, so no bounded matrix, timeout narrowing, panic reruns or unknown rows. No repository-wide uniqueness claim.
- Seven caught construction mutations prove these checks can fail; the two survivor witnesses also show concrete construction boundaries the tests miss. Neither survivor is called equivalent or silently treated as covered.

Timing and uncovered work:
Setup skipped, 0s; nproc 5. Baseline binary 0.005s; shell build/run 0.175s.
Nine isolated timing runs, ten matrix/probe runs and restored package run: 3.976s combined shell build/run wall. Standalone vet compile/check wall: 0.693s. Three survivor-witness build/runs: 0.708s. Individual measurements and exact commands retained.
Each row's seconds is the median of three fresh -count=1 own-binary ok lines. No native builds; npm install wall not recorded independently. Completed within the 20-minute budget.
Not covered: Run subprocess launch/count-only aggregation/concurrency/manifest handling; additional malformed initial headers, unterminated headers, signed numbers, huge numbers and CRLF; other Adamic packages; external/native lint behavior. Source restored, final clean test passed. No PR or main push.
