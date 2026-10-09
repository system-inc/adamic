Both assigned rows defended by unique production-mutant catches across the whole four-test package.
Two aimed mutants completed within the 90-second binary budget.
No tests, oracle or production changes remain; evidence is preserved in this session directory.

[
  {
    "test": "TestMessageRefusalsMatchGo",
    "package": "stage1/cohere/lint/helpers",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestHelpersMatchCohere"
    ],
    "defense": "defended",
    "unique_mutant": "R1 stage1/cohere/lint/helpers/policy_message.ts:15",
    "attempts": [
      {
        "mutant": "R1",
        "file_line": "stage1/cohere/lint/helpers/policy_message.ts:15",
        "change": "if(template < 0) { -> if(template < -1) {",
        "rows_failed": [
          "TestMessageRefusalsMatchGo"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/lint-helper-defense/cache/R1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/helpers/ -run . > R1.log 2>&1; helpers_test.go:236: guard 0 /tmp/adamic-gate/TestMessageRefusalsMatchGo2809949762/001/helpers: exit status 70 stderr \"adamic: panic: missing option value\\n\", Go \"adamic: panic: policy/messages: missing/rule has no message \\\"missing\\\"\\n\"",
    "rows_passed": [
      "TestHelpersMatchCohere",
      "TestHelperMutants",
      "TestKnownGapsAreExplicit"
    ],
    "bounded": false,
    "matrix_rows": [
      "TestHelpersMatchCohere",
      "TestHelperMutants",
      "TestMessageRefusalsMatchGo",
      "TestKnownGapsAreExplicit"
    ]
  },
  {
    "test": "TestKnownGapsAreExplicit",
    "package": "stage1/cohere/lint/helpers",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestHelpersMatchCohere"
    ],
    "defense": "defended",
    "unique_mutant": "G1 stage1/cohere/lint/helpers/strict_options.ts:11",
    "attempts": [
      {
        "mutant": "G1",
        "file_line": "stage1/cohere/lint/helpers/strict_options.ts:11",
        "change": "if(kind === 'unsupported') { this.unsupported = 'custom or unsupported Go option type'; return false; } -> (drop statement)",
        "rows_failed": [
          "TestKnownGapsAreExplicit"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/lint-helper-defense/cache/G1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/helpers/ -run . > G1.log 2>&1; helpers_test.go:260: output line 2: got \"valid\", Go \"NotYet: custom or unsupported Go option type\"",
    "rows_passed": [
      "TestHelpersMatchCohere",
      "TestHelperMutants",
      "TestMessageRefusalsMatchGo"
    ],
    "bounded": false,
    "matrix_rows": [
      "TestHelpersMatchCohere",
      "TestHelperMutants",
      "TestMessageRefusalsMatchGo",
      "TestKnownGapsAreExplicit"
    ]
  }
]

## Findings and evidence

Read CODE-AND-ORACLE.md, coverage-diffs.json, plan.json and corpus-analysis.json for expected-answer authorities and exclusive runtime paths. Source locations use starting origin/main 157a43552015f41a79331949c2e82b6f8c7caaab. Both standalone diffs apply cleanly. R1 compiled and ran natively in its whole-package matrix. G1 compiled in the agreement row and also in a separate sanitized native witness. The witness's native output is valid; clean Node output is NotYet: custom or unsupported Go option type. Both matrices ran every current top-level Test, including the builtin witness, and list all passing rows.

## Places the brief was unclear or cost time

- Both prior evidence commands and failing lines were truncated in the brief. All audit reports, code/oracle notes and rows were fetched with the full refspec and read before the mutants.
- The current origin/main differs from the audit's ce1c5a2f. Current source coordinates use 157a4355. All four original test names remain and no top-level test was added or vanished in this package.
- Go coverpkg cannot instrument the TypeScript port. Compiler coverage gave zero exclusive blocks for both subjects, so unchanged-test Node V8 coverage was added. V8 ranges map to transformed JavaScript, not directly to original TypeScript line numbers; the source snippets and original mutant sites make that distinction explicit.
- The refusal is caught on native stderr despite retaining exit 70. It tests Go's reason for refusal, beyond just detecting any failure. The row does not assert stdout, as confirmed by reading its body; no new stdout-insertion mutant was used or credited.
- The gap row's NotYet labels are handwritten. The compare helper labels its expected side Go, but this particular expected answer does not come from Go or an external authority. Source Node is another execution of the code under test, not an independent oracle for those labels.
- The builtin mutant witness is part of both full matrices. It passes both production mutants; that is a recorded result, not a separate proof of its local comparison. No harness weakening was needed for the two assigned rows.
- The brief permits up to three aimed attempts. Each first attempt produced a unique whole-package catch, so no further mutations were necessary. No separate executor twins or cost assertions required additional attempts.

## Names, time and limits

No assigned row remained undefended. Their names correspond to their assertions: precise Go refusal matching and explicit gap labels. The matching promise covers stderr and exit status, not stdout. The gap-label authority is self.

Warm env.sh worked, setup 0 seconds; nproc 5; stage3/api npm ci reported 378 ms. Clean baseline passed all four tests at 74.648 binary seconds. The three coverage command walls totaled about 28.159 seconds. Whole-package R1 and G1 binary times were 44.775 and 45.129 seconds. Per-command walls and the separate G1 native rebuild time are recorded in JSON; rebuild phases inside Go test were not individually timed. The session took approximately 9 minutes before pushing.

No other package tests, exhaustive fuzzer, repo-wide uniqueness replay, test deletion/rewrite, PR or main push was performed. No opt-in or skipped row was found. Raw logs and coverage are explicitly included even where repository ignore rules would otherwise omit them.
