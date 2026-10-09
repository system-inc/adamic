Starting commit: 7b18d0576930caca4e22ce2eef92fcf563af52d0, fetched origin/main.
Scope: one top-level test, five subcases, no families, skips, helpers, or witnesses.
Clean baseline passed in 0.123 s; isolated median 0.132 s; nproc 5.
Verdict: sacred within this package, three unique production kills; empty-answer probe caught.
Survivor: M4 changes a valid selection from index 0 to 1; successful selection is unguarded here.

```json
[
  {
    "test": "TestViewMixedUnionUnknownAndUnavailable",
    "package": "internal/javascript",
    "file": "internal/javascript/view_unions_mixed_test.go:5",
    "seconds": 0.132,
    "oracle": "Self-written exact panic message, empty stdout, and exit 70; Node executes the production selector. Exact stderr guards against a different failure sharing exit 70. No outside authority supplied or checked. Only rejected inputs are covered.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [
      "M1",
      "M2",
      "M3"
    ],
    "last_proven_fail": "M3: view_unions_mixed_test.go:15: exit 0, stdout \"\", stderr \"\"; want 70, \"\", \"adamic: panic: field read failed: view.value matches no member of Left | Right; expected Left | Right, found object\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_MUTANT=M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/javascript/ -run . > review/test-audit/internal-javascript/M3.log 2>&1; view_unions_mixed_test.go:15: exit 0, stdout \"\", stderr \"\"; want 70, \"\", \"adamic: panic: field read failed: view.value matches no member of Left | Right; expected Left | Right, found object\\n\""
  }
]
```

CODE UNDER TEST, declared before mutants: production embedded JavaScript adamicViewMixedUnionSelect in internal/javascript/view_unions_mixed.go. This is the complete reached production function list. The test concatenates the constant directly, so it does not reach MixedUnionRuntime(), JavaScript(), JavaScriptWith(), or viewStringUndefined().
ORACLE, declared before mutants: self-written panic text and exit status, checked against Node execution by runViewNode. Node is an executor here, not an independent expected-answer authority. The testing runtime and comparison helper were not mutated.

Fixed menu recorded before matrix execution: flip the kind comparison; drop literal guard; drop complete object-contract guard block; off-by-one successful index. All are ordinary menu mutants. P1 is an empty-answer probe only. switch.diff records the scratch switch; production was restored afterward. Each individual diff applies to the starting origin/main and passed go vet ./internal/javascript/ (standalone-validation.json). No C or native rebuild was required.

| ID | origin/main file:line | Change | Failed rows |
| --- | --- | --- | --- |
| M1 | internal/javascript/view_unions_mixed.go:14 | member.kind !== snapshot.kind becomes === | TestViewMixedUnionUnknownAndUnavailable |
| M2 | internal/javascript/view_unions_mixed.go:15 | Drop literal guard statement | TestViewMixedUnionUnknownAndUnavailable |
| M3 | internal/javascript/view_unions_mixed.go:16 | Drop entire object-contract guard block, lines 16 through 18 | TestViewMixedUnionUnknownAndUnavailable |
| M4 | internal/javascript/view_unions_mixed.go:19 | return index becomes return index + 1 | |
| P1 | internal/javascript/view_unions_mixed.go:11 | Return undefined at selector entry (probe, excluded from kills) | TestViewMixedUnionUnknownAndUnavailable |

M4 survivor witness: node review/test-audit/internal-javascript/M4-witness-before.mjs prints 0; node review/test-audit/internal-javascript/M4-witness-after.mjs prints 1 for number 42 selecting its sole number member. This is changed behavior, not an equivalent candidate. All five subcases of the row reject inputs; no successful selection is asserted. There are no other survivors.

Commands: source /workspace/adamic-tools/env.sh; go test -list . ./internal/javascript/; timeout 120 go test -json -count=1 -timeout 90s ./internal/javascript/ -run .; three isolated go test -count=1 -timeout 90s ./internal/javascript/ -run '^TestViewMixedUnionUnknownAndUnavailable$'; matrix command in rows.json with ADAMIC_MUTANT set separately to M1, M2, M3, M4, P1. All test output went to saved logs. No run cooked or panicked. Final restored-baseline.log passed. matrix.json records exact failed rows/subcases and binary elapsed times.

Brief friction and limits:
- No Your rows section was supplied. Used the entire package listing; no supplied names could be checked for movement or disappearance.
- Warm env.sh worked, but cohere and its TypeScript submodule were missing. The first listing failed before a baseline could run. Initialized both; default SSH port 22 was refused, then used CLAUDE.md's documented HTTPS URL. This was a setup failure, not a red test baseline.
- The brief's usual Lower empty-answer example does not apply: this test directly calls an embedded JavaScript selector. Probed that selector, not the Go wrapper it never calls.
- Only one production function is reached, so spreading mutants over multiple functions would leave the specified code-under-test scope. Mutants instead span four distinct guards/return locations.
- Repository-wide uniqueness is unknown and left for central replay. No other package test was run. go vet covered only the mutated package.
- Full setup wall time and isolated build wall time were not instrumented. No claim of exact totals for them is possible.

Timing: warm setup skipped (Go 1.27.1); npm ci command 0.638 s, npm itself reported 622 ms. Fetch and branch creation 2.635 s. Submodule initialization required approximately a minute, exact timing unmeasured. Initial listing/build plus baseline command completed between 09:54:00 and 09:54:09 UTC. Clean binary 0.123 s. Three row binaries 0.132, 0.130, 0.135 s, median 0.132 s. Individual standalone vet checks 0.114 to 0.115 s each, total 0.572 s. Matrix binary timings appear below. No native rebuilds. No opt-in tests or skipped rows were present. Did not measure other packages or add positive-input coverage.

```json
{
  "M1": {
    "failed_rows": [
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "failed_subcases": [
      "TestViewMixedUnionUnknownAndUnavailable/null_is_not_undefined",
      "TestViewMixedUnionUnknownAndUnavailable/literal_false"
    ],
    "binary_seconds": 0.128
  },
  "M2": {
    "failed_rows": [
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "failed_subcases": [
      "TestViewMixedUnionUnknownAndUnavailable/literal_false"
    ],
    "binary_seconds": 0.128
  },
  "M3": {
    "failed_rows": [
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "failed_subcases": [
      "TestViewMixedUnionUnknownAndUnavailable/missing_adapter",
      "TestViewMixedUnionUnknownAndUnavailable/missing_contract"
    ],
    "binary_seconds": 0.129
  },
  "M4": {
    "failed_rows": [],
    "failed_subcases": [],
    "binary_seconds": 0.123
  },
  "P1": {
    "failed_rows": [
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "failed_subcases": [
      "TestViewMixedUnionUnknownAndUnavailable/unknown",
      "TestViewMixedUnionUnknownAndUnavailable/missing_adapter",
      "TestViewMixedUnionUnknownAndUnavailable/missing_contract",
      "TestViewMixedUnionUnknownAndUnavailable/null_is_not_undefined",
      "TestViewMixedUnionUnknownAndUnavailable/literal_false"
    ],
    "binary_seconds": 0.126
  }
}
```
