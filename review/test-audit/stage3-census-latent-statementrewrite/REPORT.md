Unit u162: one top-level test at origin/main cf735d9fba9e38de6368575e5630e44375a86eaf.
Clean whole-package baseline and three individual count=1 runs passed; no skips.
Verdict: setup-check; M1 and M2 are unique catches within this package.
M3 survives and demonstrably removes comments; P1 empty-entry probe is caught.
Evidence branch: test-audit/stage3-census-latent-statementrewrite; production source restored.

```json
[
  {
    "test": "TestMissingStatementFailsWithMethodName",
    "package": "stage3/census/latent/statementrewrite",
    "file": "stage3/census/latent/statementrewrite/main_test.go",
    "seconds": 0.04,
    "oracle": "Self-written CLI contract: nonzero exit must contain lowering.statement: expected exactly one function, found 0; output file must not exist. M2 demonstrates that an unrelated failure is rejected.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [
      "M1",
      "M2"
    ],
    "last_proven_fail": "M2: main_test.go:31: missing statement mutant survived: exit status 1 panic: runtime error: index out of range [0] with length 0",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestMissingStatementFailsWithMethodName"
    ],
    "evidence": "ADAMIC_MUTANT=M2 ADAMIC_BUILD_CACHE_DIR=/tmp/u162/cache/M2 timeout 120 go test -json -count=1 -timeout 90s ./stage3/census/latent/statementrewrite/ -run . > M2.log 2>&1; main_test.go:31: missing statement mutant survived: exit status 1 panic: runtime error: index out of range [0] with length 0"
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M1 | stage3/census/latent/statementrewrite/main.go:40 | latent statement rewrite: lowering.statement: expected exactly one function, found %d -> latent statement rewrite: lowering.expression: expected exactly one function, found %d | TestMissingStatementFailsWithMethodName |
| M2 | stage3/census/latent/statementrewrite/main.go:39 | if len(matches) != 1 { -> if len(matches) == 1 { | TestMissingStatementFailsWithMethodName |
| M3 | stage3/census/latent/statementrewrite/main.go:24 | parser.ParseComments) -> 0) | none |
| P1 | stage3/census/latent/statementrewrite/main.go:19 | func main() { -> func main() {  if true { return } | TestMissingStatementFailsWithMethodName |

Survivor M3: parsing with mode 0 instead of parser.ParseComments drops // package note and // statement comment from the rewritten valid-method output. Both clean and mutant exit 0. See M3-witness.diff and the three survivor source/output files. This is observed changed behavior unguarded by this package, not an equivalent candidate. The witness commands are recorded in runs.json, with ADAMIC_MUTANT=M0 and M3 respectively.

CODE UNDER TEST: the Go statementrewrite CLI that constructs census instrumentation by locating lowering.statement, renaming it latentStatementRaw, retaining its body and adding the wrapper. ORACLE: the hand-written expected diagnostic substring, nonzero exit and absence of output file in main_test.go. No outside authority is claimed.

Scope and instruction issues

The requested package contains exactly one Test and no families, helpers, opt-in rows or witnesses. No names moved or vanished. The error assertion uses the word mutant, but the test itself checks the rewriter refusal directly; it does not witness another agreement comparator. Because the code builds census instrumentation, the suite-construction category applies and the verdict is setup-check. M1 and M2 are construction mutations in code under test, not edits to its oracle. Their unique catches do not establish repository-wide uniqueness.

The function reached by the missing-method fixture is main. printed is declared in the same package but the short-circuit method-name condition prevents the fixture calling it. The valid-method survivor witness does reach printed. The inventory lists both facts rather than treating package membership as coverage. Three mutants were fixed from main before any mutated run: one diagnostic constant, one count condition, one parser option. All matrix commands run the whole package.

The count mutant panics in a go-run subprocess; it does not abort the Go test binary. The actual test records and rejects that unrelated index panic because it lacks the required method-name diagnostic. Thus this oracle checks more than an exit code, although it checks only a substring, not exact stderr or an exact exit status. The test does not exercise successful rewriting, comment retention, signature validation or wrapper correctness. M3 demonstrates the comment-retention gap concretely.

Warm tools do not remove the required npm ci step. stage3/api dependencies were refreshed even though this package uses only Go and loads no node_modules. No tools setup script was needed, no opt-ins were required and no test skipped. No step exceeded 90 seconds, and no narrowing was needed. Compiler code and native products were not mutated, so no compiler cache isolation or clang rebuild was needed. The switched CLI is compiled once and inherited ADAMIC_MUTANT selects the child go-run behavior; each standalone diff has no switch and passes go vet for the package.

The initial pure probe used an unconditional return and go vet rejected the following unreachable code. It was replaced by if true { return }, which passes vet and has the same empty-entry behavior as the runtime probe. That correction caused a second, entirely green clean measurement round and replay of all matrices; the saved logs and timings are from the final round. Tests and their expected values were never modified. Production source is restored, and the final whole-package run passes.

Timing and uncovered work

Warm setup: 0 seconds for setup.sh; nproc=5. npm ci: 0.488 seconds. Initial baseline shell wall: 0.278 seconds. Individual binary times: 0.038, 0.040, 0.041 seconds, median 0.04 seconds. Final-round Go build/vet shell wall: 0.442 seconds. Final-round measurement, matrix and restored-test shell wall: 1.537 seconds. Raw commands and wall times are in preparation.json and runs.json. Total session work was approximately four minutes, including reading, script correction, evidence preparation and push.

Other packages were not run. The full repository gate, exact diagnostic text authority, successful-rewrite correctness beyond the comment witness, and repository-wide uniqueness remain uncovered. No test was changed or deleted, main was not pushed, and no pull request was opened.
